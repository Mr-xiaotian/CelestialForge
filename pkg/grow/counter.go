package grow

import "sync/atomic"

type Counter struct {
	total   int
	success atomic.Int64
	failed  atomic.Int64
}

func NewCounter() *Counter {
	return &Counter{}
}

func (c *Counter) AddSuccess(addNNum int) {
	c.success.Add(int64(addNNum))
}

func (c *Counter) AddFailed(addNNum int) {
	c.failed.Add(int64(addNNum))
}

func (c *Counter) GetTotal() int {
	return c.total
}

func (c *Counter) GetComplated() int {
	return int(c.success.Load() + c.failed.Load())
}
