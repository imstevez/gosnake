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
	BufSize:  512,
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
	rooms   []*GameRoom
}

func NewServer(options *ServerOptions) *Server {
	return &Server{
		options: *options,
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

	// create and run rooms
	s.rooms = make([]*GameRoom, s.options.RoomSize)
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	for i := 0; i < len(s.rooms); i++ {
		room := NewGameRoom(s.options.GameRoomOptions, conn)
		s.rooms[i] = room
		wg.Add(1)
		go func(room *GameRoom) {
			room.Run(ctx)
			wg.Done()
		}(room)
	}

	// Recieve
	buf := make([]byte, s.options.BufSize)
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
