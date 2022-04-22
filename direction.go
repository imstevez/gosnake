package gosnake

import "gosnake/keys"

type Direction int

const (
	DirUp Direction = iota + 1
	DirRight
	DirDown
	DirLeft
)

var keyCodeToDir = map[keys.Code]Direction{
	keys.CodeUp:     DirUp,
	keys.CodeRight:  DirRight,
	keys.CodeDown:   DirDown,
	keys.CodeLeft:   DirLeft,
	keys.CodeUp2:    DirUp,
	keys.CodeRight2: DirRight,
	keys.CodeDown2:  DirDown,
	keys.CodeLeft2:  DirLeft,
}

func (dir Direction) Oppsite(other Direction) bool {
	diff := dir - other
	return diff == 2 || diff == -2
}
