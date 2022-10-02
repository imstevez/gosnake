package base

import "sync/atomic"

type AtomicStatus int32

func (ats *AtomicStatus) Is(status int32) bool {
	return atomic.LoadInt32((*int32)(ats)) == status
}

func (ats *AtomicStatus) Set(status int32) {
	atomic.StoreInt32((*int32)(ats), status)
}

func (ats *AtomicStatus) SetFrom(from, to int32) (set bool) {
	return atomic.CompareAndSwapInt32((*int32)(ats), from, to)
}
