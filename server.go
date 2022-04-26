package gosnake

import (
	"bytes"
	"context"
	"encoding/gob"
	"net"
	"sync"
)

const (
	splitChildPackageSize = 512
	splitChildPackageNum  = 10
	serverReadBufferSize  = 512
)

var DefaultServerOptions = &ServerOptions{
	Addr:     "127.0.0.1:9001",
	RoomSize: 5,
	RoomOptions: &RoomOptions{
		BorderWidth:        32,
		BorderHeight:       32,
		AutoMoveIntervalMS: 300,
		PlayerSize:         5,
	},
}

func RunServer(ctx context.Context) error {
	server := NewServer(DefaultServerOptions)
	return server.Run(ctx)
}

type ServerOptions struct {
	Addr        string
	RoomSize    int
	RoomOptions *RoomOptions
}
type Server struct {
	options         ServerOptions
	rooms           []*Room
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

// Run run a sever
func (s *Server) Run(ctx context.Context) error {
	// resolve listen addr
	listenAddr, err := net.ResolveUDPAddr("udp", s.options.Addr)
	if err != nil {
		return err
	}

	// listen UDP at the specified addr
	conn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	sendData := func(data []byte, addr *net.UDPAddr) {
		s.splitDataSender.SendDataWithUDP(data, conn, addr)
	}

	// create and run rooms
	s.rooms = make([]*Room, s.options.RoomSize)
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	for i := 0; i < len(s.rooms); i++ {
		room := NewRoom(s.options.RoomOptions, sendData)
		s.rooms[i] = room
		wg.Add(1)
		go func(room *Room) {
			room.Run(ctx)
			wg.Done()
		}(room)
	}

	// Recieve
	buf := make([]byte, serverReadBufferSize)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			n, sender, err := conn.ReadFromUDP(buf)
			if err != nil || sender == nil || n <= 0 {
				continue
			}
			cliData, err := s.decodeClientData(buf[:])
			if err != nil || cliData.RoomID > len(s.rooms) {
				continue
			}
			room := s.rooms[cliData.RoomID]
			room.HandleData(&RoomData{
				Sender:     sender,
				ClientData: cliData,
			})
		}
	}
}

type ClientData struct {
	RoomID int
	CMD    CMD
}

func (s *Server) decodeClientData(data []byte) (clientData *ClientData, err error) {
	clientData = new(ClientData)
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err = decoder.Decode(clientData)
	return
}
