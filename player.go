package gosnake

import (
	"gosnake/base"
	"net"
	"time"
)

const (
	PlayerStatusPause base.Flag8 = 1 << iota
	PlayerStatusOver
)

type Player struct {
	addr       *net.UDPAddr
	score      uint16
	status     base.Flag8
	snake      *Snake
	lastPingAt time.Time
	createdAt  time.Time
}

func NewPlayer(addr *net.UDPAddr) *Player {
	return &Player{
		addr:      addr,
		score:     0,
		status:    PlayerStatusPause,
		createdAt: time.Now(),
	}
}

func (player *Player) GetAddr() *net.UDPAddr {
	return player.addr
}

func (player *Player) GetSnakeBitmap() *base.Bitmap2D {
	if player.snake == nil {
		return &base.Bitmap2D{}
	}
	return player.snake.GetBitmap()
}

func (player *Player) GetSnakeLen() int {
	return player.snake.length
}

func (player *Player) ResetScore() {
	player.score = 0
}

func (player *Player) IncreaseScore() {
	player.score += 1
}

func (player *Player) UpdateLastPingedAt() {
	player.lastPingAt = time.Now()
}

func (player *Player) SetSnake(pos base.Position2D, dir base.Direction2D) {
	player.snake = NewSnake(pos, dir)
}

func (player *Player) UnsetSnake() {
	player.snake = nil
}

func (player *Player) MoveSnake(dir base.Direction2D) {
	player.snake.Move(dir)
}

func (player *Player) GrowSnake() {
	player.snake.Grow()
}

func (player *Player) GetNextSnakeHeadPos(dir base.Direction2D) *base.Position2D {
	return player.snake.GetNextHeadPos(dir)
}

func (player *Player) GetSnakeTailPos() base.Position2D {
	return player.snake.TailPos()
}

func (player *Player) GetSnakeHeadPos() base.Position2D {
	return player.snake.HeadPos()
}
func (player *Player) GetSnakeDir() base.Direction2D {
	return player.snake.dir
}

func (player *Player) GetPlayerData() *PlayerData {
	return &PlayerData{
		Addr:          player.addr,
		Status:        player.status,
		Score:         player.score,
		Snake:         player.GetSnakeBitmap(),
		CreatedAtUnix: player.createdAt.Unix(),
	}
}

type PlayerData struct {
	Addr          *net.UDPAddr
	Status        base.Flag8
	Score         uint16
	Snake         *base.Bitmap2D
	CreatedAtUnix int64
}

type PlayersData []*PlayerData

func (ps PlayersData) Len() int {
	return len(ps)
}

func (ps PlayersData) Less(i, j int) bool {
	if ps[i].Score != ps[j].Score {
		return ps[i].Score < ps[j].Score
	}
	return ps[i].CreatedAtUnix > ps[j].CreatedAtUnix
}

func (ps PlayersData) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}
