package gosnake

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

const clearPlayerTimeInterval = 10 * time.Second

type RoomOptions struct {
	BorderWidth        int `json:"border_width"`
	BorderHeight       int `json:"border_height"`
	AutoMoveIntervalMS int `json:"auto_move_interval_ms"`
	PlayerSize         int `json:"player_size"`
}

type Room struct {
	options            RoomOptions
	players            map[string]*Player
	border             *RecBorder
	food               *Food
	autoticker         *time.Ticker
	clearPlayersTicker *time.Ticker
	dataChan           chan *RoomData
	sendData           func([]byte, *net.UDPAddr)
	posLimit           Limit
}

func NewRoom(options *RoomOptions, sendData func([]byte, *net.UDPAddr)) *Room {
	return &Room{
		options:  *options,
		sendData: sendData,
		posLimit: Limit{
			MinX: 1, MaxX: options.BorderWidth - 2,
			MinY: 1, MaxY: options.BorderHeight - 2,
		},
	}
}

func (room *Room) Init() {
	// new border
	room.border = NewRecBorder(
		room.options.BorderWidth, room.options.BorderHeight,
		"",
	)

	// new food
	room.food = NewFood(room.posLimit)

	// create auto move ticker
	room.autoticker = time.NewTicker(time.Duration(room.options.AutoMoveIntervalMS) * time.Millisecond)

	// create clear disconnected players ticker
	room.clearPlayersTicker = time.NewTicker(clearPlayerTimeInterval)

	// make room players map
	room.players = make(map[string]*Player, room.options.PlayerSize)

	// make room data channel
	room.dataChan = make(chan *RoomData, 1)

}

func (room *Room) Run(ctx context.Context) {
	room.Init()
	for {
		select {
		case <-ctx.Done():
			return
		case data := <-room.dataChan:
			room.handleData(data)
		case <-room.autoticker.C:
			room.handleAutoTicker()
		case <-room.clearPlayersTicker.C:
			room.clearDisconnectedPlayers()
		}
	}
}

type RoomData struct {
	Sender     *net.UDPAddr
	ClientData *ClientData
}

func (room *Room) HandleData(data *RoomData) {
	room.dataChan <- data
}

func (room *Room) handleData(data *RoomData) {
	player, err := room.getPlayer(data.Sender)
	if err == nil {
		fmt.Printf("[R] %s %s\n", player.GetID(), data.ClientData.CMD)
		player.UpdateLastRecv()
		room.handlePlayerCMD(data.ClientData.CMD, player)
		room.sendAllPlayersData()
	}
}

func (room *Room) clearDisconnectedPlayers() {
	now := time.Now()
	for playerID, player := range room.players {
		lastRecv := player.GetLastRecv()
		if lastRecv.Add(clearPlayerTimeInterval).Before(now) {
			fmt.Printf("[C] %s %s\n", playerID, lastRecv.Format("15:03:04"))
			delete(room.players, playerID)
		}
	}
}

func (room *Room) handleAutoTicker() {
	room.playersAutoMove()
	room.sendAllPlayersData()
}

func (room *Room) handlePlayerCMD(cmd CMD, player *Player) {
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

func (room *Room) handlePlayerMovCMD(player *Player, cmd CMD) {
	if dir, ok := GetCMDDir(cmd); ok {
		room.playerMove(player, dir, false)
	}
}

func (room *Room) getPlayer(addr *net.UDPAddr) (player *Player, err error) {
	playerID := addr.String()
	player = room.players[playerID]
	if player != nil {
		return
	}
	if len(room.players) > room.options.PlayerSize {
		err = errors.New("players are too more")
		return
	}
	player, err = NewPlayer(addr, playerID, room.posLimit)
	if err != nil {
		return
	}
	room.players[playerID] = player
	return
}

func (room *Room) sendAllPlayersData() {
	wg := &sync.WaitGroup{}
	for _, player := range room.players {
		wg.Add(1)
		go func(player *Player) {
			defer wg.Done()
			sceneData := room.getPlayerSceneData(player)
			data := sceneData.Encode()
			addr := player.GetAddr()
			room.sendData(data, &addr)
		}(player)
	}
	wg.Wait()
}

func (room *Room) getPlayerSceneData(player *Player) *SceneData {
	w := room.options.BorderWidth
	h := room.options.BorderWidth
	sceneData := &SceneData{
		PlayerID:     player.GetID(),
		BorderWidth:  w,
		BorderHeight: h,
		PlayerSnake:  NewCompressLayer(w, h),
		Snakes:       NewCompressLayer(w, h),
		Food:         NewCompressLayer(w, h),
		PlayerStats:  make(PlayerStats, 0),
	}
	sceneData.PlayerSnake.AddPositions(player.GetSnakeTakes())
	sceneData.Food.AddPositions(room.food.GetTakes())
	for _, rplayer := range room.players {
		sceneData.Snakes.AddPositions(
			rplayer.GetSnakeTakes(),
		)
		sceneData.PlayerStats = append(
			sceneData.PlayerStats,
			rplayer.GetStat(),
		)
	}
	return sceneData
}

func (room *Room) playerMove(player *Player, dir Direction, oeated bool) (ieated bool) {
	if player.GetOver() {
		return
	}
	player.UnPause()
	nextHeadPos := player.GetSnakeNextHeadPos(dir)
	if nextHeadPos == nil {
		return
	}
	if room.border.IsTaken(*nextHeadPos) {
		player.Over()
		return
	}
	if player.IsSnakeTaken(*nextHeadPos) &&
		*nextHeadPos != player.GetSnakeTailPos() {
		player.Over()
		return
	}
	for opid, oplayer := range room.players {
		if opid == player.GetID() {
			continue
		}
		if oplayer.IsSnakeTaken(*nextHeadPos) {
			player.Over()
			return
		}
	}
	player.MoveSnake(dir)
	if !oeated && room.food.IsTaken(*nextHeadPos) {
		player.GrowSnake()
		room.food.UpdatePos()
		ieated = true
	}

	return
}

func (room *Room) playerPause(player *Player) {
	if player.GetOver() {
		return
	}
	if player.GetPause() {
		player.UnPause()
		return
	}
	player.Pause()
}

func (room *Room) playerReplay(player *Player) {
	if player.GetOver() {
		player.Reset(room.posLimit)
	}
}

func (room *Room) playerQuit(player *Player) {
	delete(room.players, player.GetID())
}

func (room *Room) playersAutoMove() {
	eated := false
	for _, player := range room.players {
		if player.GetPause() {
			continue
		}
		eated = room.playerMove(
			player, player.snake.GetDir(), eated,
		)
	}
}
