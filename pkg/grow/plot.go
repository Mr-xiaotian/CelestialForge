package grow

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Mr-xiaotian/CelestialForge/pkg/funnel"
)

// Plot 并发任务执行器。将一组种子（任务）分发给 tend 池并行培育，
// 通过 funnel 系统记录日志和失败信息。
// S 为种子类型，F 为果实类型。
type Plot[S any, F any] struct {
	Name       string
	cultivator func(S) (F, error)
	numTends   int
	maxRetries int
	retryDelay func(attempt int) time.Duration
	retryIf    func(error) bool

	SeedChan   chan Payload[S]
	FruitChans []chan Payload[F]

	observers []Observer
	logSpout  *funnel.Spout[LogRecord]
	logInlet  *LogInlet
	failSpout *funnel.Spout[FailRecord[S]]
	failInlet *FailInlet[S]

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	state  atomic.Int32 // 0=idle, 1=running, 2=done
	Counter
}

// NewPlot 创建一个 Plot 实例。
// cultivator 为培育函数，接收种子返回果实。
func NewPlot[S any, F any](name string, cultivator func(S) (F, error), observers []Observer, opts ...Option) *Plot[S, F] {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	logSpout := funnel.NewSpout(&LogRecordHandler{}, 100, time.Second)
	logInlet := NewLogInlet(logSpout.GetQueue(), time.Second, "INFO")
	failSpout := funnel.NewSpout(&FailRecordHandler[S]{}, 100, time.Second)
	failInlet := NewFailInlet(failSpout.GetQueue(), time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	return &Plot[S, F]{
		Name:       name,
		cultivator: cultivator,
		numTends:   o.numTends,
		maxRetries: o.maxRetries,
		retryDelay: o.retryDelay,
		retryIf:    o.retryIf,

		SeedChan:   make(chan Payload[S], o.numTends),
		FruitChans: []chan Payload[F]{make(chan Payload[F], o.numTends)},

		observers: observers,
		logSpout:  logSpout,
		logInlet:  logInlet,
		failSpout: failSpout,
		failInlet: failInlet,

		ctx:    ctx,
		cancel: cancel,
	}
}

// ==== Observer Hooks ====

// reportProgress 报告进度
func (p *Plot[S, F]) reportProgress() {
	completed := p.GetCompleted()
	total := p.GetTotal()
	for _, observer := range p.observers {
		observer.OnProgress(completed, total)
	}
}

// notifyStart 通知开始
func (p *Plot[S, F]) notifyStart() {
	p.state.Store(1)
	total := p.GetTotal()
	for _, observer := range p.observers {
		observer.OnStart(total)
	}
}

// notifyFinish 通知完成
func (p *Plot[S, F]) notifyFinish() {
	p.state.Store(2)
	completed := p.GetCompleted()
	total := p.GetTotal()
	for _, observer := range p.observers {
		observer.OnFinish(completed, total)
	}
}

// ==== Seed Handling ====

// bearFruit 处理培育成功的种子
func (p *Plot[S, F]) bearFruit(seedPayload Payload[S], fruit F, startTime time.Time) {
	p.AddSuccess(1)
	p.reportProgress()

	seedRepr := trunc(fmt.Sprintf("%+v", seedPayload.Value), 50)
	fruitRepr := trunc(fmt.Sprintf("%+v", fruit), 25)
	useTime := time.Since(startTime).Seconds()
	p.logInlet.TendSuccess(p.Name, seedRepr, fruitRepr, useTime)

	fruitPayload := Payload[F]{Value: fruit, Prev: seedPayload.Value}
	for _, ch := range p.FruitChans {
		ch <- fruitPayload
	}
}

// bearWeed 处理培育失败的种子
func (p *Plot[S, F]) bearWeed(seedPayload Payload[S], err error) {
	p.AddFailed(1)
	p.reportProgress()

	seedRepr := trunc(fmt.Sprintf("%+v", seedPayload.Value), 50)
	p.logInlet.TendFail(p.Name, seedRepr, err)
	p.failInlet.TendFail(p.Name, seedPayload.Value, err)
}

// State 返回执行器当前状态：0=idle, 1=running, 2=done。
func (p *Plot[S, F]) State() int32 {
	return p.state.Load()
}

// ==== Internal Pipeline ====

// seed 内部批量播种
func (p *Plot[S, F]) seed(seeds []S) {
	p.AddTotal(len(seeds))
	for idx, seed := range seeds {
		p.SeedChan <- Payload[S]{ID: idx, Value: seed}
	}
	p.SeedChan <- Payload[S]{Signal: SignalSeal, Source: p.Name}
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
		p.logInlet.TendRetry(p.Name, seedRepr, attempt, err)
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
			sealPayload := Payload[F]{Signal: SignalSeal, Source: p.Name}
			for _, ch := range p.FruitChans {
				ch <- sealPayload
			}
			return
		}

		select {
		case seed := <-p.SeedChan:
			if seed.Signal == SignalSeal {
				inputClosed = true
				continue
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
	for res := range p.FruitChans[0] {
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
	p.logSpout.Start()
	p.failSpout.Start()
	p.logInlet.StartPlot(p.Name, p.numTends)
	startTime := time.Now()

	p.notifyStart()
	go p.seed(seeds)
	go p.sprout()
	karmas := p.harvest()
	p.notifyFinish()

	p.logInlet.EndPlot(p.Name, time.Since(startTime).Seconds(), p.GetSuccess(), p.GetFailed())
	p.logSpout.Stop()
	p.failSpout.Stop()
	return karmas
}

// ==== Async API ====

// Seed 播入单颗种子到 SeedChan。
func (p *Plot[S, F]) Seed(id int, seed S) {
	p.AddTotal(1)
	p.SeedChan <- Payload[S]{ID: id, Value: seed}
}

// Seal 封闭种子入口，通知 sprout 不再有新种子。
func (p *Plot[S, F]) Seal() {
	p.SeedChan <- Payload[S]{Signal: SignalSeal, Source: p.Name}
}

// Harvest 逐个收获果实，阻塞直到 FruitChans 关闭。
func (p *Plot[S, F]) Harvest(sickle func(Payload[F])) {
	for res := range p.FruitChans[0] {
		if res.Signal == SignalSeal {
			break
		}
		if sickle != nil {
			sickle(res)
		}
	}
}

// StartAsync 异步启动调度器，种子播入和果实收获由外部控制。
// 外部通过 Seed 播种，通过 Harvest 收获
// 完成后需调用 WaitAsync 进行清理
func (p *Plot[S, F]) StartAsync() {
	p.wg.Add(1)
	defer p.wg.Done()

	p.logInlet.StartPlot(p.Name, p.numTends)
	startTime := time.Now()

	p.notifyStart()
	p.sprout()
	p.notifyFinish()

	p.logInlet.EndPlot(p.Name, time.Since(startTime).Seconds(), p.GetSuccess(), p.GetFailed())
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
