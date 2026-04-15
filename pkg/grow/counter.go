package grow

import "sync/atomic"

// Counter 并发安全的种子计数器，跟踪总数、成功数和失败数。
type Counter struct {
	total   atomic.Int64
	success atomic.Int64
	failed  atomic.Int64
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) SetTotal(total int) {
	c.total.Store(int64(total))
}

func (c *Counter) AddTotal(addNNum int) {
	c.total.Add(int64(addNNum))
}

func (c *Counter) AddSuccess(addNNum int) {
	c.success.Add(int64(addNNum))
}

func (c *Counter) AddFailed(addNNum int) {
	c.failed.Add(int64(addNNum))
}

func (c *Counter) GetTotal() int {
	return int(c.total.Load())
}

func (c *Counter) GetSuccess() int {
	return int(c.success.Load())
}

func (c *Counter) GetFailed() int {
	return int(c.failed.Load())
}

func (c *Counter) GetCompleted() int {
	return c.GetSuccess() + c.GetFailed()
}

func (c *Counter) IsFinish() bool {
	total := c.GetTotal()
	return total > 0 && c.GetCompleted() == total
}
