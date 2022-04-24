package gosnake

import (
	"bytes"
	"context"
	"encoding/gob"
	"net"
	"sync"
)

var DefaultServerOptions = &ServerOptions{
	Addr:     "127.0.0.1:9001",
	RoomSize: 5,
	BufSize:  40960,
	GameRoomOptions: &GameRoomOptions{
		GroundWith:         30,
		GroundHeight:       30,
		GroundSymbol:       "  ",
		BorderWidth:        30,
		BorderHeight:       30,
		BorderSymbol:       "\033[46;1;37m[]\033[0m",
		FoodSymbol:         "\033[42;1;37m[]\033[0m",
		AutoMoveIntervalMS: 400,
		PlayerSize:         5,
		PlayerOptions: &PlayerOptions{
			SnakeSymbol: "\033[41;1;37m[]\033[0m",
			SnakeLimit: Limit{
				1, 28, 1, 28,
			},
		},
	},
}

func RunServer(ctx context.Context) error {
	server := NewServer(DefaultServerOptions)
	return server.Run(ctx)
}

type ServerOptions struct {
	Addr            string
	RoomSize        int
	BufSize         int
	GameRoomOptions *GameRoomOptions
}
type Server struct {
	options ServerOptions
	send    chan *ServerData
	rooms   []*GameRoom
}

type ServerData struct {
	Addr *net.UDPAddr
	Data []byte
}

func NewServer(options *ServerOptions) *Server {
	return &Server{
		options: *options,
		send:    make(chan *ServerData, 1),
	}
}

// Run run a sever
func (s *Server) Run(ctx context.Context) error {
	// create and run rooms
	s.rooms = make([]*GameRoom, s.options.RoomSize)
	wg := &sync.WaitGroup{}
	for i := 0; i < len(s.rooms); i++ {
		room := NewGameRoom(s.send, s.options.GameRoomOptions)
		s.rooms[i] = room
		wg.Add(1)
		go func() {
			room.Run(ctx)
			wg.Done()
		}()
	}
	defer wg.Wait()

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

	recv := make(chan *RoomData, 1)
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			buf := make([]byte, s.options.BufSize)
			n, sender, err := conn.ReadFromUDP(buf)
			if err != nil || sender == nil || n <= 0 {
				continue
			}
			cliData, err := s.decodeClientData(buf[:])
			if err != nil {
				continue
			}
			recv <- &RoomData{
				Sender:     sender,
				ClientData: cliData,
			}
		}
	}(ctx)

	// loop receive data
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case data := <-s.send:
			conn.WriteToUDP(data.Data, data.Addr)
		case data := <-recv:
			if data.ClientData.RoomID >= len(s.rooms) {
				continue
			}
			room := s.rooms[data.ClientData.RoomID]
			room.HandleData(data)
		}
	}
}

type ClientData struct {
	RoomID int
	CMD    string
}

func (s *Server) decodeClientData(data []byte) (clientData *ClientData, err error) {
	clientData = new(ClientData)
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err = decoder.Decode(clientData)
	return
}
