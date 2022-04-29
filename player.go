package gosnake

import (
	"net"
	"strings"
	"time"
)

type PlayerStatusFlag uint8

const (
	PlayerStatusPause PlayerStatusFlag = 1 << iota
	PlayerStatusOver
)

func (ps *PlayerStatusFlag) Set(flag PlayerStatusFlag) {
	*ps |= flag
}

func (ps *PlayerStatusFlag) UnSet(flag PlayerStatusFlag) {
	*ps |= ^flag
}

func (ps *PlayerStatusFlag) Is(flag PlayerStatusFlag) bool {
	return *ps&flag == flag
}

type Player struct {
	snake    *Snake
	snakeMap *Bitmap2D

	id        string
	addr      *net.UDPAddr
	state     PlayerState
	over      bool
	pause     bool
	score     uint16
	lastRecv  time.Time
	createdAt time.Time
}

func NewPlayer(addr *net.UDPAddr, playerID string, snakePosLimit Limit) (player *Player, err error) {
	now := time.Now()
	player = &Player{
		id:        playerID,
		addr:      addr,
		lastRecv:  now,
		createdAt: now,
	}
	player.snake = NewCenterPosSnake(snakePosLimit)
	return
}

func (player *Player) GetID() string {
	return player.id
}

func (player *Player) GetOver() bool {
	return player.over
}

func (player *Player) GetPause() bool {
	return player.pause
}

func (player *Player) GetScore() uint16 {
	return player.score
}

func (player *Player) GetAddr() net.UDPAddr {
	return *player.addr
}

func (player *Player) GetLastRecv() time.Time {
	return player.lastRecv
}

func (player *Player) UpdateLastRecv() {
	player.lastRecv = time.Now()
}
func (player *Player) IsSnakeTaken(pos Position) bool {
	return player.snake.IsTaken(pos)
}

func (player *Player) GetSnakeTailPos() Position {
	return player.snake.GetTailPos()
}

func (player *Player) GetSnakeDir() Direction {
	return player.snake.GetDir()
}

func (player *Player) Reset(snakePosLimit Limit) {
	player.snake = NewCenterPosSnake(snakePosLimit)
	player.UnPause()
	player.UnOver()
}

func (player *Player) GetSnakeNextHeadPos(dir Direction) *Position {
	return player.snake.GetNextHeadPos(dir)
}

func (player *Player) Pause() {
	player.pause = true
}

func (player *Player) UnPause() {
	player.pause = false
}

func (player *Player) Over() {
	player.over = true
}

func (player *Player) UnOver() {
	player.over = false
}

func (player *Player) GetSnakeTakes() map[Position]struct{} {
	return player.snake.GetTakes()
}

func (player *Player) MoveSnake(dir Direction) {
	player.snake.Move(dir)
}

func (player *Player) GrowSnake() {
	player.snake.Grow()
	player.score += 1
}

type PlayerStat struct {
	ID    string
	Score uint16
	Pause bool
	Over  bool
}

type PlayerStats []*PlayerStat

func (player *Player) GetStat() *PlayerStat {
	return &PlayerStat{
		ID:    player.id,
		Score: player.score,
		Pause: player.pause,
		Over:  player.over,
	}
}
func (stats PlayerStats) Len() int {
	return len(stats)
}

func (stats PlayerStats) Swap(i, j int) {
	stats[i], stats[j] = stats[j], stats[i]
}

func (stats PlayerStats) Less(i, j int) bool {
	if stats[i].Score == stats[j].Score {
		return strings.Compare(stats[i].ID, stats[j].ID) > 0
	}
	return stats[i].Score > stats[j].Score
}
