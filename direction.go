package gosnake

import "gosnake/keys"

type Direction int

const (
	DirUp Direction = iota + 1
	DirRight
	DirDown
	DirLeft
)

var KeyCodeToDir = map[keys.Code]Direction{
	keys.CodeUp:    DirUp,
	keys.CodeRight: DirRight,
	keys.CodeDown:  DirDown,
	keys.CodeLeft:  DirLeft,
}

func (dir Direction) RevertTo(other Direction) bool {
	diff := dir - other
	return diff == 2 || diff == -2
}
