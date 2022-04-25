package gosnake

type Direction int

const (
	DirUp Direction = iota
	DirRight
	DirDown
	DirLeft
)

func (dir Direction) Oppsite(other Direction) bool {
	diff := dir - other
	return diff == 2 || diff == -2
}
