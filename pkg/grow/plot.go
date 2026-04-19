package grow

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Mr-xiaotian/CelestialForge/pkg/funnel"
)

// PlotNode 是 Farm 管理 plot 时使用的统一接口。
// 它同时覆盖建图与运行阶段所需的最小能力。
type PlotNode interface {
	GetName() string
	GetState() int32
	GetSeedChanAny() any

	ConnectTo(next PlotNode) error
	BindInlet(logChan chan<- LogRecord, failChan chan<- FailRecord)

	StartAsync()
	WaitAsync()
	SeedAny(id int, seed any) error
	Seal()
}

// Plot 并发任务执行器。将一组种子（任务）分发给 tend 池并行培育，
// 通过 funnel 系统记录日志和失败信息。
// S 为种子类型，F 为果实类型。
type Plot[S any, F any] struct {
	name       string
	cultivator func(S) (F, error)
	observers  []Observer
	numTends   int
	maxRetries int
	retryDelay func(attempt int) time.Duration
	retryIf    func(error) bool

	seedChan   chan Payload[S]
	fruitChans []chan Payload[F]

	logSpout  *funnel.Spout[LogRecord]
	logInlet  *LogInlet
	failSpout *funnel.Spout[FailRecord]
	failInlet *FailInlet

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	state  atomic.Int32 // 0=idle, 1=running, 2=done
	Counter
}

// ==== PlotNode 实现 ====

// NewPlot 创建一个 Plot 实例。
// cultivator 为培育函数，接收种子返回果实。
func NewPlot[S any, F any](name string, cultivator func(S) (F, error), observers []Observer, opts ...Option) *Plot[S, F] {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Plot[S, F]{
		name:       name,
		cultivator: cultivator,
		observers:  observers,
		numTends:   o.numTends,
		maxRetries: o.maxRetries,
		retryDelay: o.retryDelay,
		retryIf:    o.retryIf,

		seedChan:   make(chan Payload[S], o.numTends),
		fruitChans: []chan Payload[F]{},

		ctx:    ctx,
		cancel: cancel,
	}
}

// InitLocalEnv 初始化本地环境，包括创建日志/失败 spout 并绑定 inlet。
func (p *Plot[S, F]) InitLocalEnv() {
	fruitChan := make(chan Payload[F], p.numTends)

	p.logSpout = funnel.NewSpout(&LogRecordHandler{}, 100, time.Second)
	p.failSpout = funnel.NewSpout(&FailRecordHandler{}, 100, time.Second)

	p.addFruitChan(fruitChan)
	p.BindInlet(p.logSpout.GetQueue(), p.failSpout.GetQueue())
}

// BindInlet 绑定 plot 运行时所需的四类通道。
func (p *Plot[S, F]) BindInlet(logChan chan<- LogRecord, failChan chan<- FailRecord) {
	p.logInlet = NewLogInlet(logChan, time.Second, "INFO")
	p.failInlet = NewFailInlet(failChan, time.Second)
}

func (p *Plot[S, F]) addFruitChan(fruitChan chan Payload[F]) {
	p.fruitChans = append(p.fruitChans, fruitChan)
}

// ==== Observer Hooks ====

// reportProgress 报告进度
func (p *Plot[S, F]) reportProgress() {
	completed := p.GetCompleted()
	seedNum := p.GetSeedNum()
	for _, observer := range p.observers {
		observer.OnProgress(completed, seedNum)
	}
}

// notifyStart 通知开始
func (p *Plot[S, F]) notifyStart() {
	p.state.Store(1)
	seedNum := p.GetSeedNum()
	for _, observer := range p.observers {
		observer.OnStart(seedNum)
	}
}

// notifyFinish 通知完成
func (p *Plot[S, F]) notifyFinish() {
	p.state.Store(2)
	completed := p.GetCompleted()
	seedNum := p.GetSeedNum()
	for _, observer := range p.observers {
		observer.OnFinish(completed, seedNum)
	}
}

// ==== Seed Handling ====

