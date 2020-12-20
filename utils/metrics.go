package utils

import "sync/atomic"

//Counter simple
type Counter int32

//Inc iccrement by one
func (c *Counter) Inc() int32 {
	return atomic.AddInt32((*int32)(c), 1)
}

//Get get current value
func (c *Counter) Get() int32 {
	return atomic.LoadInt32((*int32)(c))
}

//Dec decrement counter
func (c *Counter) Dec(previousValue int32) int32 {
	return atomic.AddInt32((*int32)(c), -previousValue)
}

//Reset reset counter
func (c *Counter) Reset() float64 {
	countValue := c.Get()
	if countValue == 0 {
		return 0
	}
	c.Dec(countValue)
	return float64(countValue)
}
