package gosnake

import (
	"bytes"
	"context"
	"encoding/gob"
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
	GroundWith         int            `json:"ground_width"`
	GroundHeight       int            `json:"ground_height"`
	GroundSymbol       string         `json:"ground_symbol"`
	BorderWidth        int            `json:"border_width"`
	BorderHeight       int            `json:"border_height"`
	BorderSymbol       string         `json:"border_symbol"`
	FoodSymbol         string         `json:"food_symbol"`
	AutoMoveIntervalMS int            `json:"auto_move_interval_ms"`
	PlayerSize         int            `json:"player_size"`
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
		case <-ctx.Done():
			return
		case data := <-room.dataChan:
			func() {
				room.mu.Lock()
				defer room.mu.Unlock()
				uid := data.Sender.String()
				player := room.players[uid]
				if player == nil {
					if len(room.players) < room.options.PlayerSize {
						player, err := NewPlayer(data.Sender, *room.options.PlayerOptions)
						if err != nil {
							return
						}
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
					room.playerMove(uid, DirUp, false)
				case CMDMovDown:
					room.playerMove(uid, DirDown, false)
				case CMDMovLeft:
					room.playerMove(uid, DirLeft, false)
				case CMDMovRight:
					room.playerMove(uid, DirRight, false)
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

func (room *GameRoom) encode() []byte {
	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(room)
	return buf.Bytes()
}

func (room *GameRoom) sendAllPlayersData() {
	data := room.encode()
	for _, player := range room.players {
		player.ReceiveData(data)
	}
}

func (room *GameRoom) playerMove(playerID string, dir Direction, oeated bool) (ieated bool) {
	player := room.players[playerID]
	if player.over {
		return
	}
	player.pause = false
	nextHeadPos := player.snake.GetNextHeadPos(dir)
	if nextHeadPos == nil {
		return
	}
	if room.border.IsTaken(*nextHeadPos) {
		player.over = true
		return
	}
	if player.snake.IsTaken(*nextHeadPos) && *nextHeadPos != player.snake.GetTailPos() {
		player.over = true
		return
	}
	for uid, oplayer := range room.players {
		if uid == playerID {
			continue
		}
		if oplayer.snake.IsTaken(*nextHeadPos) {
			player.over = true
			return
		}
	}
	player.snake.Move(dir)
	if !oeated && room.food.IsTaken(*nextHeadPos) {
		player.snake.Grow()
		room.food.UpdatePos()
		ieated = true
	}
	return
}

func (room *GameRoom) playerPause(playerID string) {
	player := room.players[playerID]
	if !player.over {
		player.pause = true
	}
}

func (room *GameRoom) playerReplay(playerID string) {
	player := room.players[playerID]
	limit := Limit{
		1,
		room.options.BorderWidth - 2,
		1,
		room.options.BorderWidth - 2,
	}
	player.snake = NewCenterPosSnake(
		limit, room.options.PlayerOptions.SnakeSymbol,
	)
	player.over = false
	player.pause = false
}

func (room *GameRoom) allPlayersMove() {
	eated := false
	for pid, player := range room.players {
		eated = room.playerMove(pid, player.snake.GetDir(), eated)
	}
	return
}

type RoomData struct {
	Sender *net.UDPAddr
	Data   []byte
}

func (room *GameRoom) HandleData(data *RoomData) {
	room.dataChan <- data
}

type PlayerOptions struct {
	SnakeSymbol string
	SnakeLimit  Limit
	DefaultName string
}

type Player struct {
	options    *PlayerOptions
	name       string
	snake      *Snake
	over       bool
	pause      bool
	conn       *net.UDPConn
	lastRecv   time.Time
	recv       chan []byte
	addr       *net.UDPAddr
	clearFuncs []func()
}

func NewPlayer(addr *net.UDPAddr, options PlayerOptions) (player *Player, err error) {
	player = &Player{
		options: &options,
		addr:    addr,
	}
	player.name = player.options.DefaultName
	player.snake = NewCenterPosSnake(
		player.options.SnakeLimit, player.options.SnakeSymbol,
	)
	player.lastRecv = time.Now()
	player.recv = make(chan []byte, 1)
	player.conn, err = net.DialUDP("udp", nil, player.addr)
	if err != nil {
		return
	}
	player.clearFuncs = append(
		player.clearFuncs, func() {
			player.conn.Close()
		},
	)
	return
}

func (player *Player) UpdateLastRecv() {
	player.lastRecv = time.Now()
}

func (player *Player) ReceiveData(data []byte) {
	player.recv <- data
}

func (player *Player) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-player.recv:
			player.conn.Write(data[:])
		}
	}
}
