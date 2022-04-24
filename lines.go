package gosnake

import (
	"fmt"
	"strings"
)

type Lines []string

func (ls Lines) Sprintlines(args ...interface{}) (ols Lines) {
	ols = make([]string, len(ls))
	argsi := 0
	for i, tpl := range ls {
		nargs := strings.Count(tpl, "%")
		nargs -= strings.Count(tpl, "%%") * 2
		argsj := argsi + nargs
		if argsj > len(args) {
			panic("args is too short")
		}
		ols[i] = fmt.Sprintf(
			tpl, args[argsi:argsj]...,
		)
		argsi += nargs
	}
	return ols
}

func (ls Lines) HozJoin(rls Lines, leftWidth int) (ols Lines) {
	n1, n2 := len(ls), len(rls)
	n := IfInt(n1 > n2, n1, n2)
	ols = make([]string, n)
	for i := 0; i < n; i++ {
		var l1, l2 string
		if i < n1 {
			l1 = ls[i]
		}
		if i < n2 {
			l2 = rls[i]
		}
		ols[i] = fmt.Sprintf(
			"%s\r\033[%dC%s", l1, leftWidth, l2,
		)
	}
	return
}

func (ls Lines) Merge() (str string) {
	str = fmt.Sprintf("\033[%dA", len(ls))
	for _, l := range ls {
		str += "\r" + l + "\033[K\n"
	}
	return
}
