package base

type Flag16 uint16

func (flag *Flag16) Set(slag Flag16) {
	*flag |= slag
}

func (flag *Flag16) UnSet(slag Flag16) {
	*flag |= ^slag
}

func (flag Flag16) Is(slag Flag16) bool {
	return flag&slag == slag
}
