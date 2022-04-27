package gosnake

type Layer interface {
	GetSymbolAt(Position) string
}

type Ground struct {
	width  int
	height int
	symbol string
}

func NewGround(width, height int, symbol string) *Ground {
	return &Ground{
		width: width, height: height, symbol: symbol,
	}
}

func (g *Ground) Render(layers ...Layer) (ls Lines) {
	ls = make([]string, g.height)
	for y := 0; y < g.height; y++ {
		l := ""
		for x := 0; x < g.width; x++ {
			pos := Position{x, y}
			symbol := g.symbol
			for _, layer := range layers {
				sbl := layer.GetSymbolAt(pos)
				if sbl != "" {
					symbol = sbl
				}
			}
			l += symbol
		}
		ls[y] = l
	}
	return
}
