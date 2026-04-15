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

func (c *Counter) SetTotal(total int) {
	c.total = total
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
	return c.GetCompleted() == c.GetTotal()
}
