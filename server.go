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
			SendData(data.Data, conn, data.Addr)
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

var SendNum uint64

const PackagePayloadSize = 480

type Package struct {
	ID     uint64
	Number int
	Total  int
	Data   []byte
}

func EncodePackage(pac *Package) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	enc.Encode(pac)
	return buf.Bytes()
}

func DecodePackage(data []byte) *Package {
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	pac := &Package{}
	decoder.Decode(pac)
	return pac
}

func SendData(data []byte, conn *net.UDPConn, addr *net.UDPAddr) {
	num := len(data) / PackagePayloadSize
	if len(data)%PackagePayloadSize != 0 {
		num += 1
	}
	SendNum++
	for i := 0; i < num; i++ {
		s := i * PackagePayloadSize
		e := s + PackagePayloadSize
		if e > len(data) {
			e = len(data)
		}
		data := EncodePackage(&Package{
			ID:     SendNum,
			Number: i,
			Total:  num,
			Data:   data[s:e],
		})
		if len(data) == 0 {
			break
		}
		conn.WriteToUDP(data, addr)
	}
}

var (
	PackagesBuf         = make([][]byte, 10)
	CurrentPackageID    uint64
	CurrentPackageTotal int
	ReceivedNums        int
)

func ReceiveData(data []byte) []byte {
	pac := DecodePackage(data)
	if pac == nil {
		return nil
	}
	if pac.ID != CurrentPackageID {
		CurrentPackageID = pac.ID
		CurrentPackageTotal = pac.Total
		ReceivedNums = 0
	}
	PackagesBuf[pac.Number] = pac.Data
	ReceivedNums += 1
	if ReceivedNums == CurrentPackageTotal {
		return bytes.Join(PackagesBuf, []byte{})
	}
	return nil
}
