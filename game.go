package gosnake

import (
	"errors"
	"fmt"
	"gosnake/keys"
	"os"
	"os/exec"
	"time"
)

type GameOptions struct {
	GroundWith    int       `json:"ground_width"`
	GroundHeight  int       `json:"ground_height"`
	GroundSymbol  string    `json:"ground_symbol"`
	BordersWidth  int       `json:"borders_width"`
	BordersHeight int       `json:"borders_height"`
	BordersSymbol string    `json:"borders_symbol"`
	SnakeInitPosX int       `json:"snake_init_pos_x"`
	SnakeInitPosY int       `json:"snake_init_pos_y"`
	SnakeInitDir  Direction `json:"snake_init_dir"`
	SnakeSymbol   string    `json:"snake_symbol"`
	SnakeSpeedMS  int64     `json:"snake_speed_ms"`
	FoodSymbol    string    `json:"food_symbol"`
}

type Game struct {
	options   *GameOptions
	ground    *Ground
	borders   *Borders
	Snake     *Snake
	food      *Food
	ticker    *time.Ticker
	keycodech <-chan keys.Code
	status    gameStatusCode
}

func validateGameOptions(options GameOptions) error {
	// todo: validate game options, if invalid, return no-nil error.
	return nil
}

func NewGame(options *GameOptions) (*Game, error) {
	if err := validateGameOptions(*options); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}
	return &Game{
		options: options,
		status:  gameStatusCodeStopped,
	}, nil
}

func (game *Game) reload() {
	game.ground = NewGround(
		game.options.GroundWith,
		game.options.GroundHeight,
		game.options.GroundSymbol,
	)
	game.borders = NewRecBorders(
		game.options.BordersWidth,
		game.options.BordersHeight,
		game.options.BordersSymbol,
	)
	game.Snake = NewSnake(
		game.options.SnakeInitPosX,
		game.options.SnakeInitPosY,
		game.options.SnakeInitDir,
		game.options.SnakeSymbol,
	)
	game.food = NewFood(
		1, game.options.BordersWidth-2,
		1, game.options.BordersHeight-2,
		game.options.FoodSymbol,
	)
}

func (game *Game) Run() (err error) {
	// set running status from stopped
	if !game.status.setRunningFromStopped() {
		err = errors.New("game is not stopped")
	}
	defer game.status.setStopped()

	// load game objects
	game.reload()

	// listen keys event
	game.keycodech, err = keys.ListenEvent()
	if err != nil {
		return
	}
	defer keys.StopEventListen()

	// create game ticker
	game.ticker = time.NewTicker(
		time.Duration(game.options.SnakeSpeedMS) * time.Millisecond,
	)
	defer game.ticker.Stop()

	// close the cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	// clear screen
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	// render scene
	for {
		select {
		case keycode := <-game.keycodech:
			switch keycode {
			case keys.CodeQuit:
				fmt.Print("\r\033[K\033[3m* \033[0m\033[3;7mQUIT\033[0m\n\r")
				return
			case keys.CodePause:
				if game.status.isRunning() {
					game.status.setPaused()
					fmt.Print("\r\033[K\033[3m* \033[0m\033[3;7mPAUSE\033[0m\r")
				}
			case keys.CodeReplay:
				if game.status.isOver() {
					game.reload()
					game.status.setRunning()
				}
			default:
				if dir, ok := KeyCodeToDir[keycode]; ok {
					if game.status.isPaused() {
						game.status.setRunning()
					}
					game.Snake.Move(dir)
				}
			}
		case <-game.ticker.C:
			if game.status.isRunning() {
				game.Snake.Move(game.Snake.GetDir())
			}
		}

		if game.status.isRunning() {
			if game.food.IsTaken(game.Snake.head.pos) {
				game.Snake.Grow()
				game.food.UpdatePos()
				// bell
				fmt.Print("\a")
			}
			if game.borders.IsTaken(game.Snake.head.pos) ||
				game.Snake.IsTouchSelf() {
				fmt.Print("\r\033[K\033[3m* \033[0m\033[3;7mGAME OVER\033[0m")
				game.status.setOver()
				continue
			}
			game.ground.Render(game.borders, game.Snake, game.food)
			fmt.Printf("\r==================================================\n")
			fmt.Printf("\r\033[3m* Copyright 2022 Steve Zhang. All rights reserved.\033[0m\n")
			fmt.Printf("\r\033[3m* p) Pause; r) Replay; q) Quit\033[0m\n")
			fmt.Printf("\r\033[3m* Score: %03d\033[0m\n", game.Snake.Len()-1)
			fmt.Print("\r\033[K\033[3m* \033[0m\033[3;7mRUN\033[0m\r")
		}
	}
}
