package gosnake

import (
	"fmt"
)

type Layer interface {
	IsTaken(Position) bool
	GetSymbol() string
}

type Ground struct {
	width, height int
	symbol        string
}

func NewGround(width, height int, symbol string) *Ground {
	return &Ground{
		width:  width,
		height: height,
		symbol: symbol,
	}
}

func (g *Ground) Render(layers ...Layer) {
	g.Clear()
	s := ""
	for y := 0; y < g.height; y++ {
		s += "\r"
		for x := 0; x < g.width; x++ {
			pos := Position{x, y}
			symbol := g.symbol
			for _, l := range layers {
				if l.IsTaken(pos) {
					symbol = l.GetSymbol()
				}
			}
			s += symbol
		}
		s += "\n\r"
	}
	fmt.Print(s)
}

func (g *Ground) Clear() {
	for i := 0; i < 35; i++ {
		fmt.Print("\033[A")
	}
}
