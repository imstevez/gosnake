package gosnake

import (
	"math/rand"
	"time"
)

type Food struct {
	pos    Position
	symbol string
	rminx  int
	rmaxx  int
	rminy  int
	rmaxy  int
}

func init() {
	rand.Seed(time.Now().Unix())
}

func NewFood(rminx, rmaxx, rminy, rmaxy int, symbol string) *Food {
	food := &Food{
		symbol: symbol,
		rminx:  rminx,
		rmaxx:  rmaxx,
		rminy:  rminy,
		rmaxy:  rmaxy,
	}
	food.UpdatePos()
	return food
}

func (f *Food) UpdatePos() {
	rx := f.rmaxx - f.rminx + 1
	ry := f.rmaxy - f.rminy + 1
	f.pos.x = rand.Intn(rx) + f.rminx
	f.pos.y = rand.Intn(ry) + f.rminy
}

func (f *Food) IsTaken(pos Position) bool {
	return f.pos == pos
}

func (f *Food) GetSymbol() string {
	return f.symbol
}
