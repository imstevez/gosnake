package gosnake

import (
	"net"
)

type Network struct {
	conn      *net.UDPConn
	Recv      chan []byte
	Send      chan []byte
	clearFunc func()
}

func NewNetWork() *Network {
	return &Network{
		Recv: make(chan []byte),
		Send: make(chan []byte),
	}
}

const bufsize = 40960

func (nw *Network) Start(localAddr, remoteAddr string) error {
	var err error
	laddr, err := net.ResolveUDPAddr("udp", localAddr)
	if err != nil {
		return err
	}
	raddr, err := net.ResolveUDPAddr("udp", remoteAddr)
	if err != nil {
		return err
	}
	nw.conn, err = net.DialUDP("udp", laddr, raddr)
	if err != nil {
		return err
	}

	nw.clearFunc = func() {
		nw.conn.Close()
	}

	go func() {
		for {
			buf := make([]byte, bufsize)
			n, _ := nw.conn.Read(buf)
			nw.Recv <- buf[:n]
		}
	}()

	go func() {
		for {
			buf := <-nw.Send
			nw.conn.Write(buf)
		}
	}()

	return nil
}

func (nw *Network) Stop() {
	if nw.clearFunc != nil {
		nw.clearFunc()
	}
}
