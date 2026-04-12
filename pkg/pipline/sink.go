package pipline

import (
	"fmt"
	"time"
)

// Sink 的转换
type Sink[T any] struct {
	ch      chan<- T
	timeout time.Duration
}

func NewSink[T any](ch chan<- T, timeout time.Duration) *Sink[T] {
	return &Sink[T]{ch: ch, timeout: timeout}
}

func (s *Sink[T]) Sink(record T) {
	// 非阻塞写入，或带超时/上下文控制
	select {
	case s.ch <- record:
	case <-time.After(s.timeout):
		fmt.Printf("Sink timeout after %v\n", s.timeout)
	}
}
