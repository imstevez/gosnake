package base

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type Limit2D struct {
	Minx, Maxx, Miny, Maxy int
}

func (lmt Limit2D) GetCenter() Position2D {
	return Position2D{
		X: lmt.Minx + (lmt.Maxx-lmt.Minx)/2,
		Y: lmt.Miny + (lmt.Maxy-lmt.Miny)/2,
	}
}

func (lmt Limit2D) GetRandom() Position2D {
	return Position2D{
		X: rand.Intn(lmt.Maxx-lmt.Minx+1) + lmt.Minx,
		Y: rand.Intn(lmt.Maxy-lmt.Miny+1) + lmt.Miny,
	}
}
