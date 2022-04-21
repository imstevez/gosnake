package gosnake

import (
	"strings"
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

func (g *Ground) Render(layers ...Layer) (result string) {
	for y := 0; y < g.height; y++ {
		result += "\r"
		for x := 0; x < g.width; x++ {
			pos := Position{x, y}
			symbol := g.symbol
			for _, l := range layers {
				if l.IsTaken(pos) {
					symbol = l.GetSymbol()
				}
			}
			result += symbol
		}
		result += "\n\r"
	}
	line := strings.Count(result, "\n") + 10
	for i := 0; i < line; i++ {
		result = "\033[A" + result
	}
	return result
}
