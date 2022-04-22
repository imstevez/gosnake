package gosnake

import (
	"sync/atomic"
)

type gameStatusCode int32

const (
	gameStatusCodeStopped gameStatusCode = iota
	gameStatusCodeRunning
)

func (c *gameStatusCode) setRunningFromStopped() bool {
	return atomic.CompareAndSwapInt32((*int32)(c), int32(gameStatusCodeStopped), int32(gameStatusCodeRunning))
}

func (c *gameStatusCode) setStopped() {
	atomic.StoreInt32((*int32)(c), int32(gameStatusCodeStopped))
}
