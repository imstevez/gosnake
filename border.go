package gosnake

type Border struct {
	width, height int
	symbol        string
}

func NewRecBorder(width, height int, symbol string) *Border {
	return &Border{
		width:  width,
		height: height,
		symbol: symbol,
	}
}

func (b *Border) GetSymbolAt(pos Position) string {
	if (pos.x == 0 && pos.y < b.height) ||
		(pos.x == b.width-1 && pos.y < b.height) ||
		(pos.y == 0 && pos.x < b.width) ||
		(pos.y == b.height-1 && pos.x < b.width) {
		return b.symbol
	}
	return ""
}
