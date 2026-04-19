package grow

import "sync/atomic"

// Counter 并发安全的种子计数器，跟踪总数、成功数和失败数。
type Counter struct {
	seedNum  atomic.Int64
	fruitNum atomic.Int64
	weedNum  atomic.Int64
}

func NewCounter() *Counter {
	return &Counter{}
}

// ==== Set ====

// SetSeedNum 设置总数。
func (c *Counter) SetSeedNum(seedNum int) {
	c.seedNum.Store(int64(seedNum))
}

// ==== Add ====

// AddSeedNum 增加总数。
func (c *Counter) AddSeedNum(addNNum int) {
	c.seedNum.Add(int64(addNNum))
}

// AddFruitNum 增加成功数。
func (c *Counter) AddFruitNum(addNNum int) {
	c.fruitNum.Add(int64(addNNum))
}

// AddWeedNum 增加失败数。
func (c *Counter) AddWeedNum(addNNum int) {
	c.weedNum.Add(int64(addNNum))
}

// ==== Get ====

// GetSeedNum 获取总数。
func (c *Counter) GetSeedNum() int {
	return int(c.seedNum.Load())
}

// GetFruitNum 获取成功数。
func (c *Counter) GetFruitNum() int {
	return int(c.fruitNum.Load())
}

// GetWeedNum 获取失败数。
func (c *Counter) GetWeedNum() int {
	return int(c.weedNum.Load())
}

// GetCompleted 获取已完成数。
func (c *Counter) GetCompleted() int {
	return c.GetFruitNum() + c.GetWeedNum()
}

// ==== Is ====

// IsFinish 是否完成。
func (c *Counter) IsFinish() bool {
	return c.GetCompleted() == c.GetSeedNum()
}
