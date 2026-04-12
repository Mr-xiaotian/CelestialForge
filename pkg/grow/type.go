package grow

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
