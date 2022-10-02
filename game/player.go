package game

type CMD int

type Player struct {
	input <-chan CMD
}
