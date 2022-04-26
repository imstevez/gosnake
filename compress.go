package gosnake

type CompressLayer struct {
	Takes  []byte
	W      int
	H      int
	Symbol string
}

func NewCompressLayer(w, h int) *CompressLayer {
	lineBytes := IfInt(w%8 == 0, w/8, w/8+1)
	layer := &CompressLayer{
		W:     w,
		H:     h,
		Takes: make([]byte, lineBytes*h),
	}
	return layer
}

func (cl *CompressLayer) SetSymbol(symbol string) {
	cl.Symbol = symbol
}

func (cl *CompressLayer) AddPositions(postions map[Position]struct{}) {
	for pos := range postions {
		if pos.X >= cl.W || pos.Y >= cl.H {
			continue
		}
		nbyte, nbit := cl.getPosOffset(pos)
		cl.Takes[nbyte] |= (1 << nbit)
	}
}

func (cl *CompressLayer) IsTaken(pos Position) bool {
	nbyte, nbit := cl.getPosOffset(pos)
	return cl.Takes[nbyte]&(1<<nbit) != 0
}

func (cl *CompressLayer) GetSymbolAt(pos Position) string {
	return IfStr(cl.IsTaken(pos), cl.Symbol, "")
}

func (cl *CompressLayer) getPosOffset(pos Position) (nbyte, nbit int) {
	lineBytes := IfInt(cl.W%8 == 0, cl.W/8, cl.W/8+1)
	offset := pos.Y*lineBytes*8 + pos.X
	nbyte = offset / 8
	nbit = 8 - (offset % 8) - 1
	return
}
