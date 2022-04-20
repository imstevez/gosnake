package gosnake

type Node struct {
	next *Node
	prev *Node
	pos  Position
}

type Snake struct {
	head     *Node
	tail     *Node
	prevTail *Node
	length   int
	dir      Direction
	takes    map[Position]struct{}
	symbol   string
}

func NewSnake(initialPosX, initialPosY int, initialDir Direction, symbol string) *Snake {
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

func (s *Snake) Move(dir Direction) {
	if s.dir.RevertTo(dir) {
		return
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
}

func (s *Snake) Grow() {
	defer func() {
		s.takes[s.tail.pos] = struct{}{}
	}()

	s.tail.next = s.prevTail
	s.tail = s.prevTail
	s.length += 1
}

func (s *Snake) IsTouchSelf() bool {
	return len(s.takes) < s.length
}

func (s *Snake) IsTaken(pos Position) bool {
	_, ok := s.takes[pos]
	return ok
}

func (s *Snake) GetSymbol() string {
	return s.symbol
}
