package grow

const (
	SignalNone = iota
	SignalSeal
)

// Payload 管道阶段的统一数据载体。
// ID 用于在管道中追踪种子到原始输入。
// Prev 存储上一阶段的种子值。
// Signal 用于传递控制信号（SignalNone=正常数据, SignalSeal=终止）。
// Source 标识信号来源的 Plot 名称。
type Payload[V any] struct {
	ID     int
	Value  V
	Prev   any
	Signal int
	Source string
}

// Karma 种子与果实的配对，记录一颗种子培育后的结果。
type Karma[S any, F any] struct {
	Seed  S
	Fruit F
}
