package gosnake

type node struct {
	next *node
	prev *node
	pos  Position
}

type Snake struct {
	head   *node
	tail   *node
	grow   *node
	dir    Direction
	length int
}

func NewSnake(initialPos Position, initialDir Direction) *Snake {
	node := &node{
		next: nil,
		prev: nil,
		pos:  initialPos,
	}
	return &Snake{
		head:   node,
		tail:   node,
		grow:   nil,
		length: 1,
		dir:    initialDir,
	}
}

func (s *Snake) Dir() Direction {
	return s.dir
}

func (s *Snake) Len() int {
	return s.length
}

func (s *Snake) HeadPos() Position {
	return s.head.pos
}

func (s *Snake) TailPos() Position {
	return s.tail.pos
}

func (s *Snake) Move(dir Direction) {
	if s.dir.Oppsite(dir) {
		return
	}

	s.dir = dir

	newHeadPos := Position{
		X: s.head.pos.X,
		Y: s.head.pos.Y,
	}

	switch dir {
	case DirUp:
		newHeadPos.Y -= 1
	case DirDown:
		newHeadPos.Y += 1
	case DirRight:
		newHeadPos.X += 1
	case DirLeft:
		newHeadPos.X -= 1
	}

	newHead := &node{
		next: s.head,
		prev: nil,
		pos:  newHeadPos,
	}

	s.head.prev = newHead
	s.head = newHead
	s.grow = s.tail
	s.tail = s.tail.prev
	s.tail.next = nil
}

func (s *Snake) Grow() {
	if s.grow == nil {
		return
	}
	s.tail.next = s.grow
	s.tail = s.tail.next
	s.length += 1
	s.grow = nil
}
