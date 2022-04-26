package gosnake

import (
	"math/rand"
	"time"
)

type Limit struct {
	MinX, MaxX int
	MinY, MaxY int
}

type Food struct {
	pos   Position
	limit Limit
}

func init() {
	rand.Seed(time.Now().Unix())
}

func NewFood(limit Limit) *Food {
	food := &Food{
		limit: limit,
	}
	food.UpdatePos()
	return food
}

func (f *Food) UpdatePos() {
	f.pos.X = rand.Intn(f.limit.MaxX-f.limit.MinX+1) + f.limit.MinX
	f.pos.Y = rand.Intn(f.limit.MaxY-f.limit.MinY+1) + f.limit.MinY
}

func (f *Food) GetTakes() map[Position]struct{} {
	return map[Position]struct{}{f.pos: {}}
}

func (f *Food) IsTaken(pos Position) bool {
	return f.pos == pos
}
