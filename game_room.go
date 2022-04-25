package gosnake

import (
	"context"
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
	PlayerOptions      *PlayerOptions `json:"player_options"`
}

type GameRoom struct {
	options    GameRoomOptions
	players    map[string]*Player
	ground     *Ground
	border     *RecBorder
	food       *Food
	autoticker *time.Ticker
	dataChan   chan *RoomData
	mu         sync.Mutex
	conn       *net.UDPConn
}

func NewGameRoom(conn *net.UDPConn, options *GameRoomOptions) *GameRoom {
	return &GameRoom{
		options: *options,
		conn:    conn,
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
	room.players = make(
		map[string]*Player, room.options.PlayerSize,
	)
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
				playerID := data.Sender.String()
				player := room.players[playerID]
				if player == nil {
					if len(room.players) < room.options.PlayerSize {
						player, err := NewPlayer(
							room.options.PlayerOptions,
							data.Sender,
							playerID,
						)
						if err != nil {
							return
						}
						room.players[player.ID] = player
					}
					return
				}
				player.UpdateLastRecv()
				fmt.Printf("[R] %s %s\n", player.ID, data.ClientData.CMD)
				switch data.ClientData.CMD {
				case CMDMovUp:
					room.playerMove(playerID, DirUp, false)
				case CMDMovDown:
					room.playerMove(playerID, DirDown, false)
				case CMDMovLeft:
					room.playerMove(playerID, DirLeft, false)
				case CMDMovRight:
					room.playerMove(playerID, DirRight, false)
				case CMDPause:
					room.playerPause(playerID)
				case CMDReplay:
					room.playerReplay(playerID)
				case CMDQuit:
					delete(room.players, playerID)
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

func (room *GameRoom) getSceneData() *GameSceneData {
	data := &GameSceneData{
		Options: room.options,
		FoodPos: room.food.pos,
		Players: make(PlayerList, len(room.players)),
	}
	i := 0
	for _, player := range room.players {
		data.Players[i] = player
		i++
	}
	return data
}

func (room *GameRoom) sendAllPlayersData() {
	sceneData := room.getSceneData()
	wg := &sync.WaitGroup{}
	for _, player := range room.players {
		data := sceneData.EncodeForPlayer(player.ID)
		wg.Add(1)
		go func(data []byte, player *Player) {
			defer wg.Done()
			SendData(
				data, room.conn,
				player.Addr,
			)
		}(data, player)
	}
}

func (room *GameRoom) playerMove(playerID string, dir Direction, oeated bool) (ieated bool) {
	player := room.players[playerID]
	if player.Over {
		return
	}
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
	if player.Over {
		return
	}
	player.Pause = true
}

func (room *GameRoom) playerReplay(playerID string) {
	player := room.players[playerID]
	if !player.Over {
		return
	}
	player.Reset()
}

func (room *GameRoom) playersAutoMove() {
	eated := false
	for pid, player := range room.players {
		if player.Pause {
			continue
		}
		eated = room.playerMove(
			pid, player.snake.GetDir(), eated,
		)
	}
}

type RoomData struct {
	Sender     *net.UDPAddr
	ClientData *ClientData
}

func (room *GameRoom) HandleData(data *RoomData) {
	room.dataChan <- data
}
