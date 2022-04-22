// +build !windows

package keys

import "syscall"

func getch() byte {
	var buf [1]byte
	n, err := syscall.Read(int(fd), buf[:])
	if n == 0 || err != nil {
		panic(err)
	}
	return buf[0]
}
