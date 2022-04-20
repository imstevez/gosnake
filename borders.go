package gosnake

type Borders struct {
	takes  map[Position]struct{}
	symbol string
}

func NewRecBorders(width, height int, symbol string) *Borders {
	takes := make(map[Position]struct{})
	for x := 0; x < width; x++ {
		takes[Position{x, 0}] = struct{}{}
		y := height - 1
		takes[Position{x, y}] = struct{}{}
	}
	for y := 1; y < height-1; y++ {
		takes[Position{0, y}] = struct{}{}
		x := width - 1
		takes[Position{x, y}] = struct{}{}
	}
	return &Borders{
		takes:  takes,
		symbol: symbol,
	}
}

func (b *Borders) IsTaken(pos Position) bool {
	_, ok := b.takes[pos]
	return ok
}

func (b *Borders) GetSymbol() string {
	return b.symbol
}
