package grow

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Mr-xiaotian/CelestialForge/pkg/funnel"
)

// ==== Interface ====

// PlotNode 是 Farm 管理 plot 时使用的统一接口。
// 它擦除了泛型参数，使 Farm 可以用同一类型持有不同种子/果实类型的 Plot。
// 同时覆盖建图阶段（ConnectTo、AddUpstream、BindInlet）
// 和运行阶段（StartAsync、WaitAsync、SeedAny、Seal）所需的最小能力。
type PlotNode interface {
	GetName() string
	GetState() int32
	GetSeedChanAny() any

	ConnectTo(next PlotNode) error
	AddUpstream(name string)
	BindInlet(logChan chan<- LogRecord, failChan chan<- FailRecord)

	StartAsync()
	WaitAsync()
	SeedAny(seed any) error
	Seal()
}

// ==== Struct ====

// Plot 并发种子培育器。将一组种子分发给 tend 池并行培育，
// 通过 funnel 系统异步记录日志和失败信息。
// S 为种子类型，F 为果实类型。
type Plot[S any, F any] struct {
	name       string
	cultivator func(S) (F, error)
	observers  []Observer
	plotOptions

	seedChan   chan Payload[S]
	fruitChans map[string]chan Payload[F]
	upstreams  map[string]struct{}

	logSpout  *funnel.Spout[LogRecord]
	failSpout *funnel.Spout[FailRecord]
	logInlet  *LogInlet
	failInlet *FailInlet

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	state  atomic.Int32 // 0=idle, 1=running, 2=done
	Counter
}

// ==== Constructor ====

// NewPlot 创建一个 Plot 实例。
// name 为 plot 名称（在 Farm 中需唯一），cultivator 为培育函数，
// opts 为可选配置项。
func NewPlot[S any, F any](name string, cultivator func(S) (F, error), opts ...Option) *Plot[S, F] {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Plot[S, F]{
		name:        name,
		cultivator:  cultivator,
		plotOptions: o,

		seedChan:   make(chan Payload[S], o.chanSize),
		fruitChans: make(map[string]chan Payload[F]),
		upstreams:  make(map[string]struct{}),

		ctx:    ctx,
		cancel: cancel,
	}
}

// AddObserver 添加一个进度观察者。
func (p *Plot[S, F]) AddObserver(observer Observer) {
	p.observers = append(p.observers, observer)
}

// ==== Initialization ====

// InitLocalEnv 初始化 standalone 模式所需的本地环境。
// 创建本地 fruitChan、日志/失败 spout 并绑定 inlet。
// Farm 模式下不需要调用此方法，由 Farm 统一管理 spout 和 inlet。
func (p *Plot[S, F]) InitLocalEnv() {
	p.logSpout = funnel.NewSpout(&LogRecordHandler{}, 100, time.Second)
	p.failSpout = funnel.NewSpout(&FailRecordHandler{}, 100, time.Second)

	p.BindInlet(p.logSpout.GetQueue(), p.failSpout.GetQueue())
}

// BindInlet 绑定日志和失败记录的写入通道。
// standalone 模式由 InitLocalEnv 内部调用；Farm 模式由 Farm.Start 统一调用。
func (p *Plot[S, F]) BindInlet(logChan chan<- LogRecord, failChan chan<- FailRecord) {
	p.logInlet = NewLogInlet(logChan, time.Second, p.logLevel)
	p.failInlet = NewFailInlet(failChan, time.Second)
}

// StartSpouts 启动本地日志/失败 spout。仅 standalone 模式使用。
func (p *Plot[S, F]) StartSpouts() {
	p.logSpout.Start()
	p.failSpout.Start()
}

// StopSpouts 停止本地日志/失败 spout 并刷盘。仅 standalone 模式使用。
func (p *Plot[S, F]) StopSpouts() {
	p.logSpout.Stop()
	p.failSpout.Stop()
}

// ==== Connection ====

// AddUpstream 登记一个上游 plot 名称。
// 当未收到外部 input seal 时，sprout 需要等所有已登记上游都发送过
// seal 信号后，才会将输入视为关闭。
func (p *Plot[S, F]) AddUpstream(name string) {
	if name == "" {
		return
	}
	p.upstreams[name] = struct{}{}
}

