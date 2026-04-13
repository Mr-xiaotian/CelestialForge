package pipline

import (
	"context"
	"fmt"
	"time"
)

// Source 生产端，向通道写入记录
type Source[T any] struct {
	ch      chan<- T
	ctx     context.Context
	cancel  context.CancelFunc
	timeout time.Duration
}

func NewSource[T any](ch chan<- T, timeout time.Duration) *Source[T] {
	ctx, cancel := context.WithCancel(context.Background())
	return &Source[T]{ch: ch, timeout: timeout, ctx: ctx, cancel: cancel}
}

// Send 发送记录，支持上下文取消和超时控制
func (s *Source[T]) Send(record T) error {
	select {
	case s.ch <- record:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	case <-time.After(s.timeout):
		return fmt.Errorf("source send timeout after %v", s.timeout)
	}
}

// Close 关闭 Source，后续 Send 调用将返回错误
func (s *Source[T]) Close() {
	s.cancel()
}
