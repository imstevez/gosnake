package gosnake

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"time"
)

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

type RoomFactory func(int) *GameRoom

// Run run a sever
func (s *Server) Run(ctx context.Context) error {
	// create and run rooms
	s.rooms = make([]*GameRoom, s.options.RoomSize)
	wg := &sync.WaitGroup{}
	for i := 0; i < len(s.rooms); i++ {
		room := NewGameRoom(i, s.options.GameRoomOptions)
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

	// loop receive data
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			buf := make([]byte, s.options.BufSize)
			n, sender, err := conn.ReadFromUDP(buf)
			if err == nil && sender != nil {
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
	room.HandleData(&RoomData{
		Sender: sender,
		Data:   svrData.Data,
	})
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

type GameRoomOptions struct {
	GroundWith         int    `json:"ground_width"`
	GroundHeight       int    `json:"ground_height"`
	GroundSymbol       string `json:"ground_symbol"`
	BorderWidth        int    `json:"border_width"`
	BorderHeight       int    `json:"border_height"`
	BorderSymbol       string `json:"border_symbol"`
	FoodSymbol         string `json:"food_symbol"`
	AutoMoveIntervalMS int    `json:"auto_move_interval_ms"`
	PlayerSize         int
	PlayerOptions      *PlayerOptions `json:"player_options`
}

type GameRoom struct {
	options    GameRoomOptions
	id         int
	players    map[string]*Player
	ground     *Ground
	border     *RecBorder
	food       *Food
	autoticker *time.Ticker
	dataChan   chan *RoomData
	mu         sync.Mutex
}

func NewGameRoom(roomID int, options *GameRoomOptions) *GameRoom {
	return &GameRoom{
		id:      roomID,
		options: *options,
	}
}

func (room *GameRoom) Init() {
	// new ground
	room.ground = NewGround(
		room.options.GroundWith, room.options.GroundHeight,
		room.options.GroundSymbol,
	)

	// new border
	room.border = NewRecBorder(
		room.options.BorderWidth, room.options.BorderHeight,
		room.options.BorderSymbol,
	)

	// new food
	room.food = NewFood(
		room.options.FoodSymbol, Limit{
			MinX: 1, MaxX: room.options.BorderWidth - 2,
			MinY: 1, MaxY: room.options.BorderHeight - 2,
		},
	)

	// create auto move ticker
	room.autoticker = time.NewTicker(
		time.Duration(room.options.AutoMoveIntervalMS) * time.Millisecond,
	)

	// make room data channel
	room.dataChan = make(chan *RoomData, 1)
}

func (room *GameRoom) Run(ctx context.Context) {
	room.Init()
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	for {
		select {
		case data := <-room.dataChan:
			func() {
				room.mu.Lock()
				defer room.mu.Unlock()
				uid := data.Sender.String()
				player := room.players[uid]
				if player == nil {
					if len(room.players) < room.options.PlayerSize {
						player = NewPlayer(*room.options.PlayerOptions)
						room.players[uid] = player
						wg.Add(1)
						go func() {
							player.Run(ctx)
							wg.Done()
						}()
					}
					return
				}
				player.UpdateLastRecv()
				switch string(data.Data) {
				case CMDMovUp:
					room.playerMove(uid, DirUp)
				case CMDMovDown:
					room.playerMove(uid, DirDown)
				case CMDMovLeft:
					room.playerMove(uid, DirLeft)
				case CMDMovRight:
					room.playerMove(uid, DirRight)
				case CMDPause:
					room.playerPause(uid)
				case CMDReplay:
					room.playerReplay(uid)
				}
				room.sendAllPlayersData()
			}()

		case <-room.autoticker.C:
			func() {
				room.mu.Lock()
				defer room.mu.Unlock()
				room.allPlayersMove()
				room.sendAllPlayersData()
			}()
		}
	}
	return
}

type RoomSendData struct {
}

func (room *GameRoom) getSendData() *RoomSendData {
	return &RoomSendData{}
}

func (room *GameRoom) sendAllPlayersData() {

}

func (room *GameRoom) playerMove(playerID string, dir Direction) {

}
func (room *GameRoom) playerPause(playerID string) {

}
func (room *GameRoom) playerReplay(playerID string) {

}

func (room *GameRoom) allPlayersMove() {

}

type RoomData struct {
	Sender *net.UDPAddr
	Data   []byte
}

func (room *GameRoom) HandleData(data *RoomData) {
	room.dataChan <- data
}

type PlayerOptions struct {
}

type Player struct {
	name     string
	snake    *Snake
	over     bool
	pause    bool
	conn     net.Conn
	lastRecv time.Time
}

func NewPlayer(options PlayerOptions) *Player {
	return &Player{}
}

func (player *Player) UpdateLastRecv() {
	player.lastRecv = time.Now()
}

func (player *Player) Run(ctx context.Context) {

}
