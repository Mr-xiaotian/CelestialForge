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

	SeedChan    chan Payload[S]
	FruitChan   chan Payload[F]
	ControlChan chan ControlSignal

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

// State 返回执行器当前状态：0=idle, 1=running, 2=done。
func (e *Plot[S, F]) State() int32 {
	return e.state.Load()
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

		SeedChan:    make(chan Payload[S], o.numTends),
		FruitChan:   make(chan Payload[F], o.numTends),
		ControlChan: make(chan ControlSignal, o.numTends),

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
func (e *Plot[S, F]) reportProgress() {
	completed := e.GetCompleted()
	total := e.GetTotal()
	for _, observer := range e.observers {
		observer.OnProgress(completed, total)
	}
}

// notifyStart 通知开始
func (e *Plot[S, F]) notifyStart() {
	e.state.Store(1)
	total := e.GetTotal()
	for _, observer := range e.observers {
		observer.OnStart(total)
	}
}

// notifyFinish 通知完成
func (e *Plot[S, F]) notifyFinish() {
	e.state.Store(2)
	completed := e.GetCompleted()
	total := e.GetTotal()
	for _, observer := range e.observers {
		observer.OnFinish(completed, total)
	}
}

// ==== Seed Handling ====

// processFruit 处理培育成功的种子
func (e *Plot[S, F]) processFruit(seedPayload Payload[S], fruit F, startTime time.Time) {
	e.AddSuccess(1)
	e.reportProgress()

	seedRepr := trunc(fmt.Sprintf("%+v", seedPayload.Value), 50)
	fruitRepr := trunc(fmt.Sprintf("%+v", fruit), 25)
	useTime := time.Since(startTime).Seconds()
	e.logInlet.TendSuccess(e.Name, seedRepr, fruitRepr, useTime)

	e.FruitChan <- Payload[F]{ID: seedPayload.ID, Value: fruit, Prev: seedPayload.Value}
}

// handleWeed 处理培育失败的种子
func (e *Plot[S, F]) handleWeed(seedPayload Payload[S], err error) {
	e.AddFailed(1)
	e.reportProgress()

	seedRepr := trunc(fmt.Sprintf("%+v", seedPayload.Value), 50)
	e.logInlet.TendFail(e.Name, seedRepr, err)
	e.failInlet.TendFail(e.Name, seedPayload.ID, seedPayload.Value, err)
}

// ==== Internal Pipeline ====

// seed 内部批量播种
func (e *Plot[S, F]) seed(seeds []S) {
	e.AddTotal(len(seeds))
	for idx, seed := range seeds {
		e.SeedChan <- Payload[S]{ID: idx, Value: seed}
	}
	e.ControlChan <- ControlSignal{Source: "plot"}
}

// tend 照料单个任务
func (e *Plot[S, F]) tend(seedPayload Payload[S], sem chan struct{}, done chan struct{}) {
	defer func() {
		if r := recover(); r != nil {
			e.handleWeed(seedPayload, fmt.Errorf("cultivator panic: %v", r))
		}
		<-sem              // 释放并发令牌
		done <- struct{}{} // 发送完成信号
	}()

	startTime := time.Now()
	var fruit F
	var err error

	for attempt := range e.maxRetries + 1 {
		fruit, err = e.cultivator(seedPayload.Value)
		if err == nil {
			break
		}
		if !e.retryIf(err) {
			break
		}
		time.Sleep(e.retryDelay(attempt))
	}

	if err != nil {
		e.handleWeed(seedPayload, err)
	} else {
		e.processFruit(seedPayload, fruit, startTime)
	}
}

// sprout 调度器，将种子分发给 tend 协程
func (e *Plot[S, F]) sprout() {
	sem := make(chan struct{}, e.numTends)  // 控制并发数
	done := make(chan struct{}, e.numTends) // 控制tend完成信号

	ctxCancel := false
	inputClosed := false
	inFlight := 0
	shouldFinish := func() bool {
		return ctxCancel || (inputClosed && inFlight == 0 && e.IsFinish())
	}

	for {
		if shouldFinish() {
			close(e.FruitChan)
			return
		}

		select {
		case seed := <-e.SeedChan:
			sem <- struct{}{} // 获取并发令牌
			inFlight++
			go e.tend(seed, sem, done)
		case <-done: // tend完成信号
			inFlight--
		case <-e.ControlChan:
			inputClosed = true
		case <-e.ctx.Done():
			ctxCancel = true
		}
	}
}

// harvest 收获所有果实，并保留对应的种子信息
func (e *Plot[S, F]) harvest() []Karma[S, F] {
	fruits := make([]Karma[S, F], 0)
	for res := range e.FruitChan {
		fruits = append(fruits, Karma[S, F]{
			Seed:  res.Prev.(S),
			Fruit: res.Value,
		})
	}
	return fruits
}

// ==== Sync API ====

// Start 同步启动 Plot，阻塞直到所有种子培育完成并返回果实。
func (e *Plot[S, F]) Start(seeds []S) []Karma[S, F] {
	e.logSpout.Start()
	e.failSpout.Start()
	e.logInlet.StartPlot(e.Name, e.numTends)
	startTime := time.Now()

	e.notifyStart()
	go e.seed(seeds)
	go e.sprout()
	karmas := e.harvest()
	e.notifyFinish()

	e.logInlet.EndPlot(e.Name, time.Since(startTime).Seconds(), e.GetSuccess(), e.GetFailed())
	e.logSpout.Stop()
	e.failSpout.Stop()
	return karmas
}

// ==== Async API ====

// Seed 播入单颗种子到 SeedChan。
func (e *Plot[S, F]) Seed(id int, seed S) {
	e.AddTotal(1)
	e.SeedChan <- Payload[S]{ID: id, Value: seed}
}

// Seal 封闭种子入口，通知 sprout 不再有新种子。
func (e *Plot[S, F]) Seal() {
	e.ControlChan <- ControlSignal{Source: e.Name}
}

// Harvest 逐个收获果实，阻塞直到 FruitChan 关闭。
func (e *Plot[S, F]) Harvest(onSuccess func(Payload[F])) {
	for res := range e.FruitChan {
		if onSuccess != nil {
			onSuccess(res)
		}
	}
}

// StartAsync 异步启动调度器，种子播入和果实收获由外部控制。
// 外部通过 Seed 播种，通过 Harvest 收获
// 完成后需调用 WaitAsync 进行清理
func (e *Plot[S, F]) StartAsync() {
	e.wg.Add(1)
	defer e.wg.Done()

	e.logInlet.StartPlot(e.Name, e.numTends)
	startTime := time.Now()

	e.notifyStart()
	e.sprout()
	e.notifyFinish()

	e.logInlet.EndPlot(e.Name, time.Since(startTime).Seconds(), e.GetSuccess(), e.GetFailed())
}

// WaitAsync 等待异步 Plot 结束并清理资源
func (e *Plot[S, F]) WaitAsync() {
	e.wg.Wait()
}

// ==== Cleanup API ====

// Close 立即取消 Plot，强制停止所有操作。慎用，可能导致未完成的任务丢失。
func (e *Plot[S, F]) Close() {
	e.cancel()
	e.notifyFinish()
}
