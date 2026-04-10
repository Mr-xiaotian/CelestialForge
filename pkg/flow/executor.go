package flow

// ExecuteTask 输入类型
type ExecuteTask[T any] struct {
	ID   int
	Task T
}

// ExecuteResult 输出类型
type ExecuteResult[T any, R any] struct {
	Task   T
	Result R
}

// ExecuteError 错误包装
type ExecuteError[T any] struct {
	Task  T
	Error error
}

type Executor[T any, R any] struct {
	processor  func(T) (R, error) // 可替换的处理函数
	numWorkers int

	TaskChan   chan ExecuteTask[T]
	ResultChan chan ExecuteResult[T, R]
	ErrorChan  chan ExecuteError[T]
}

func (e *Executor[T, R]) worker() {
	for task := range e.TaskChan {
		result, err := e.processor(task.Task)
		if err != nil {
			e.handleTaskError(ExecuteError[T]{Task: task.Task, Error: err})
		} else {
			e.processTaskSuccess(ExecuteResult[T, R]{Task: task.Task, Result: result})
		}
	}
}

func (e *Executor[T, R]) processTaskSuccess(result ExecuteResult[T, R]) {
	e.ResultChan <- result
}

func (e *Executor[T, R]) handleTaskError(err ExecuteError[T]) {
	e.ErrorChan <- err
}

func (e *Executor[T, R]) Start(tasks []T) {
	for i := 0; i < e.numWorkers; i++ {
		go e.worker()
	}

	// 发送任务
	for idx, task := range tasks {
		e.TaskChan <- ExecuteTask[T]{ID: idx, Task: task}
	}
	close(e.TaskChan)
}

func NewExecutor[T any, R any](processor func(T) (R, error), numWorkers int) *Executor[T, R] {
	TaskChan := make(chan ExecuteTask[T], numWorkers)
	resultChan := make(chan ExecuteResult[T, R], numWorkers)
	errorChan := make(chan ExecuteError[T], numWorkers)

	return &Executor[T, R]{
		processor:  processor,
		numWorkers: numWorkers,
		TaskChan:   TaskChan,
		ResultChan: resultChan,
		ErrorChan:  errorChan,
	}
}
