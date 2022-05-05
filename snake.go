package gosnake

import "gosnake/base"

type node struct {
	next *node
	prev *node
	pos  base.Position2D
}

type Snake struct {
	head   *node
	tail   *node
	grow   *node
	dir    base.Direction2D
	length int
	bitmap *base.Bitmap2D
}

func NewSnake(initPos base.Position2D, initDir base.Direction2D) *Snake {
	node := &node{
		next: nil,
		prev: nil,
		pos:  initPos,
	}
	bitmap := &base.Bitmap2D{}
	bitmap.Set(initPos, true)
	return &Snake{
		head:   node,
		tail:   node,
		grow:   nil,
		length: 1,
		dir:    initDir,
		bitmap: bitmap,
	}
}

func (s *Snake) Dir() base.Direction2D {
	return s.dir
}

func (s *Snake) Len() int {
	return s.length
}

func (s *Snake) HeadPos() base.Position2D {
	return s.head.pos
}

func (s *Snake) TailPos() base.Position2D {
	return s.tail.pos
}

func (s *Snake) GetBitmap() *base.Bitmap2D {
	return s.bitmap
}

func (s *Snake) GetNextHeadPos(dir base.Direction2D) *base.Position2D {
	if s.dir.OppsiteTo(dir) {
		return nil
	}
	nextHeadPos := s.HeadPos()
	switch dir {
	case base.Dir2DUp:
		nextHeadPos.Y--
	case base.Dir2DRight:
		nextHeadPos.X++
	case base.Dir2DDown:
		nextHeadPos.Y++
	case base.Dir2DLeft:
		nextHeadPos.X--
	}
	return &nextHeadPos
}

func (s *Snake) Move(dir base.Direction2D) {
	nextHeadPos := s.GetNextHeadPos(dir)
	if nextHeadPos == nil {
		return
	}

	s.dir = dir

	newHead := &node{
		next: s.head,
		prev: nil,
		pos:  *nextHeadPos,
	}

	s.head.prev = newHead
	s.head = newHead
	s.grow = s.tail
	s.bitmap.Set(s.tail.pos, false)
	s.tail = s.tail.prev
	s.tail.next = nil
	s.bitmap.Set(s.head.pos, true)
}

func (s *Snake) Grow() {
	if s.grow == nil {
		return
	}
	s.tail.next = s.grow
	s.tail = s.tail.next
	s.length += 1
	s.grow = nil
	s.bitmap.Set(s.tail.pos, true)
}
