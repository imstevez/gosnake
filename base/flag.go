package base

type Flag8 uint8

func (flag *Flag8) Set(sFlag Flag8) {
	*flag |= sFlag
}

func (flag *Flag8) UnSet(sFlag Flag8) {
	*flag |= ^sFlag
}

func (ps *PlayerStatusFlag) Is(flag PlayerStatusFlag) bool {
	return *ps&flag == flag
}
