package gosnake

import (
	"errors"
	"gosnake/keys"
	"sync/atomic"
	"time"
)

type GameOptions struct {
	GroundWith          int       `json:"ground_width"`
	GroundHeight        int       `json:"ground_height"`
	GroundSymbol        string    `json:"ground_symbol"`
	BordersWidth        int       `json:"borders_width"`
	BordersHeight       int       `json:"borders_height"`
	BordersSymbol       string    `json:"borders_symbol"`
	SnakeInitPosX       int       `json:"snake_init_pos_x"`
	SnakeInitPosY       int       `json:"snake_init_pos_y"`
	SnakeInitDir        Direction `json:"snake_init_dir"`
	SnakeSymbol         string    `json:"snake_symbol"`
	ClientSnakeInitPosX int       `json:"client_snake_init_pos_x"`
	ClientSnakeInitPosY int       `json:"client_snake_init_pos_y"`
	ClientSnakeInitDir  Direction `json:"client_snake_init_dir"`
	ClientSnakeSymbol   string    `json:"client_snake_symbol"`
	SnakeSpeedMS        int64     `json:"snake_speed_ms"`
	FoodSymbol          string    `json:"food_symbol"`
	Online              bool      `json:"online"`
	Server              bool      `json:"server"`
	LocalAddr           string    `json:"local_addr"`
	RemoteAddr          string    `json:"remote_addr"`
	FPS                 int       `json:"fps"`
}

type Game struct {
	options        *GameOptions
	ground         *Ground
	border         *Border
	snake1         *Snake
	snake2         *Snake
	food           *Food
	texts          Lines
	clears         []func()
	keyEvents      <-chan keys.Code
	renderTicker   *time.Ticker
	autoMoveTicker *time.Ticker
	network        *Network
	runnig         int32
}

func NewGame(options *GameOptions) (*Game, error) {
	return &Game{
		options: options,
	}, nil
}

var ErrGameIsNotStopped = errors.New("game is not stoppted")

func (game *Game) setRunningFromStopped() error {
	if !atomic.CompareAndSwapInt32((*int32)(&game.runnig), 0, 1) {
		return ErrGameIsNotStopped
	}
	return nil
}

func (game *Game) setStopped() {
	atomic.StoreInt32((*int32)(&game.runnig), 0)
}

func (game *Game) Run() error {
	if err := game.setRunningFromStopped(); err != nil {
		return err
	}
	defer game.setStopped()
	if !game.options.Online {
		return game.runOffline()
	}
	return game.runOnline()
}

func (game *Game) clear() {
	for i := len(game.clears) - 1; i >= 0; i-- {
		cfunc := game.clears[i]
		cfunc()
	}
}
