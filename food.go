package gosnake

import (
	"math/rand"
	"time"
)

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
	rx := f.limit.MaxX - f.limit.MinX + 1
	ry := f.limit.MaxY - f.limit.MinY + 1
	f.pos.x = rand.Intn(rx) + f.limit.MinX
	f.pos.y = rand.Intn(ry) + f.limit.MinY
}

func (f *Food) GetSymbolAt(pos Position) string {
	return NoitherStr(f.pos == pos, f.symbol, "")
}
