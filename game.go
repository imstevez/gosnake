package gosnake

import (
	"errors"
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
	options     *GameOptions
	ground      *Ground
	border      *Border
	snake       *Snake
	clientSnake *Snake
	food        *Food
	status      gameStatusCode
	network     *Network
	texts       Lines
}

func NewGame(options *GameOptions) (*Game, error) {
	return &Game{
		options: options,
		status:  gameStatusCodeStopped,
	}, nil
}

func (game *Game) Run() error {
	if !game.status.setRunningFromStopped() {
		return errors.New("game is not stopped")
	}
	defer game.status.setStopped()
	if !game.options.Online {
		return game.runOffline()
	}
	return game.runOnline()
}

func (game *Game) IsEeatFood() bool {
	return game.food.GetSymbolAt(game.snake.head.pos) != ""
}

func (game *Game) IsClientEeatFood() bool {
	return game.food.GetSymbolAt(game.clientSnake.head.pos) != ""
}
