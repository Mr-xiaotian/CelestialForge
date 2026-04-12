package grow

type Executor[T any, R any] struct {
	processor  func(T) (R, error)
	numWorkers int

	TaskChan    chan Payload[T]
	SuccChan    chan Payload[R]
	ErrChan     chan ExecuteError
	ControlChan chan ControlSignal

	observers []Observer
	Counter
}

func (e *Executor[T, R]) reportProgress() {
	completed := e.GetComplated()
	total := e.GetTotal()
	for _, observer := range e.observers {
		observer.OnProgress(completed, total)
	}
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

func (e *Executor[T, R]) inputTask(tasks []T) {
	for idx, task := range tasks {
		e.TaskChan <- Payload[T]{ID: idx, Value: task}
	}
	e.ControlChan <- ControlSignal{Source: "executor"}
}

func (e *Executor[T, R]) worker(task Payload[T], sem chan struct{}, done chan struct{}) {
	defer func() {
		<-sem // 释放并发令牌
		done <- struct{}{}
	}()
	result, err := e.processor(task.Value)
	if err != nil {
		e.handleTaskError(task, err)
	} else {
		e.processTaskSuccess(task, result)
	}
}

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
			if inFlight > 0 {
				inFlight--
			}
		case <-e.ControlChan:
			inputClosed = true
		}
	}
}

func (e *Executor[T, R]) Start(tasks []T) {
	e.SetTotal(len(tasks))
	e.notifyStart()

	go e.inputTask(tasks)
	e.runner()

	e.notifyFinish()
	close(e.SuccChan)
	close(e.ErrChan)
}

func NewExecutor[T any, R any](processor func(T) (R, error), numWorkers int, observers ...Observer) *Executor[T, R] {
	return &Executor[T, R]{
		processor:  processor,
		numWorkers: numWorkers,

		TaskChan:    make(chan Payload[T], numWorkers),
		SuccChan:    make(chan Payload[R], numWorkers),
		ErrChan:     make(chan ExecuteError, numWorkers),
		ControlChan: make(chan ControlSignal, numWorkers),

		observers: observers,
	}
}
