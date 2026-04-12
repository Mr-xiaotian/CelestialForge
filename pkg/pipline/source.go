package pipline

import (
	"context"
	"errors"
	"sync"
	"time"
)

// 定义抽象行为的 Interface
type RecordHandler[T any] interface {
	BeforeStart() error
	HandleRecord(record T) error
	AfterStop() error
}

// 基础结构体，包含状态和通用逻辑
type Source[T any] struct {
	// 状态字段
	ch      chan T
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	timeout time.Duration

	// 依赖：具体的处理逻辑（接口注入）
	handler RecordHandler[T]
}

// 构造函数
func NewSource[T any](handler RecordHandler[T], bufferSize int, timeout time.Duration) *Source[T] {
	ctx, cancel := context.WithCancel(context.Background())
	return &Source[T]{
		ch:      make(chan T, bufferSize),
		ctx:     ctx,
		cancel:  cancel,
		handler: handler,
		timeout: timeout,
	}
}

// GetQueue 返回写入通道
func (b *Source[T]) GetQueue() chan<- T {
	return b.ch
}

// Start 启动监听
func (b *Source[T]) Start() error {
	if err := b.handler.BeforeStart(); err != nil {
		return err
	}

	b.wg.Add(1)
	go b.listen()
	return nil
}

// listen 持续消费通道中的记录
func (b *Source[T]) listen() {
	defer b.wg.Done()

	for {
		select {
		case record, ok := <-b.ch:
			if !ok {
				// 通道关闭，优雅退出
				return
			}
			b.handler.HandleRecord(record)

		case <-b.ctx.Done():
			// 收到取消信号
			return
		}
	}
}

// Stop 停止监听
func (b *Source[T]) Stop() error {
	close(b.ch) // 先尝试优雅关闭

	// 等待 Handler 处理完，但最多等 timeout 秒
	done := make(chan struct{})
	go func() { b.wg.Wait(); close(done); b.handler.AfterStop() }()

	select {
	case <-done:
		return nil // 正常结束
	case <-time.After(b.timeout):
		// ⚠️ 超时：触发 ctx.Done()，强制 listen() 退出
		b.cancel()
		return errors.New("shutdown timeout")
	}
}