// bearFruit 处理培育成功的种子
func (p *Plot[S, F]) bearFruit(seedPayload Payload[S], fruit F, startTime time.Time) {
	p.AddFruitNum(1)
	p.reportProgress()

	seedRepr := trunc(fmt.Sprintf("%+v", seedPayload.Value), 50)
	fruitRepr := trunc(fmt.Sprintf("%+v", fruit), 25)
	useTime := time.Since(startTime).Seconds()
	p.logInlet.SeedRipen(p.name, seedRepr, fruitRepr, useTime)

	fruitPayload := Payload[F]{Value: fruit, Prev: seedPayload.Value, Source: p.name}
	for _, ch := range p.fruitChans {
		ch <- fruitPayload
	}
}

// bearWeed 处理培育失败的种子
func (p *Plot[S, F]) bearWeed(seedPayload Payload[S], err error) {
	p.AddWeedNum(1)
	p.reportProgress()

	seedRepr := trunc(fmt.Sprintf("%+v", seedPayload.Value), 50)
	p.logInlet.SeedWither(p.name, seedRepr, err)
	p.failInlet.SeedWither(p.name, seedPayload.Value, err)
}

// ==== Getters ====

// GetName 返回 plot 名称。
func (p *Plot[S, F]) GetName() string {
	return p.name
}

// GetState 返回执行器当前状态：0=idle, 1=running, 2=done。
func (p *Plot[S, F]) GetState() int32 {
	return p.state.Load()
}

// GetSeedChanAny 返回 SeedChan 的动态类型表示，供 Farm 在连接时做类型对齐。
func (p *Plot[S, F]) GetSeedChanAny() any {
	return p.seedChan
}

// ==== Connection ====

// ConnectTo 将当前 plot 的 fruit 输出连接到下游 plot 的 seed 输入。
func (p *Plot[S, F]) ConnectTo(next PlotNode) error {
	seedChan, ok := next.GetSeedChanAny().(chan Payload[F])
	if !ok {
		return fmt.Errorf("plot %q fruit type is incompatible with plot %q seed type", p.name, next.GetName())
	}

	p.addFruitChan(seedChan)
	return nil
}

// ==== Spout ====

// StartSpouts 启动本地日志/失败 spout。
// 用于 standalone 模式运行。
func (p *Plot[S, F]) StartSpouts() {
	p.logSpout.Start()
	p.failSpout.Start()
}

// StopSpouts 停止本地日志/失败 spout。
func (p *Plot[S, F]) StopSpouts() {
	p.logSpout.Stop()
	p.failSpout.Stop()
}

// ==== Internal Pipeline ====

// seed 内部批量播种
func (p *Plot[S, F]) seed(seeds []S) {
	p.AddSeedNum(len(seeds))
	for idx, seed := range seeds {
		p.seedChan <- Payload[S]{ID: idx, Value: seed, Source: p.name}
	}
	p.seedChan <- Payload[S]{Signal: SignalSeal, Source: p.name}
}

// tend 照料单个任务
func (p *Plot[S, F]) tend(seedPayload Payload[S], sem chan struct{}, done chan struct{}) {
	defer func() {
		if r := recover(); r != nil {
			p.bearWeed(seedPayload, fmt.Errorf("cultivator panic: %v", r))
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
		p.logInlet.SeedReplant(p.name, seedRepr, attempt, err)
		time.Sleep(p.retryDelay(attempt))
	}

	if err != nil {
		p.bearWeed(seedPayload, err)
	} else {
		p.bearFruit(seedPayload, fruit, startTime)
	}
}

// sprout 调度器，将种子分发给 tend 协程
func (p *Plot[S, F]) sprout() {
	sem := make(chan struct{}, p.numTends)  // 控制并发数
	done := make(chan struct{}, p.numTends) // 控制tend完成信号

	ctxCancel := false
	inputClosed := false
	inFlight := 0
	shouldFinish := func() bool {
		return ctxCancel || (inputClosed && inFlight == 0 && p.IsFinish())
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
				inputClosed = true
				continue
			}
			if seed.Source != p.name {
				p.AddSeedNum(1)
			}
			sem <- struct{}{} // 获取并发令牌
			inFlight++
			go p.tend(seed, sem, done)
		case <-done: // tend完成信号
			inFlight--
		case <-p.ctx.Done():
			ctxCancel = true
		}
	}
}

