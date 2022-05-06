package base

import (
	"fmt"
	"math/bits"
)

type Bitmap2D [][]uint

func (bm *Bitmap2D) String() string {
	s := ""
	format := fmt.Sprintf("%%0%db", bits.UintSize)
	for i := 0; i < len(*bm); i++ {
		for j := 0; j < len((*bm)[i]); j++ {
			x := (*bm)[i][j]
			s += "\r" + fmt.Sprintf(format, x)
		}
		s += "\n"
	}
	return s
}

// little endian mask
func (bm *Bitmap2D) mask(nBits uint) uint {
	return 1 << (bits.UintSize - 1) >> nBits
}

func (bm *Bitmap2D) Set(pos Position2D, value bool) {
	if pos.Y+1 > uint(len(*bm)) {
		yn := pos.Y + 1 - uint(len(*bm))
		tmp := make(Bitmap2D, yn)
		*bm = append(*bm, tmp...)
	}
	if pos.X/bits.UintSize+1 > uint(len((*bm)[pos.Y])) {
		xn := pos.X/bits.UintSize + 1 - uint(len((*bm)[pos.Y]))
		tmp := make([]uint, xn)
		(*bm)[pos.Y] = append((*bm)[pos.Y], tmp...)
	}
	nWords, nBits := pos.X/bits.UintSize, pos.X%bits.UintSize
	word := &((*bm)[pos.Y][nWords])
	if value {
		*word |= bm.mask(nBits)
	} else {
		*word &= ^(bm.mask(nBits))
	}
}

func (bm *Bitmap2D) Get(pos Position2D) bool {
	nWords, nBits := pos.X/bits.UintSize, pos.X%bits.UintSize
	if pos.Y >= uint(len(*bm)) || nWords >= uint(len((*bm)[pos.Y])) {
		return false
	}
	return (*bm)[pos.Y][nWords]&bm.mask(nBits) != 0
}

func (bm *Bitmap2D) Stack(bmx *Bitmap2D) {
	yn := len(*bmx) - len(*bm)
	if yn > 0 {
		tmp := make([][]uint, yn)
		*bm = append(*bm, tmp...)
	}
	for i := 0; i < len((*bmx)); i++ {
		for j := 0; j < len((*bmx)[i]); j++ {
			if j >= len((*bm)[i]) {
				(*bm)[i] = append((*bm)[i], (*bmx)[i][j:]...)
				break
			}
			(*bm)[i][j] |= (*bmx)[i][j]
		}
	}
}

func (bm *Bitmap2D) Cull(bmx *Bitmap2D) {
	for i := 0; i < len(*bmx); i++ {
		if i >= len(*bm) {
			return
		}
		for j := 0; j < len((*bmx)[i]); j++ {
			if j >= len((*bm)[i]) {
				break
			}
			(*bm)[i][j] &= ^((*bmx)[i][j])
		}
	}
}
