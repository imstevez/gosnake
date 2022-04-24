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
	pos    Position
	symbol string
	limit  Limit
}

func init() {
	rand.Seed(time.Now().Unix())
}

func NewFood(symbol string, limit Limit) *Food {
	food := &Food{
		symbol: symbol,
		limit:  limit,
	}
	food.UpdatePos()
	return food
}

func (f *Food) UpdatePos() {
	f.pos.X = rand.Intn(f.limit.MaxX-f.limit.MinX+1) + f.limit.MinX
	f.pos.Y = rand.Intn(f.limit.MaxY-f.limit.MinY+1) + f.limit.MinY
}

func (f *Food) IsTaken(pos Position) bool {
	return f.pos == pos
}

func (f *Food) GetSymbolAt(pos Position) string {
	return IfStr(
		f.IsTaken(pos), f.symbol, "",
	)
}
