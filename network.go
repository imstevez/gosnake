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

func (nw *Network) GetLocalAddr() string {
	return nw.conn.LocalAddr().String()
}

func (nw *Network) Start(localAddr, remoteAddr string) error {
	var (
		err          error
		laddr, raddr *net.UDPAddr
	)

	if localAddr != "" {
		laddr, err = net.ResolveUDPAddr("udp", localAddr)
		if err != nil {
			return err
		}
	}
	if remoteAddr != "" {
		raddr, err = net.ResolveUDPAddr("udp", remoteAddr)
		if err != nil {
			return err
		}
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
			buf := make([]byte, PackagePayloadSize)
			n, _ := nw.conn.Read(buf)
			data := ReceiveData(buf[:n])
			if data != nil {
				nw.Recv <- data
			}
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