// ConnectTo 将当前 plot 的果实输出连接到下游 plot 的种子输入。
// 通过类型断言校验上游 F 与下游 S 是否匹配。
func (p *Plot[S, F]) ConnectTo(next PlotNode) error {
	seedChan, ok := next.GetSeedChanAny().(chan Payload[F])
	if !ok {
		return fmt.Errorf("plot %q fruit type is incompatible with plot %q seed type", p.name, next.GetName())
	}

	p.fruitChans[next.GetName()] = seedChan
	return nil
}

// markSealed 处理一条 seal 信号并判断输入是否关闭。
// sourceInput 表示外部调用者显式终止该 plot 的输入，此时直接关闭；
// 否则需要等待所有已登记上游都发送过 seal 信号。
func (p *Plot[S, F]) markSealed(source string, sealedFrom map[string]struct{}) bool {
	if source == sourceInput {
		return true
	}
	if source == "" {
		return false
	}
	if _, ok := p.upstreams[source]; !ok {
		return false
	}
	sealedFrom[source] = struct{}{}
	return len(sealedFrom) == len(p.upstreams)
}

// ==== Getters ====

// GetName 返回 plot 名称。
func (p *Plot[S, F]) GetName() string {
	return p.name
}

// GetState 返回当前状态：0=idle, 1=running, 2=done。
func (p *Plot[S, F]) GetState() int32 {
	return p.state.Load()
}

// GetSeedChanAny 以 any 类型返回 seedChan，供 Farm 连接时做类型断言。
func (p *Plot[S, F]) GetSeedChanAny() any {
	return p.seedChan
}

// ==== Observer Hooks ====

// reportProgress 通知所有 Observer 当前进度。
func (p *Plot[S, F]) reportProgress() {
	completed := p.GetCompleted()
	seedNum := p.GetSeedNum()
	for _, observer := range p.observers {
		observer.OnProgress(completed, seedNum)
	}
}

// notifyStart 将状态设为 running 并通知所有 Observer。
func (p *Plot[S, F]) notifyStart() {
	p.state.Store(1)
	seedNum := p.GetSeedNum()
	for _, observer := range p.observers {
		observer.OnStart(seedNum)
	}
}

// notifyFinish 将状态设为 done 并通知所有 Observer。
func (p *Plot[S, F]) notifyFinish() {
	p.state.Store(2)
	completed := p.GetCompleted()
	seedNum := p.GetSeedNum()
	for _, observer := range p.observers {
		observer.OnFinish(completed, seedNum)
	}
}

// ==== Result Handling ====

// bearFruit 处理培育成功的种子：更新计数、记录日志、将果实发送到所有下游通道。
func (p *Plot[S, F]) bearFruit(seedPayload Payload[S], fruit F, startTime time.Time) {
	p.AddFruitNum(1)
	p.reportProgress()

	seedRepr := trunc(fmt.Sprintf("%+v", seedPayload.Value), 50)
	fruitRepr := trunc(fmt.Sprintf("%+v", fruit), 25)
	useTime := time.Since(startTime).Seconds()
	p.logInlet.SeedRipen(p.name, seedRepr, fruitRepr, useTime)

	fruitPayload := Payload[F]{Value: fruit}
	for _, ch := range p.fruitChans {
		ch <- fruitPayload
	}
}

// bearWeed 处理培育失败的种子：更新计数、记录日志和失败记录。
func (p *Plot[S, F]) bearWeed(seedPayload Payload[S], err error, startTime time.Time) {
	p.AddWeedNum(1)
	p.reportProgress()

	seedString := fmt.Sprintf("%+v", seedPayload.Value)
	seedRepr := trunc(seedString, 50)
	useTime := time.Since(startTime).Seconds()
	p.logInlet.SeedWither(p.name, seedRepr, err, useTime)
	p.failInlet.SeedWither(p.name, seedString, err)
}

// ==== Internal Pipeline ====

