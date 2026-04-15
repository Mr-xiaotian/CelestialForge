package grow

// Payload 管道阶段的统一数据载体。
// ID 用于在管道中追踪种子到原始输入。
// Prev 存储上一阶段的种子值。
type Payload[V any] struct {
	ID    int
	Value V
	Prev  any
}

// Karma 种子与果实的配对，记录一颗种子培育后的结果。
type Karma[T any, R any] struct {
	Seed  T
	Fruit R
}
