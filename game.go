package gosnake

import (
	"errors"
	"gosnake/keys"
	"sync/atomic"
	"time"
)

type GameMode int

const (
	GameModeOFLSPS GameMode = iota // offline single player as server
	GameModeONLDPS                 // online double players as server
	GameModeONLDPC                 // online double players as client
)

// GameOptions contain all options that can be personalized
type GameOptions struct {
	GroundWith     int       `json:"ground_width"`
	GroundHeight   int       `json:"ground_height"`
	GroundSymbol   string    `json:"ground_symbol"`
	BorderWidth    int       `json:"border_width"`
	BorderHeight   int       `json:"border_height"`
	BorderSymbol   string    `json:"border_symbol"`
	Snake1InitPosX int       `json:"snake_1_init_pos_x"`
	Snake1InitPosY int       `json:"snake_1_init_pos_y"`
	Snake1InitDir  Direction `json:"snake_1_init_dir"`
	Snake1Symbol   string    `json:"snake_1_symbol"`
	Snake2InitPosX int       `json:"snake_2_init_pos_x"`
	Snake2InitPosY int       `json:"snake_2_init_pos_y"`
	Snake2InitDir  Direction `json:"snake_2_init_dir"`
	Snake2Symbol   string    `json:"snake_2_symbol"`
	SnakeSpeedMS   int       `json:"snake_speed_ms"`
	FoodSymbol     string    `json:"food_symbol"`
	Online         bool      `json:"online"`
	Server         bool      `json:"server"`
	LocalAddr      string    `json:"local_addr"`
	RemoteAddr     string    `json:"remote_addr"`
	FPS            int       `json:"fps"`
	Mode           int       `json:"mode"`
}

// Game caps all members of a snake game, include objects, status, events channel
// and other objects and methods, it provide a run entrance for
// running the game at the specified options
type Game struct {
	options        *GameOptions
	ground         *Ground
	border         *RecBorder
	snake1         *Snake
	snake2         *Snake
	over1          bool
	over2          bool
	pause1         bool
	pause2         bool
	quit1          bool
	quit2          bool
	food           *Food
	texts          Lines
	clears         []func()
	keyEvents      <-chan keys.Code
	renderTicker   *time.Ticker
	autoMoveTicker *time.Ticker
	network        *Network
	running        int32
	handleKey      func(keys.Code)
	handleAutoMove func()
}

// NewGame return Game ptr with specified options
func NewGame(options *GameOptions) *Game {
	return &Game{
		options: options,
	}
}

// Run run game with the setted options
func (game *Game) Run() (err error) {
	err = game.setRunningFromStopped()
	if err != nil {
		return err
	}

	defer game.setStopped()
	defer game.clear()

	if game.options.Online {
		err = game.runOnline()
		return
	}

	err = game.runOnline()
	return
}

// game can not be runned multiply at the same time
// this can be confirmed by a atomic swap check on game starting

var ErrGameIsNotStopped = errors.New("game is not stopped")

func (game *Game) setRunningFromStopped() error {
	if !atomic.CompareAndSwapInt32((*int32)(&game.running), 0, 1) {
		return ErrGameIsNotStopped
	}
	return nil
}

func (game *Game) setStopped() {
	atomic.StoreInt32((*int32)(&game.running), 0)
}

// clear registered resources be created at initializing
func (game *Game) clear() {
	for i := len(game.clears) - 1; i >= 0; i-- {
		cFunc := game.clears[i]
		cFunc()
	}
}

func (game *Game) moveServerSnakeInOFLSPS(dir Direction) {
	if game.quit1 || game.over1 {
		return
	}

	// disable pause
	game.pause1 = false

	// oppsite direction
	pos := game.snake1.GetNextHeadPos(dir)
	if pos == nil {
		return
	}

	// touch border
	if game.border.IsTaken(*pos) ||
		game.snake1.IsTaken(*pos) {
		game.over1 = true
		return
	}

	// touch self
	if game.snake1.IsTaken(*pos) &&
		game.snake1.GetTailPos() != *pos {
		game.over1 = true
		return
	}

	// move snake
	game.snake1.Move(dir)

	// eat food
	if game.food.IsTaken(game.snake1.GetHeadPos()) {
		game.food.UpdatePos()
		game.snake1.Grow()
	}
}

func (game *Game) moveServerSnakeInONLDPS(dir Direction) {
	if game.quit1 || game.over1 {
		return
	}

	// disable pause
	game.pause1 = false

	// oppsite direction
	pos := game.snake1.GetNextHeadPos(dir)
	if pos == nil {
		return
	}

	// touch border
	if game.border.IsTaken(*pos) ||
		game.snake1.IsTaken(*pos) {
		game.over1 = true
		return
	}

	// touch self
	if game.snake1.IsTaken(*pos) &&
		game.snake1.GetTailPos() != *pos {
		game.over1 = true
		return
	}

	// touch snake2
	if game.snake1.IsTaken(*pos) {

	}

	// move snake
	game.snake1.Move(dir)

	// eat food
	if game.food.IsTaken(game.snake1.GetHeadPos()) {
		game.food.UpdatePos()
		game.snake1.Grow()
	}
}

func (game *Game) moveClientSnakeInONLDPS(dir Direction) {
	if game.quit1 || game.over2 {
		return
	}

	// disable pause
	game.pause2 = false

	// oppsite direction
	pos := game.snake2.GetNextHeadPos(dir)
	if pos == nil {
		return
	}

	// touch border
	if game.border.IsTaken(*pos) ||
		game.snake2.IsTaken(*pos) {
		game.over2 = true
		return
	}

	// touch self
	if game.snake2.IsTaken(*pos) &&
		*pos != game.snake2.GetTailPos() {
		game.over2 = true
		return
	}

	// touch snake1
	if game.snake1.IsTaken(*pos) {
		game.over2 = true
		return
	}

	// mov snake2
	game.snake2.Move(dir)

	// eat food
	if game.food.IsTaken(*pos) {
		game.snake2.Grow()
		game.food.UpdatePos()
	}

	return
}
