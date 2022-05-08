package base

import (
	"fmt"
	"math/bits"
	"strings"
)

const size = bits.UintSize

var wordLayout = fmt.Sprintf("%%0%db", size)

type word uint

func posmask(ibit uint) word {
	return 1 << (size - 1) >> ibit
}

func negmask(ibit uint) word {
	return ^posmask(ibit)
}

func (w *word) set(ibit uint, val bool) {
	if val {
		*w |= posmask(ibit)
	} else {
		*w &= negmask(ibit)
	}
}

func (w word) get(ibit uint) (val bool) {
	if ibit >= size {
		return
	}
	val = w&posmask(ibit) != 0
	return
}

func (w *word) merge(m word) {
	*w |= m
}

func (w *word) cull(m word) {
	*w &= ^m
}

func (w word) string() string {
	return fmt.Sprintf(wordLayout, w)
}

type words []word

func (ws words) nbits() uint {
	return uint(len(ws) * size)
}

func (ws *words) expand(nbits uint) {
	if ws.nbits() >= nbits {
		tmp := make(words, (nbits-ws.nbits())/size+1)
		*ws = append(*ws, tmp...)
	}
}

func (ws words) set(ibit uint, val bool) {
	ws.expand(ibit + 1)
	ws[ibit/size].set(ibit%size, val)
}

func (ws words) get(ibit uint) (val bool) {
	if ibit >= ws.nbits() {
		return
	}
	val = ws[ibit/size].get(ibit % size)
	return
}

func (ws words) merge(ms words) {
	ws.expand(ms.nbits())
	for i := 0; i < len(ms); i++ {
		ws[i].merge(ms[i])
	}
}

func (ws words) cull(ms words) {
	for i := 0; i < len(ms); i++ {
		if i >= len(ws) {
			return
		}
		ws[i].cull(ms[i])
	}
}

func (ws words) string() string {
	strs := make([]string, len(ws))
	for i := 0; i < len(ws); i++ {
		strs[i] = ws[i].string()
	}
	return strings.Join(strs, " ")
}

type Bitmap2D []words

func (bm Bitmap2D) String() string {
	strs := make([]string, len(bm))
	for i := 0; i < len(bm); i++ {
		strs[i] = "\r" + bm[i].string()
	}
	return strings.Join(strs, "\n")
}

func (bm Bitmap2D) nrows() uint {
	return uint(len(bm))
}

func (bm *Bitmap2D) expand(nrows uint) {
	if bm.nrows() > nrows {
		tmp := make(Bitmap2D, nrows-bm.nrows())
		*bm = append(*bm, tmp...)
	}
}

func (bm Bitmap2D) Set(pos Position2D, val bool) {
	bm.expand(pos.Y + 1)
	bm[pos.Y].set(pos.X, val)
}

func (bm Bitmap2D) Get(pos Position2D) (val bool) {
	if pos.Y >= bm.nrows() {
		return
	}
	val = bm[pos.Y].get(pos.X)
	return
}

func (bm Bitmap2D) merge(pm Bitmap2D) {
	bm.expand(pm.nrows())
	for i := 0; i < len(pm); i++ {
		bm[i].merge(pm[i])
	}
}

func (bm Bitmap2D) Cull(pm Bitmap2D) {
	for i := 0; i < len(pm); i++ {
		if i >= len(bm) {
			return
		}
		bm[i].cull(pm[i])
	}
}
