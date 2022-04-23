package server

import (
	"context"
	"encoding/json"
	"net"
	"sync"
)

type ServerOptions struct {
	Addr     string
	RoomSize int
	BufSize  int
}
type Server struct {
	options ServerOptions
	rooms   []*Room
}

func NewServer(options *ServerOptions) *Server {
	return &Server{
		options: *options,
	}
}

type RoomFactory func(int) *Room

// Run run a sever
func (s *Server) Run(ctx context.Context, newRoom RoomFactory) error {
	// create and run rooms
	s.rooms = make([]*Room, s.options.RoomSize)
	wg := &sync.WaitGroup{}
	for i := 0; i < len(s.rooms); i++ {
		room := newRoom(i)
		s.rooms[i] = newRoom(i)
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

	// loop receive data
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			buf := make([]byte, s.options.BufSize)
			n, sender, err := conn.ReadFromUDP(buf)
			if err != nil {
				s.handleData(sender, buf[:n])
			}
		}
	}
}

// handleData decode received data, and send the wrapped content to the correspond room
func (s *Server) handleData(sender *net.UDPAddr, data []byte) {
	svrData, err := s.decodeData(data)
	if err != nil {
		return
	}
	if svrData.RoomID >= len(s.rooms) {
		return
	}
	room := s.rooms[svrData.RoomID]
	room.HandleData(sender, svrData.Data)
}

type ServerData struct {
	RoomID int    `json:"room_id"`
	Data   []byte `json:"data"`
}

func (s *Server) decodeData(data []byte) (svrData *ServerData, err error) {
	svrData = new(ServerData)
	err = json.Unmarshal(data, svrData)
	return
}

type Room struct {
}

func NewRoom() *Room {
	return nil
}

func (room *Room) Run(ctx context.Context) {
	return
}

func (room *Room) HandleData(sender *net.UDPAddr, data []byte) {

}
