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

func (s *Snake) Len() int {
	return s.length
}

func (s *Snake) GetHeadPos() Position {
	return s.head.pos
}

func (s *Snake) GetTailPos() Position {
	return s.tail.pos
}

func (s *Snake) GetNextHeadPos(dir Direction) *Position {
	if s.dir.Oppsite(dir) {
		return nil
	}

	pos := s.GetHeadPos()

	switch dir {
	case DirUp:
		pos.y -= 1
	case DirRight:
		pos.x += 1
	case DirDown:
		pos.y += 1
	case DirLeft:
		pos.x -= 1
	}

	return &pos
}

func (s *Snake) Move(dir Direction) {
	nextPos := s.GetNextHeadPos(dir)
	if nextPos == nil {
		return
	}
	newHead := &Node{
		next: s.head,
		prev: nil,
		pos:  *nextPos,
	}

	s.head.prev = newHead
	s.head = newHead
	s.prevTail = s.tail
	s.tail = s.tail.prev
	s.tail.next = nil

	delete(s.takes, s.tail.pos)
	s.takes[s.head.pos] = struct{}{}
	s.dir = dir
}

func (s *Snake) Grow() {
	defer func() {
		s.takes[s.tail.pos] = struct{}{}
	}()

	s.tail.next = s.prevTail
	s.tail = s.prevTail
	s.length += 1
}

func (s *Snake) IsTaken(pos Position) bool {
	_, ok := s.takes[pos]
	return ok
}

func (s *Snake) GetSymbolAt(pos Position) string {
	return IfStr(
		s.IsTaken(pos), s.symbol, "",
	)
}
