package grow

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Mr-xiaotian/CelestialForge/pkg/funnel"
)

type Executor[T any, R any] struct {
	Name       string
	processor  func(T) (R, error)
	numWorkers int

	TaskChan    chan Payload[T]
	ResultChan  chan Payload[R]
	ControlChan chan ControlSignal

	observers []Observer
	logSpout  *funnel.Spout[LogRecord]
	logInlet  *LogInlet
	failSpout *funnel.Spout[FailRecord[T]]
	failInlet *FailInlet[T]

	state atomic.Int32 // 0=idle, 1=running, 2=done
	Counter
}

func (e *Executor[T, R]) State() int32 {
	return e.state.Load()
}

// NewExecutor 创建执行器
func NewExecutor[T any, R any](name string, processor func(T) (R, error), numWorkers int, observers ...Observer) *Executor[T, R] {
	logSpout := funnel.NewSpout(&LogRecordHandler{}, 100, time.Second)
	logInlet := NewLogInlet(logSpout.GetQueue(), time.Second, "INFO")
	failSpout := funnel.NewSpout(&FailRecordHandler[T]{}, 100, time.Second)
	failInlet := NewFailInlet(failSpout.GetQueue(), time.Second)

	return &Executor[T, R]{
		Name:       name,
		processor:  processor,
		numWorkers: numWorkers,

		TaskChan:    make(chan Payload[T], numWorkers),
		ResultChan:  make(chan Payload[R], numWorkers),
		ControlChan: make(chan ControlSignal, numWorkers),

		observers: observers,
		logSpout:  logSpout,
		logInlet:  logInlet,
		failSpout: failSpout,
		failInlet: failInlet,
	}
}

// reportProgress 报告进度
func (e *Executor[T, R]) reportProgress() {
	completed := e.GetComplated()
	total := e.GetTotal()
	for _, observer := range e.observers {
		observer.OnProgress(completed, total)
	}
}

// notifyStart 通知开始
func (e *Executor[T, R]) notifyStart() {
	total := e.GetTotal()
	for _, observer := range e.observers {
		observer.OnStart(total)
	}
}

// notifyFinish 通知完成
func (e *Executor[T, R]) notifyFinish() {
	completed := e.GetComplated()
	total := e.GetTotal()
	for _, observer := range e.observers {
		observer.OnFinish(completed, total)
	}
}

// processTaskSuccess 处理成功任务
func (e *Executor[T, R]) processTaskSuccess(taskPayload Payload[T], result R, startTime time.Time) {
	e.AddSuccess(1)
	e.reportProgress()

	taskRepr := trunc(fmt.Sprintf("%+v", taskPayload.Value), 50)
	resultRepr := trunc(fmt.Sprintf("%+v", result), 25)
	useTime := time.Since(startTime).Seconds()
	e.logInlet.TaskSuccess(e.Name, taskRepr, resultRepr, useTime)

	e.ResultChan <- Payload[R]{ID: taskPayload.ID, Value: result, Prev: taskPayload.Value}
}

// handleTaskError 处理错误任务
func (e *Executor[T, R]) handleTaskError(taskPayload Payload[T], err error) {
	e.AddFailed(1)
	e.reportProgress()

	taskRepr := trunc(fmt.Sprintf("%+v", taskPayload.Value), 50)
	e.logInlet.TaskError(e.Name, taskRepr, err)
	e.failInlet.TaskError(e.Name, taskPayload.ID, taskPayload.Value, err)
}

// Drain 消费成功结果，直到执行器结束
func (e *Executor[T, R]) Drain(onSuccess func(Payload[R])) {
	for res := range e.ResultChan {
		if onSuccess != nil {
			onSuccess(res)
		}
	}
}

// seed 输入任务
func (e *Executor[T, R]) seed(tasks []T) {
	for idx, task := range tasks {
		e.TaskChan <- Payload[T]{ID: idx, Value: task}
	}
	e.ControlChan <- ControlSignal{Source: "executor"}
}

// worker 工作线
func (e *Executor[T, R]) worker(taskPayload Payload[T], sem chan struct{}, done chan struct{}) {
	defer func() {
		if r := recover(); r != nil {
			e.handleTaskError(taskPayload, fmt.Errorf("processor panic: %v", r))
		}
		<-sem              // 释放并发令牌
		done <- struct{}{} // 发送完成信号
	}()

	startTime := time.Now()
	result, err := e.processor(taskPayload.Value)
	if err != nil {
		e.handleTaskError(taskPayload, err)
	} else {
		e.processTaskSuccess(taskPayload, result, startTime)
	}
}

// dispatch 调度器
func (e *Executor[T, R]) dispatch() {
	sem := make(chan struct{}, e.numWorkers)  // 控制并发数
	done := make(chan struct{}, e.numWorkers) // 控制worker完成信号

	inputClosed := false
	inFlight := 0
	shouldFinish := func() bool {
		return inputClosed && inFlight == 0 && e.IsFinish()
	}

	for {
		if shouldFinish() {
			close(e.ResultChan)
			return
		}

		select {
		case task := <-e.TaskChan:
			sem <- struct{}{} // 获取并发令牌
			inFlight++
			go e.worker(task, sem, done)
		case <-done: // worker完成信号
			inFlight--
		case <-e.ControlChan:
			inputClosed = true
		}
	}
}

// collect 收集所有成功结果，并保留对应的任务信息
func (e *Executor[T, R]) collect() []TaskResult[T, R] {
	results := make([]TaskResult[T, R], 0)
	for res := range e.ResultChan {
		results = append(results, TaskResult[T, R]{
			Task:   res.Prev.(T),
			Result: res.Value,
		})
	}
	return results
}

// Start 启动执行器
func (e *Executor[T, R]) Start(tasks []T) []TaskResult[T, R] {
	e.logSpout.Start()
	e.failSpout.Start()
	e.logInlet.StartExecutor(e.Name, len(tasks))
	startTime := time.Now()

	e.state.Store(1)
	e.SetTotal(len(tasks))
	e.notifyStart()

	go e.seed(tasks)
	go e.dispatch()
	results := e.collect()

	e.notifyFinish()
	e.state.Store(2)

	e.logInlet.EndExecutor(e.Name, time.Since(startTime).Seconds(), int(e.success.Load()), int(e.failed.Load()))
	e.logSpout.Stop()
	e.failSpout.Stop()
	return results
}
