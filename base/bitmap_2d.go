package base

import (
	"fmt"
)

const BitsPerByte = 8

type Bitmap2D [][]byte

func (bm *Bitmap2D) String() string {
	s := ""
	for i := 0; i < len(*bm); i++ {
		for j := 0; j < len((*bm)[i]); j++ {
			x := (*bm)[i][j]
			s += fmt.Sprintf("%08b ", uint8(x))
		}
		s += "\n"
	}
	return s
}

func (bm *Bitmap2D) Set(pos Position2D, v bool) {
	ny := pos.Y - len(*bm) + 1
	if ny > 0 {
		ya := make([][]byte, ny)
		*bm = append(*bm, ya...)
	}
	nx := pos.X/BitsPerByte - len((*bm)[pos.Y]) + 1
	if nx > 0 {
		xa := make([]byte, nx)
		(*bm)[pos.Y] = append((*bm)[pos.Y], xa...)
	}
	nBytes, nBits := pos.X/BitsPerByte, pos.X%BitsPerByte
	if v {
		(*bm)[pos.Y][nBytes] |= bm.littleEndian(nBits)
	} else {
		(*bm)[pos.Y][nBytes] &= ^(bm.littleEndian(nBits))
	}
}

func (bm *Bitmap2D) littleEndian(nBits int) byte {
	return 128 >> byte(nBits)
}

func (bm *Bitmap2D) Get(pos Position2D) (v bool) {
	nBytes, nBits := pos.X/BitsPerByte, pos.X%BitsPerByte
	if pos.Y >= len(*bm) || nBytes >= len((*bm)[pos.Y]) {
		return
	}
	v = (*bm)[pos.Y][nBytes]&bm.littleEndian(nBits) != 0
	return
}

func (bm *Bitmap2D) Add(bmx *Bitmap2D) {
	for i := 0; i < len((*bmx)); i++ {
		if i >= len(*bm) {
			for k := i; k < len(*bmx); k++ {
				dst := make([]byte, len((*bmx)[k]))
				copy(dst, (*bmx)[k])
				(*bm) = append((*bm), dst)
			}
		}
		for j := 0; j < len((*bmx)[i]); j++ {
			if j >= len((*bm)[i]) {
				(*bm)[i] = append((*bm)[i], (*bmx)[i][j:]...)
				break
			}
			(*bm)[i][j] |= (*bmx)[i][j]
		}
	}
}

func (bm *Bitmap2D) Minus(bmx *Bitmap2D) {
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
