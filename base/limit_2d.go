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
	MinX, MaxX uint
	MinY, MaxY uint
}

func (lmt Limit2D) CenterPos() Position2D {
	return Position2D{
		X: lmt.MinX + (lmt.MaxX-lmt.MinX)/2,
		Y: lmt.MinY + (lmt.MaxY-lmt.MinY)/2,
	}
}

func (lmt Limit2D) RandPos() Position2D {
	return Position2D{
		X: randUint()%(lmt.MaxX-lmt.MinX+1) + lmt.MinX,
		Y: randUint()%(lmt.MaxY-lmt.MinY+1) + lmt.MinY,
	}
}

func randUint() uint {
	if bits.UintSize == 64 {
		return uint(rand.Uint64())
	}
	return uint(rand.Uint32())
}
