package grow

// Payload is the unified data carrier for pipeline stages.
// ID tracks the task through the pipeline for tracing back to the original input.
// Prev stores the previous task when a stage emits derived results.
type Payload[V any] struct {
	ID    int
	Value V
	Prev  any
}

// TaskResult stores the current result together with its originating task.
type TaskResult[T any, R any] struct {
	Task   T
	Result R
}
