package base

import (
	"math/bits"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type Limit2D struct {
	Minx, Maxx, Miny, Maxy uint
}

func (lmt Limit2D) GetCenter() Position2D {
	return Position2D{
		X: lmt.Minx + (lmt.Maxx-lmt.Minx)/2,
		Y: lmt.Miny + (lmt.Maxy-lmt.Miny)/2,
	}
}

func (lmt Limit2D) GetRandom() Position2D {
	return Position2D{
		X: lmt.RandNum()%(lmt.Maxx-lmt.Minx+1) + lmt.Minx,
		Y: lmt.RandNum()%(lmt.Maxy-lmt.Miny+1) + lmt.Miny,
	}
}

func (lmt Limit2D) RandNum() uint {
	if bits.UintSize == 32 {
		return uint(rand.Uint32())
	}
	return uint(rand.Uint64())
}
