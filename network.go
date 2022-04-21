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

func (nw *Network) Start(localIP string, localPort int, dialIP string, dialPort int) error {
	localAddr := &net.UDPAddr{
		IP:   net.ParseIP(localIP),
		Port: localPort,
	}
	dialAddr := &net.UDPAddr{
		IP:   net.ParseIP(dialIP),
		Port: dialPort,
	}
	var err error
	nw.conn, err = net.DialUDP("udp", localAddr, dialAddr)
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
