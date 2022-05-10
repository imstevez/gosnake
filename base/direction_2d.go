package base

import "math/rand"

type Direction2D int

const (
	Dir2DUp Direction2D = iota + 1
	Dir2DRight
	Dir2DDown
	Dir2DLeft
)

func RandDir2D() Direction2D {
	return Direction2D(
		rand.Intn(4) + 1,
	)
}

func (dir Direction2D) OppositeTo(otherDir Direction2D) bool {
	diff := dir - otherDir
	return diff == 2 || diff == -2
}
