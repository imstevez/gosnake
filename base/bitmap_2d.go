package base

const bitsPerByte = 8

type Bitmap2D struct {
	bytesPerRow int
	width       int
	height      int
	data        []byte
}

func NewBitmap(bytesPerRow int) *Bitmap2D {
	if bytesPerRow < 0 {
		panic("bytes per row cannot be less than 0")
	}
	return &Bitmap2D{
		bytesPerRow: bytesPerRow,
		width:       bytesPerRow * bitsPerByte,
	}
}

func (bm *Bitmap2D) offsetOf(x, y int) (nBytes, nBits int) {
	nBytes = (y * bm.bytesPerRow) + (x / bitsPerByte)
	nBits = x % bitsPerByte
	if nBytes+nBits < 0 {
		panic("offset less than 0")
	}
	return
}

func (bm *Bitmap2D) Set(x, y int, v bool) {
	nBytes, nBits := bm.offsetOf(x, y)
	if nBytes > len(bm.data) {
		add := make([]byte, nBytes-len(bm.data))
		bm.data = append(bm.data, add...)
	}
	if v {
		bm.data[nBytes] |= (byte(1) << byte(nBits))
	} else {
		bm.data[nBytes] &= ^(byte(1) << byte(nBits))
	}
}

func (bm *Bitmap2D) Get(x, y int) (v bool) {
	nBytes, nBits := bm.offsetOf(x, y)
	if nBytes > len(bm.data) {
		return
	}
	v = bm.data[nBytes]&(byte(1)<<byte(nBits)) != 0
	return
}

func (bm *Bitmap2D) Add(data []byte) {
	for i := 0; i < len(data); i++ {
		if i >= len(bm.data) {
			bm.data = append(bm.data, data[i:]...)
			return
		}
		bm.data[i] |= data[i]
	}
}

func (bm *Bitmap2D) Minus(data []byte) {
	for i := 0; i < len(data); i++ {
		if i >= len(bm.data) {
			return
		}
		bm.data[i] &= ^(data[i])
	}
}
