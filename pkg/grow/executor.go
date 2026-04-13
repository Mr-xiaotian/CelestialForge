package grow

import (
	"fmt"
	"time"

	"github.com/Mr-xiaotian/CelestialForge/pkg/pipline"
)

type Executor[T any, R any] struct {
	Name       string
	processor  func(T) (R, error)
	numWorkers int

	TaskChan    chan Payload[T]
	SuccChan    chan Payload[R]
	ErrChan     chan ExecuteError
	ControlChan chan ControlSignal

	observers []Observer
	logSink   *pipline.Sink[LogRecord]
	logSource *LogSource
	Counter
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
	e.logSource.TaskSuccess(e.Name, taskRepr, resultRepr, useTime)

	e.SuccChan <- Payload[R]{ID: taskPayload.ID, Value: result}
}

// handleTaskError 处理错误任务
func (e *Executor[T, R]) handleTaskError(taskPayload Payload[T], err error) {
	e.AddFailed(1)
	e.reportProgress()

	taskRepr := trunc(fmt.Sprintf("%+v", taskPayload.Value), 50)
	e.logSource.TaskError(e.Name, taskRepr, err)

	e.ErrChan <- ExecuteError{ID: taskPayload.ID, Error: err}
}

// Drain 消费指定数量的成功/错误结果，确保结果通道被完整读取
func (e *Executor[T, R]) Drain(expected int, onSuccess func(Payload[R]), onError func(ExecuteError)) {
	succChan := e.SuccChan
	errChan := e.ErrChan
	received := 0

	for received < expected {
		select {
		case res, _ := <-succChan:
			received++
			if onSuccess != nil {
				onSuccess(res)
			}
		case execErr, _ := <-errChan:
			received++
			if onError != nil {
				onError(execErr)
			}
		}
	}
}

// inputTask 输入任务
func (e *Executor[T, R]) inputTask(tasks []T) {
	for idx, task := range tasks {
		e.TaskChan <- Payload[T]{ID: idx, Value: task}
	}
	e.ControlChan <- ControlSignal{Source: "executor"}
}

// worker 工作线
func (e *Executor[T, R]) worker(taskPayload Payload[T], sem chan struct{}, done chan struct{}) {
	defer func() {
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

// runner 运行器
func (e *Executor[T, R]) runner() {
	sem := make(chan struct{}, e.numWorkers)  // 控制并发数
	done := make(chan struct{}, e.numWorkers) // 控制worker完成信号

	inputClosed := false
	inFlight := 0
	shouldFinish := func() bool {
		return inputClosed && inFlight == 0 && e.IsFinish()
	}

	for {
		if shouldFinish() {
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

// Start 启动执行器
func (e *Executor[T, R]) Start(tasks []T) {
	e.logSink.Start()
	e.logSource.StartExecutor(e.Name, len(tasks))

	e.SetTotal(len(tasks))
	e.notifyStart()

	go e.inputTask(tasks)
	e.runner()

	e.notifyFinish()
	close(e.SuccChan)
	close(e.ErrChan)

	e.logSource.EndExecutor(e.Name, 0, int(e.success.Load()), int(e.failed.Load()))
	e.logSink.Stop()
}

// NewExecutor 创建执行器
func NewExecutor[T any, R any](name string, processor func(T) (R, error), numWorkers int, observers ...Observer) *Executor[T, R] {
	logSink := pipline.NewSink(&LogRecordHandler{}, 100, time.Second)
	logSource := NewLogSource(logSink.GetQueue(), time.Second, "INFO")

	return &Executor[T, R]{
		Name:       name,
		processor:  processor,
		numWorkers: numWorkers,

		TaskChan:    make(chan Payload[T], numWorkers),
		SuccChan:    make(chan Payload[R], numWorkers),
		ErrChan:     make(chan ExecuteError, numWorkers),
		ControlChan: make(chan ControlSignal, numWorkers),

		observers: observers,
		logSink:   logSink,
		logSource: logSource,
	}
}