// harvest 收获所有果实，并保留对应的种子信息
func (p *Plot[S, F]) harvest() []Karma[S, F] {
	fruits := make([]Karma[S, F], 0)
	for res := range p.fruitChans[0] {
		if res.Signal == SignalSeal {
			break
		}
		fruits = append(fruits, Karma[S, F]{
			Seed:  res.Prev.(S),
			Fruit: res.Value,
		})
	}
	return fruits
}

// ==== Sync API ====

// Start 同步启动 Plot，阻塞直到所有种子培育完成并返回果实。
func (p *Plot[S, F]) Start(seeds []S) []Karma[S, F] {
	p.InitLocalEnv()

	p.StartSpouts()
	p.logInlet.StartPlot(p.name, p.numTends)
	startTime := time.Now()

	p.notifyStart()
	go p.seed(seeds)
	go p.sprout()
	karmas := p.harvest()
	p.notifyFinish()

	p.logInlet.EndPlot(p.name, time.Since(startTime).Seconds(), p.GetFruitNum(), p.GetWeedNum())
	p.StopSpouts()
	return karmas
}

// ==== Async API ====

// SeedAny 以弱类型方式播入单颗种子，供 Farm 统一注入初始任务时使用。
func (p *Plot[S, F]) SeedAny(id int, seed any) error {
	typedSeed, ok := seed.(S)
	if !ok {
		return fmt.Errorf("plot %q seed type mismatch: got %T", p.name, seed)
	}
	p.Seed(id, typedSeed)
	return nil
}

// Seed 播入单颗种子到 SeedChan。
func (p *Plot[S, F]) Seed(id int, seed S) {
	p.wg.Add(1)
	defer p.wg.Done()

	p.AddSeedNum(1)
	p.seedChan <- Payload[S]{ID: id, Value: seed, Source: p.name}
}

// Seal 通过发送 SignalSeal 显式封闭种子入口。
// 异步模式约定使用信号终止，而不是关闭 SeedChan。
func (p *Plot[S, F]) Seal() {
	p.wg.Add(1)
	defer p.wg.Done()

	p.seedChan <- Payload[S]{Signal: SignalSeal, Source: p.name}
}

// Harvest 逐个收获果实，阻塞直到收到 SignalSeal。
func (p *Plot[S, F]) Harvest(sickle func(Payload[F]), chanIndex int) {
	p.wg.Add(1)
	defer p.wg.Done()

	for res := range p.fruitChans[chanIndex] {
		if res.Signal == SignalSeal {
			break
		}
		if sickle != nil {
			sickle(res)
		}
	}
}

// StartAsync 异步启动调度器，种子播入和果实收获由外部控制。
// 调用前需先完成 BindInlet，确保 Seed/Fruit/Log/Fail 通道已注入。
// 外部通过 Seed 播种，通过 Harvest 收获，并通过 Seal 显式发送终止信号。
// 完成后需调用 WaitAsync 等待调度器及外部交互协程收尾。
func (p *Plot[S, F]) StartAsync() {
	p.wg.Add(1)
	defer p.wg.Done()

	p.logInlet.StartPlot(p.name, p.numTends)
	startTime := time.Now()

	p.notifyStart()
	p.sprout()
	p.notifyFinish()

	p.logInlet.EndPlot(p.name, time.Since(startTime).Seconds(), p.GetFruitNum(), p.GetWeedNum())
}

// WaitAsync 等待异步 Plot 结束并清理资源
func (p *Plot[S, F]) WaitAsync() {
	p.wg.Wait()
}

// ==== Cleanup API ====

// Close 立即取消 Plot，强制停止所有操作。慎用，可能导致未完成的任务丢失。
// func (p *Plot[S, F]) Close() {
// 	p.cancel()
// }
