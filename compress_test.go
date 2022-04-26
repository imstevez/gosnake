package gosnake

import (
	"fmt"
	"testing"
	"unsafe"
)

var set = map[Position]struct{}{
	{X: 0, Y: 0}: {},
	{X: 1, Y: 0}: {},
	{X: 2, Y: 0}: {},
	{X: 3, Y: 0}: {},
	{X: 3, Y: 1}: {},
	{X: 3, Y: 2}: {},
	{X: 3, Y: 3}: {},
}

var unset = map[Position]struct{}{
	{X: 0, Y: 1}: {},
	{X: 1, Y: 1}: {},
	{X: 2, Y: 1}: {},
	{X: 0, Y: 2}: {},
	{X: 0, Y: 3}: {},
	{X: 1, Y: 2}: {},
	{X: 1, Y: 3}: {},
}

func TestCompressLayer(t *testing.T) {
	fmt.Println(unsafe.Sizeof(ChildPackage{}))
	layer := NewCompressLayer(4, 4)
	layer.AddPositions(set)
	for pos := range set {
		if !layer.IsTaken(pos) {
			t.Errorf("set pos %v is not be taken", pos)
		}
	}
	for pos := range unset {
		if layer.IsTaken(pos) {
			t.Errorf("unset pos %v is be taken", pos)
		}
	}
}