// tend 照料单颗种子：执行 cultivator 并在失败时按策略重试。
// 完成后通过 bearFruit 或 bearWeed 路由结果。
func (p *Plot[S, F]) tend(seedPayload Payload[S], sem chan struct{}, done chan struct{}) {
	defer func() {
		if r := recover(); r != nil {
			p.bearWeed(seedPayload, fmt.Errorf("cultivator panic: %v", r), time.Now())
		}
		<-sem              // 释放并发令牌
		done <- struct{}{} // 发送完成信号
	}()

	startTime := time.Now()
	seedRepr := trunc(fmt.Sprintf("%+v", seedPayload.Value), 50)

	var fruit F
	var err error

	for attempt := 1; attempt <= p.maxRetries+1; attempt++ {
		fruit, err = p.cultivator(seedPayload.Value)
		if err == nil {
			break
		}
		if !p.retryIf(err) {
			break
		}
		if attempt <= p.maxRetries {
			p.logInlet.SeedReplant(p.name, seedRepr, attempt, err)
		}
		time.Sleep(p.retryDelay(attempt))
	}

	if err != nil {
		p.bearWeed(seedPayload, err, startTime)
	} else {
		p.bearFruit(seedPayload, fruit, startTime)
	}
}

// sprout 调度器：从 seedChan 读取种子，分发给 tend 协程并行处理。
// 通过信号量控制最大并发数，并根据 seal 信号判断输入是否结束。
// 外部 input seal 具有强终止语义：一旦收到，即不再等待剩余上游 seal。
// 所有已接收种子处理完毕后，向所有 fruitChans 发送 SignalSeal 通知下游。
func (p *Plot[S, F]) sprout() {
	sem := make(chan struct{}, p.numTends)
	done := make(chan struct{}, p.numTends)
	sealedFrom := make(map[string]struct{}, len(p.upstreams))

	ctxCancel := false
	inputClosed := false
	inFlight := 0
	shouldFinish := func() bool {
		return ctxCancel || (inputClosed && inFlight == 0)
	}

	for {
		if shouldFinish() {
			sealPayload := Payload[F]{Signal: SignalSeal, Source: p.name}
			for _, ch := range p.fruitChans {
				ch <- sealPayload
			}
			return
		}

		select {
		case seed := <-p.seedChan:
			if seed.Signal == SignalSeal {
				inputClosed = p.markSealed(seed.Source, sealedFrom)
				continue
			}
			p.AddSeedNum(1)

			sem <- struct{}{}
			inFlight++
			go p.tend(seed, sem, done)
		case <-done:
			inFlight--
		case <-p.ctx.Done():
			ctxCancel = true
		}
	}
}

// ==== Async API ====

// SeedAny 以 any 类型播入单颗种子，内部做类型断言。
// 供 Farm 统一注入初始任务时使用。
func (p *Plot[S, F]) SeedAny(seed any) error {
	typedSeed, ok := seed.(S)
	if !ok {
		return fmt.Errorf("plot %q seed type mismatch: got %T", p.name, seed)
	}
	p.Seed(typedSeed)
	return nil
}

// Seed 播入单颗种子到 seedChan。
// 该输入视为外部调用者注入，而非来自某个上游 plot。
func (p *Plot[S, F]) Seed(seed S) {
	p.seedChan <- Payload[S]{Value: seed}
}

// Seal 向 seedChan 发送来自外部 input 的 SignalSeal。
// 这表示外部调用者要求终止该 plot 的后续输入处理；对于已连接上游的
// plot，这同样会触发强终止，而不是继续等待剩余上游 seal。
func (p *Plot[S, F]) Seal() {
	p.seedChan <- Payload[S]{Signal: SignalSeal, Source: sourceInput}
}

// StartAsync 异步启动 sprout 调度器。
// 调用前需先完成 BindInlet 绑定通道。
// 外部通过 Seed 播种、Seal 终止；完成后需调用 WaitAsync 等待退出。
func (p *Plot[S, F]) StartAsync() {
	p.wg.Go(func() {
		p.logInlet.StartPlot(p.name, p.numTends)
		startTime := time.Now()

		p.notifyStart()
		p.sprout()
		p.notifyFinish()

		p.logInlet.EndPlot(p.name, time.Since(startTime).Seconds(), p.GetFruitNum(), p.GetWeedNum())
	})
}

// WaitAsync 等待异步 Plot 的所有协程（sprout、Harvest、Seed、Seal）退出。
func (p *Plot[S, F]) WaitAsync() {
	p.wg.Wait()
}
