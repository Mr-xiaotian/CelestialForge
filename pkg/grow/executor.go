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
// T 为种子类型，R 为果实类型。
type Plot[T any, R any] struct {
	Name       string
	cultivator func(T) (R, error)
	numWorkers int

	SeedChan    chan Payload[T]
	FruitChan   chan Payload[R]
	ControlChan chan ControlSignal

	observers []Observer
	logSpout  *funnel.Spout[LogRecord]
	logInlet  *LogInlet
	failSpout *funnel.Spout[FailRecord[T]]
	failInlet *FailInlet[T]

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
	state  atomic.Int32 // 0=idle, 1=running, 2=done
	Counter
}

// State 返回执行器当前状态：0=idle, 1=running, 2=done。
func (e *Plot[T, R]) State() int32 {
	return e.state.Load()
}

// NewPlot 创建一个 Plot 实例。
// cultivator 为培育函数，接收种子返回果实。
func NewPlot[T any, R any](name string, cultivator func(T) (R, error), observers []Observer, opts ...Option) *Plot[T, R] {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	logSpout := funnel.NewSpout(&LogRecordHandler{}, 100, time.Second)
	logInlet := NewLogInlet(logSpout.GetQueue(), time.Second, "INFO")
	failSpout := funnel.NewSpout(&FailRecordHandler[T]{}, 100, time.Second)
	failInlet := NewFailInlet(failSpout.GetQueue(), time.Second)
	ctx, cancel := context.WithCancel(context.Background())

	return &Plot[T, R]{
		Name:       name,
		cultivator: cultivator,
		numWorkers: o.numWorkers,

		SeedChan:    make(chan Payload[T], o.numWorkers),
		FruitChan:   make(chan Payload[R], o.numWorkers),
		ControlChan: make(chan ControlSignal, o.numWorkers),

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
func (e *Plot[T, R]) reportProgress() {
	completed := e.GetCompleted()
	total := e.GetTotal()
	for _, observer := range e.observers {
		observer.OnProgress(completed, total)
	}
}

// notifyStart 通知开始
func (e *Plot[T, R]) notifyStart() {
	e.state.Store(1)
	total := e.GetTotal()
	for _, observer := range e.observers {
		observer.OnStart(total)
	}
}

// notifyFinish 通知完成
func (e *Plot[T, R]) notifyFinish() {
	e.state.Store(2)
	completed := e.GetCompleted()
	total := e.GetTotal()
	for _, observer := range e.observers {
		observer.OnFinish(completed, total)
	}
}

// ==== Task Handling ====

// processTaskSuccess 处理培育成功的种子
func (e *Plot[T, R]) processTaskSuccess(taskPayload Payload[T], result R, startTime time.Time) {
	e.AddSuccess(1)
	e.reportProgress()

	taskRepr := trunc(fmt.Sprintf("%+v", taskPayload.Value), 50)
	resultRepr := trunc(fmt.Sprintf("%+v", result), 25)
	useTime := time.Since(startTime).Seconds()
	e.logInlet.TaskSuccess(e.Name, taskRepr, resultRepr, useTime)

	e.FruitChan <- Payload[R]{ID: taskPayload.ID, Value: result, Prev: taskPayload.Value}
}

// handleTaskError 处理培育失败的种子
func (e *Plot[T, R]) handleTaskError(taskPayload Payload[T], err error) {
	e.AddFailed(1)
	e.reportProgress()

	taskRepr := trunc(fmt.Sprintf("%+v", taskPayload.Value), 50)
	e.logInlet.TaskError(e.Name, taskRepr, err)
	e.failInlet.TaskError(e.Name, taskPayload.ID, taskPayload.Value, err)
}

// ==== Internal Pipeline ====

// seed 内部批量播种
func (e *Plot[T, R]) seed(tasks []T) {
	e.AddTotal(len(tasks))
	for idx, task := range tasks {
		e.SeedChan <- Payload[T]{ID: idx, Value: task}
	}
	e.ControlChan <- ControlSignal{Source: "plot"}
}

// tend 照料单个任务
func (e *Plot[T, R]) tend(taskPayload Payload[T], sem chan struct{}, done chan struct{}) {
	defer func() {
		if r := recover(); r != nil {
			e.handleTaskError(taskPayload, fmt.Errorf("cultivator panic: %v", r))
		}
		<-sem              // 释放并发令牌
		done <- struct{}{} // 发送完成信号
	}()

	startTime := time.Now()
	result, err := e.cultivator(taskPayload.Value)
	if err != nil {
		e.handleTaskError(taskPayload, err)
	} else {
		e.processTaskSuccess(taskPayload, result, startTime)
	}
}

// sprout 调度器，将种子分发给 tend 协程
func (e *Plot[T, R]) sprout() {
	sem := make(chan struct{}, e.numWorkers)  // 控制并发数
	done := make(chan struct{}, e.numWorkers) // 控制tend完成信号

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
		case task := <-e.SeedChan:
			sem <- struct{}{} // 获取并发令牌
			inFlight++
			go e.tend(task, sem, done)
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
func (e *Plot[T, R]) harvest() []Karma[T, R] {
	results := make([]Karma[T, R], 0)
	for res := range e.FruitChan {
		results = append(results, Karma[T, R]{
			Seed:  res.Prev.(T),
			Fruit: res.Value,
		})
	}
	return results
}

// ==== Sync API ====

// Start 同步启动 Plot，阻塞直到所有种子培育完成并返回果实。
func (e *Plot[T, R]) Start(tasks []T) []Karma[T, R] {
	e.logSpout.Start()
	e.failSpout.Start()
	e.logInlet.StartPlot(e.Name, e.numWorkers)
	startTime := time.Now()

	e.notifyStart()
	go e.seed(tasks)
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
func (e *Plot[T, R]) Seed(id int, task T) {
	e.AddTotal(1)
	e.SeedChan <- Payload[T]{ID: id, Value: task}
}

// Harvest 逐个收获果实，阻塞直到 FruitChan 关闭。
func (e *Plot[T, R]) Harvest(onSuccess func(Payload[R])) {
	for res := range e.FruitChan {
		if onSuccess != nil {
			onSuccess(res)
		}
	}
}

// StartAsync 异步启动调度器，种子播入和果实收获由外部控制。
// 外部通过 Seed 播种，通过 Harvest 收获
// 完成后需调用 WaitAsync 进行清理
func (e *Plot[T, R]) StartAsync() {
	e.wg.Add(1)
	defer e.wg.Done()

	e.logInlet.StartPlot(e.Name, e.numWorkers)
	startTime := time.Now()

	e.notifyStart()
	e.sprout()
	e.notifyFinish()

	e.logInlet.EndPlot(e.Name, time.Since(startTime).Seconds(), e.GetSuccess(), e.GetFailed())
}

// Seal 封闭种子入口，通知 sprout 不再有新种子。
func (e *Plot[T, R]) Seal() {
	e.ControlChan <- ControlSignal{Source: e.Name}
}

// WaitAsync 等待异步 Plot 结束并清理资源
func (e *Plot[T, R]) WaitAsync() {
	e.wg.Wait()
}

// ==== Cleanup API ====

// Close 立即取消 Plot，强制停止所有操作。慎用，可能导致未完成的任务丢失。
func (e *Plot[T, R]) Close() {
	e.cancel()
	e.notifyFinish()
}
