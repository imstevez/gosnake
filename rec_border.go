package gosnake

type RecBorder struct {
	width, height int
	symbol        string
}

func NewRecBorder(width, height int, symbol string) *RecBorder {
	return &RecBorder{
		width:  width,
		height: height,
		symbol: symbol,
	}
}

func (b *RecBorder) IsTaken(pos Position) bool {
	return (pos.x == 0 && pos.y < b.height) ||
		(pos.x == b.width-1 && pos.y < b.height) ||
		(pos.y == 0 && pos.x < b.width) ||
		(pos.y == b.height-1 && pos.x < b.width)
}

func (b *RecBorder) GetSymbolAt(pos Position) string {
	return IfStr(
		b.IsTaken(pos), b.symbol, "",
	)
}
