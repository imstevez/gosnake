package gosnake

import (
	"errors"
)

type Node struct {
	next *Node
	prev *Node
	pos  Position
}

type Limit struct {
	MinX, MaxX int
	MinY, MaxY int
}
type Snake struct {
	head     *Node
	tail     *Node
	prevTail *Node
	length   int
	dir      Direction
	takes    map[Position]struct{}
	symbol   string
	limit    Limit
}

func NewSnake(initialPosX, initialPosY int, initialDir Direction, symbol string, limit Limit) *Snake {
	pos := Position{
		x: initialPosX,
		y: initialPosY,
	}
	node := &Node{
		next: nil,
		prev: nil,
		pos:  pos,
	}
	return &Snake{
		head:     node,
		tail:     node,
		prevTail: node,
		length:   1,
		dir:      initialDir,
		symbol:   symbol,
		takes: map[Position]struct{}{
			pos: {},
		},
		limit: limit,
	}
}

func (s *Snake) GetDir() Direction {
	return s.dir
}

func (s *Snake) GetHeadPos() Position {
	return s.head.pos
}

func (s *Snake) Len() int {
	return s.length
}

var (
	ErrSnakeMovOutLimit  = errors.New("mov out limit")
	ErrSnakeMovTouchSelf = errors.New("touch self")
	ErrSnakeMovGoOppsite = errors.New("go oppsite")
)

func (s *Snake) Move(dir Direction) error {
	if s.dir.Oppsite(dir) {
		return ErrSnakeMovGoOppsite
	}

	s.dir = dir

	var newHead = *(s.head)
	switch dir {
	case DirUp:
		newHead.pos.y -= 1
	case DirRight:
		newHead.pos.x += 1
	case DirDown:
		newHead.pos.y += 1
	case DirLeft:
		newHead.pos.x -= 1
	}

	if _, ok := s.takes[newHead.pos]; ok {
		return ErrSnakeMovTouchSelf
	}

	if newHead.pos.x < s.limit.MinX || newHead.pos.x > s.limit.MaxX ||
		newHead.pos.y < s.limit.MinY || newHead.pos.y > s.limit.MaxY {
		return ErrSnakeMovOutLimit
	}

	delete(s.takes, s.tail.pos)
	defer func() {
		s.takes[s.head.pos] = struct{}{}
	}()

	newHead.next = s.head
	s.head.prev = &newHead
	s.head = &newHead
	s.prevTail = s.tail
	s.tail = s.tail.prev
	s.tail.next = nil

	return nil
}

func (s *Snake) Grow() {
	defer func() {
		s.takes[s.tail.pos] = struct{}{}
	}()

	s.tail.next = s.prevTail
	s.tail = s.prevTail
	s.length += 1
}

func (s *Snake) GetSymbolAt(pos Position) string {
	if _, ok := s.takes[pos]; ok {
		return s.symbol
	}
	return ""
}
