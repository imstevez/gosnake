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
	return (pos.X == 0 && pos.Y < b.height) ||
		(pos.X == b.width-1 && pos.Y < b.height) ||
		(pos.Y == 0 && pos.X < b.width) ||
		(pos.Y == b.height-1 && pos.X < b.width)
}

func (b *RecBorder) GetSymbolAt(pos Position) string {
	return IfStr(
		b.IsTaken(pos), b.symbol, "",
	)
}
