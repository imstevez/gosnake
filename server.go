package gosnake

import (
	"context"
	"net"
	"sync"
	"time"
)

const (
	splitChildPackageSize = 1350
	splitChildPackageNum  = 10
	serverReadBufferSize  = 512
)

var DefaultServerOptions = &ServerOptions{
	Addr: "127.0.0.1:9001",
	GameOptions: &GameOptions{
		GroundWidth:         32,
		GroundHeight:        32,
		AutoMoveInterval:    300 * time.Millisecond,
		ClearPlayerInterval: 10 * time.Second,
		PlayerSize:          5,
	},
}

func RunServer(ctx context.Context) error {
	server := NewServer(DefaultServerOptions)
	return server.Run(ctx)
}

type ServerOptions struct {
	Addr        string
	GameOptions *GameOptions
}

type Server struct {
	options         ServerOptions
	splitDataSender *SplitDataSender
}

func NewServer(options *ServerOptions) *Server {
	return &Server{
		options: *options,
		splitDataSender: NewSplitDataSender(
			splitChildPackageSize,
		),
	}
}

func (s *Server) Run(ctx context.Context) error {
	listenAddr, err := net.ResolveUDPAddr("udp", s.options.Addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	inputs := make(chan *GameInput, 1)
	game := NewGame(s.options.GameOptions)
	outputs := game.Start(inputs)

	go func() {
		defer close(inputs)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				buf := make([]byte, serverReadBufferSize)
				n, sender, err := conn.ReadFromUDP(buf)
				if err != nil || sender == nil || n <= 0 {
					continue
				}
				inputs <- &GameInput{
					From: sender,
					Data: buf[:n],
				}
			}
		}
	}()

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	threads := make(chan struct{}, 5)
	for output := range outputs {
		threads <- struct{}{}
		wg.Add(1)
		go func(output *GameOutput) {
			defer func() {
				<-threads
				wg.Done()
			}()
			s.splitDataSender.SendDataWithUDP(
				output.Data, conn, output.To,
			)
		}(output)
	}

	return nil
}
