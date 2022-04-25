package gosnake

import (
	"context"
	"errors"
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
	conn       *net.UDPConn
}

func NewGameRoom(options *GameRoomOptions, conn *net.UDPConn) *GameRoom {
	return &GameRoom{
		options: *options, conn: conn,
	}
}

func (room *GameRoom) Init() {
	// new ground
	room.ground = NewGround(
		room.options.GroundWith, room.options.GroundHeight, room.options.GroundSymbol,
	)

	// new border
	room.border = NewRecBorder(
		room.options.BorderWidth, room.options.BorderHeight, room.options.BorderSymbol,
	)

	// new food
	room.food = NewFood(
		room.options.FoodSymbol, Limit{
			MinX: 1, MaxX: room.options.BorderWidth - 2,
			MinY: 1, MaxY: room.options.BorderHeight - 2,
		},
	)

	// create auto move ticker
	room.autoticker = time.NewTicker(time.Duration(room.options.AutoMoveIntervalMS) * time.Millisecond)

	// make room players map
	room.players = make(map[string]*Player, room.options.PlayerSize)

	// make room data channel
	room.dataChan = make(chan *RoomData, 1)

}

func (room *GameRoom) Run(ctx context.Context) {
	room.Init()
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-room.dataChan:
			room.handleData(data)
		case <-room.autoticker.C:
			room.handleAutoTicker()
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

func (room *GameRoom) handleData(data *RoomData) {
	player, err := room.getPlayer(data.Sender)
	if err == nil {
		fmt.Printf("[R] %s %s\n", player.ID, data.ClientData.CMD)
		player.UpdateLastRecv()
		room.handlePlayerCMD(data.ClientData.CMD, player)
		room.sendAllPlayersData()
	}
}

func (room *GameRoom) handleAutoTicker() {
	room.playersAutoMove()
	room.sendAllPlayersData()
}

func (room *GameRoom) handlePlayerCMD(cmd CMD, player *Player) {
	switch cmd {
	case CMDPause:
		room.playerPause(player)
	case CMDReplay:
		room.playerReplay(player)
	case CMDQuit:
		room.playerQuit(player)
	default:
		room.handlePlayerMovCMD(player, cmd)
	}
}

func (room *GameRoom) handlePlayerMovCMD(player *Player, cmd CMD) {
	if dir, ok := GetCMDDir(cmd); ok {
		room.playerMove(player, dir, false)
	}
}

func (room *GameRoom) getPlayer(addr *net.UDPAddr) (player *Player, err error) {
	playerID := addr.String()
	player = room.players[playerID]
	if player != nil {
		return
	}
	if len(room.players) > room.options.PlayerSize {
		err = errors.New("players are too more")
		return
	}
	player, err = NewPlayer(
		room.options.PlayerOptions,
		addr,
		playerID,
	)
	if err != nil {
		return
	}
	room.players[player.ID] = player
	return
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

func (room *GameRoom) playerMove(player *Player, dir Direction, oeated bool) (ieated bool) {
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
	for opid, oplayer := range room.players {
		if opid == player.ID {
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

func (room *GameRoom) playerPause(player *Player) {
	if !player.Over {
		player.Pause = true
	}
}

func (room *GameRoom) playerReplay(player *Player) {
	if player.Over {
		player.Reset()
	}
}

func (room *GameRoom) playerQuit(player *Player) {
	delete(room.players, player.ID)
}

func (room *GameRoom) playersAutoMove() {
	eated := false
	for _, player := range room.players {
		if player.Pause {
			continue
		}
		eated = room.playerMove(
			player, player.snake.GetDir(), eated,
		)
	}
}
