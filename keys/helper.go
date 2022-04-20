package keys

import "syscall"

func getCode() Code {
	var buf [1]byte
	if n, err := syscall.Read(0, buf[:]); n == 0 || err != nil {
		panic(err)
	}
	return Code(buf[0])
}
