package gosnake

import (
	"sync/atomic"
)

type gameStatusCode int32

const (
	gameStatusCodeStopped gameStatusCode = iota
	gameStatusCodeRunning
	gameStatusCodePaused
	gameStatusCodeOver
)

func (c *gameStatusCode) setRunningFromStopped() bool {
	return atomic.CompareAndSwapInt32((*int32)(c), int32(gameStatusCodeStopped), int32(gameStatusCodeRunning))
}

func (c *gameStatusCode) setPaused() {
	atomic.StoreInt32((*int32)(c), int32(gameStatusCodePaused))
}

func (c *gameStatusCode) setRunning() {
	atomic.StoreInt32((*int32)(c), int32(gameStatusCodeRunning))
}

func (c *gameStatusCode) setStopped() {
	atomic.StoreInt32((*int32)(c), int32(gameStatusCodeStopped))
}

func (c *gameStatusCode) setOver() {
	atomic.StoreInt32((*int32)(c), int32(gameStatusCodeOver))
}

func (c *gameStatusCode) isPaused() bool {
	return atomic.LoadInt32((*int32)(c)) == int32(gameStatusCodePaused)
}

func (c *gameStatusCode) isRunning() bool {
	return atomic.LoadInt32((*int32)(c)) == int32(gameStatusCodeRunning)
}

func (c *gameStatusCode) isOver() bool {
	return atomic.LoadInt32((*int32)(c)) == int32(gameStatusCodeOver)
}
