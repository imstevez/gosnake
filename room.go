package gosnake

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net"
	"sync"
	"time"
)

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
	players    map[string]*Player
	ground     *Ground
	border     *RecBorder
	food       *Food
	autoticker *time.Ticker
	dataChan   chan *RoomData
	serverSend chan<- *ServerData
	mu         sync.Mutex
}

func NewGameRoom(serverSend chan<- *ServerData, options *GameRoomOptions) *GameRoom {
	return &GameRoom{
		options:    *options,
		serverSend: serverSend,
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

	// make room players map
	room.players = make(map[string]*Player, room.options.PlayerSize)
}

func (room *GameRoom) Run(ctx context.Context) {
	room.Init()
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
						player, err := NewPlayer(data.Sender, room.options.PlayerOptions)
						if err != nil {
							return
						}
						room.players[uid] = player
					}
					return
				}
				player.UpdateLastRecv()
				fmt.Println(uid, data.ClientData.RoomID, data.ClientData.CMD)
				switch data.ClientData.CMD {
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
				case CMDQuit:
					delete(room.players, uid)
				}
				room.sendAllPlayersData()
			}()

		case <-room.autoticker.C:
			func() {
				room.mu.Lock()
				defer room.mu.Unlock()
				room.playersAutoMove()
				room.sendAllPlayersData()
			}()
		}
	}
}

type GameData struct {
	Options    GameRoomOptions
	Players    PlayerList
	FoodPos    Position
	FoodSymbol string
}
type PlayerList []*Player

func (pl PlayerList) Len() int { return len(pl) }
func (pl PlayerList) Less(i, j int) bool {
	if pl[i].Score == pl[j].Score {
		return pl[i].CreatedAt.Before(pl[j].CreatedAt)
	}
	return pl[i].Score > pl[j].Score
}
func (pl PlayerList) Swap(i, j int) { pl[i], pl[j] = pl[j], pl[i] }

func (room *GameRoom) encode() []byte {
	data := GameData{
		Options: room.options,
		FoodPos: room.food.pos,
		Players: make(PlayerList, len(room.players)),
	}
	i := 0
	for _, player := range room.players {
		data.Players[i] = player
		i++
	}
	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(&data)
	return buf.Bytes()
}

func (room *GameRoom) sendAllPlayersData() {
	data := room.encode()
	for _, player := range room.players {
		room.serverSend <- &ServerData{
			Addr: player.Addr,
			Data: data,
		}
	}
}

func (room *GameRoom) playerMove(playerID string, dir Direction, oeated bool) (ieated bool) {
	player := room.players[playerID]
	player.Pause = false
	nextHeadPos := player.snake.GetNextHeadPos(dir)
	if nextHeadPos == nil {
		return
	}
	if room.border.IsTaken(*nextHeadPos) {
		player.Over = true
		return
	}
	if player.snake.IsTaken(*nextHeadPos) && *nextHeadPos != player.snake.GetTailPos() {
		player.Over = true
		return
	}
	for uid, oplayer := range room.players {
		if uid == playerID {
			continue
		}
		if oplayer.snake.IsTaken(*nextHeadPos) {
			player.Over = true
			return
		}
	}
	player.snake.Move(dir)
	player.SnakeTakes = player.snake.GetTakes()
	if !oeated && room.food.IsTaken(*nextHeadPos) {
		player.snake.Grow()
		room.food.UpdatePos()
		ieated = true
	}
	player.Score = player.snake.Len() - 1
	return
}

func (room *GameRoom) playerPause(playerID string) {
	player := room.players[playerID]
	if !player.Over {
		player.Pause = true
	}
}

func (room *GameRoom) playerReplay(playerID string) {
	player := room.players[playerID]
	player.Reset()
}

func (room *GameRoom) playersAutoMove() {
	eated := false
	for pid, player := range room.players {
		if !player.Pause {
			eated = room.playerMove(pid, player.snake.GetDir(), eated)
		}
	}
}

type RoomData struct {
	Sender     *net.UDPAddr
	ClientData *ClientData
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
	Name       string
	SnakeTakes map[Position]struct{}
	Over       bool
	Pause      bool
	Score      int
	snake      *Snake
	LastRecv   time.Time
	CreatedAt  time.Time
	Addr       *net.UDPAddr
}

func NewPlayer(addr *net.UDPAddr, options *PlayerOptions) (player *Player, err error) {
	player = &Player{
		options: options,
		Addr:    addr,
	}
	player.Name = addr.String()
	player.snake = NewCenterPosSnake(
		options.SnakeLimit, options.SnakeSymbol,
	)
	player.LastRecv = time.Now()
	player.CreatedAt = time.Now()
	return
}

func (player *Player) UpdateLastRecv() {
	player.LastRecv = time.Now()
}

func (player *Player) Reset() {
	player.snake = NewCenterPosSnake(
		player.options.SnakeLimit, player.options.SnakeSymbol,
	)
	player.Pause = false
	player.Over = false
}
