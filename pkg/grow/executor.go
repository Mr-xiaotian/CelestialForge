package grow

import (
	"sync"
)

// Payload is the unified data carrier for pipeline stages.
// ID tracks the task through the pipeline for tracing back to the original input.
type Payload[V any] struct {
	ID    int
	Value V
}

// ExecuteError wraps a failed task
type ExecuteError struct {
	ID    int
	Error error
}

type Executor[T any, R any] struct {
	processor  func(T) (R, error)
	numWorkers int

	TaskChan chan Payload[T]
	SuccChan chan Payload[R]
	ErrChan  chan ExecuteError

	observers []Observer
	Counter
}

func (e *Executor[T, R]) worker() {
	for p := range e.TaskChan {
		result, err := e.processor(p.Value)
		if err != nil {
			e.handleTaskError(p, err)
		} else {
			e.processTaskSuccess(p, result)
		}
	}
}

func (e *Executor[T, R]) reportProgress() {
	completed := e.GetComplated()
	total := e.GetTotal()
	for _, observer := range e.observers {
		observer.OnProgress(completed, total)
	}
}

func (e *Executor[T, R]) UseObserver(observer Observer) *Executor[T, R] {
	if observer != nil {
		e.observers = append(e.observers, observer)
	}
	return e
}

func (e *Executor[T, R]) notifyStart() {
	total := e.GetTotal()
	for _, observer := range e.observers {
		observer.OnStart(total)
	}
}

func (e *Executor[T, R]) notifyFinish() {
	completed := e.GetComplated()
	total := e.GetTotal()
	for _, observer := range e.observers {
		observer.OnFinish(completed, total)
	}
}

func (e *Executor[T, R]) processTaskSuccess(taskPayload Payload[T], result R) {
	e.AddSuccess(1)
	e.reportProgress()
	e.SuccChan <- Payload[R]{ID: taskPayload.ID, Value: result}
}

func (e *Executor[T, R]) handleTaskError(taskPayload Payload[T], err error) {
	e.AddFailed(1)
	e.reportProgress()
	e.ErrChan <- ExecuteError{ID: taskPayload.ID, Error: err}
}

func (e *Executor[T, R]) Start(tasks []T) {
	e.total = len(tasks)
	e.notifyStart()

	var wg sync.WaitGroup
	for i := 0; i < e.numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			e.worker()
		}()
	}

	for idx, task := range tasks {
		e.TaskChan <- Payload[T]{ID: idx, Value: task}
	}
	close(e.TaskChan)

	wg.Wait()
	e.notifyFinish()
	close(e.SuccChan)
	close(e.ErrChan)
}

func NewExecutor[T any, R any](processor func(T) (R, error), numWorkers int, observers ...Observer) *Executor[T, R] {
	return &Executor[T, R]{
		processor:  processor,
		numWorkers: numWorkers,
		TaskChan:   make(chan Payload[T], numWorkers),
		SuccChan:   make(chan Payload[R], numWorkers),
		ErrChan:    make(chan ExecuteError, numWorkers),
		observers:  observers,
	}
}
