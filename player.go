package gosnake

import (
	"net"
	"time"
)

type PlayerOptions struct {
	SnakeSymbol string
	SnakeLimit  Limit
}

type Player struct {
	options *PlayerOptions
	snake   *Snake

	ID         string
	Addr       *net.UDPAddr
	SnakeTakes map[Position]struct{}
	Over       bool
	Pause      bool
	Score      int
	LastRecv   time.Time
	CreatedAt  time.Time
}

func NewPlayer(options *PlayerOptions, addr *net.UDPAddr, playerID string) (player *Player, err error) {
	now := time.Now()
	player = &Player{
		options:   options,
		ID:        playerID,
		Addr:      addr,
		LastRecv:  now,
		CreatedAt: now,
	}
	player.snake = NewCenterPosSnake(
		options.SnakeLimit, options.SnakeSymbol,
	)
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
