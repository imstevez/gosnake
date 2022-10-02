package base

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"
)

const (
	statusStopping = 0
	statusRunning  = 1
)

type UDPClient struct {
	remoteAddr *net.UDPAddr
	sender     *PacketSplitSender
	receiver   *PacketSplitReceiver
	status     AtomicStatus
}

type UDPClientDialOptions struct {
	LocalAddr, RemoteAddr        *net.UDPAddr
	SendChildSize, RecvChildSize int
	ReadTimeout                  time.Duration
}

type RecvData struct {
	Data []byte
	Addr *net.UDPAddr
	Err  error
}

func (cli *UDPClient) Dial(ctx context.Context, options *UDPClientDialOptions) (err error, recv <-chan *RecvData) {
	set := cli.status.SetFrom(statusStopping, statusRunning)
	if !set {
		err = errors.New("client is running")
		return
	}
	conn, err := net.DialUDP("udp", options.LocalAddr, options.RemoteAddr)
	if err != nil {
		return
	}
	cli.remoteAddr = options.RemoteAddr
	cli.sender = NewPacketSplitSender(options.SendChildSize, conn)
	cli.receiver = NewPacketSplitReceiver(options.RecvChildSize, conn, options.ReadTimeout)
	ch := make(chan *RecvData, 1)
	recv = ch
	go func(ctx context.Context, recv chan<- *RecvData) {
		for {
			select {
			case <-ctx.Done():
				_ = conn.Close()
				close(recv)
				cli.status.Set(statusStopping)
				return
			default:
				data, addr, err := cli.receiver.Recv(ctx)
				recv <- &RecvData{
					Data: data,
					Addr: addr.(*net.UDPAddr),
					Err:  err,
				}
			}
		}
	}(ctx, ch)
	return
}

func (cli *UDPClient) Send(data []byte) (err error) {
	_, err = cli.sender.Send(data, cli.remoteAddr)
	return
}

type UDPServer struct {
	sender   *PacketSplitSender
	receiver *PacketSplitRecver
	status   AtomicStatus
}

type UDPServerListenOptions struct {
	Addr                         *net.UDPAddr
	SendChildSize, RecvChildSize int
	ReadTimeout                  time.Duration
}

type UDPKind string

const (
	UDPKindDial   = "dial"
	UDPKindListen = "listen"
)

type UDP struct {
	sender *PacketSplitSender
	recver *PacketSplitRecver
	rwlock *sync.RWMutex
	runing bool
}

func (udp *UDP) setSender(conn *net.UDPConn, writeSize int) {
	udp.sender = NewPacketSplitSender(conn, writeSize)
}

func (udp *UDP) setRecver(conn *net.UDPConn, readSize int, readTimeout time.Duration) {
	udp.recver = NewPacketSplitRecver(conn, readSize, readTimeout)
}

func (udp *UDP) startRecv(ctx context.Context) (recv <-chan *RecvData) {
	ch := make(chan *RecvData, 1)
	recv = ch
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(ch)
				udp.status.Set(statusStopping)
				return
			default:
				data, addr, err := udp.recver.Recv(ctx)
				ch <- &RecvData{
					Data: data,
					Addr: addr.(*net.UDPAddr),
					Err:  err,
				}
			}
		}
	}()
	return
}

func (svr *UDPServer) Listen(ctx context.Context, options *UDPServerListenOptions) (err error, recv chan<- *RecvData) {
	set := svr.status.SetFrom(statusStopping, statusRunning)
	if !set {
		err = errors.New("server is running")
		return
	}
	conn, err := net.ListenUDP("udp", options.Addr)
	if err != nil {
		return
	}
	svr.sender = NewPacketSplitSender(options.SendChildSize, conn)
	svr.recver = NewPacketSplitRecver(options.RecvChildSize, conn, options.ReadTimeout)
	ch := make(chan *RecvData, 1)
	recv = ch
	go func(ctx context.Context, recv chan<- *RecvData) {
		for {
			select {
			case <-ctx.Done():
				_ = conn.Close()
				close(recv)
				svr.status.Set(statusStopping)
				return
			default:
				data, addr, err := svr.receiver.Recv(ctx)
				recv <- &RecvData{
					Data: data,
					Addr: addr.(*net.UDPAddr),
					Err:  err,
				}
			}
		}
	}(ctx, ch)
	return
}
