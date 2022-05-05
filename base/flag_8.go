package base

type Flag8 uint8

func (flag *Flag8) Set(slag Flag8) {
	*flag |= slag
}

func (flag *Flag8) UnSet(slag Flag8) {
	*flag &= (^slag)
}

func (flag Flag8) Is(slag Flag8) bool {
	return flag&slag == slag
}
