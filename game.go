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
	borders        *Borders
	snake          *Snake
	clientSnake    *Snake
	food           *Food
	autoMoveTicker *time.Ticker
	keycodech      <-chan keys.Code
	status         gameStatusCode
	network        *Network
	renderTicker   *time.Ticker
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
	game.snake = NewSnake(
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
	if !game.status.setRunningFromStopped() {
		err = errors.New("game is not stopped")
		return
	}
	defer game.status.setStopped()

	if game.options.Online {
		if game.options.Server {
			return game.RunOnlineServer()
		}
		return game.RunOnlineClient()
	}
	return game.RunOffline()
}

func (game *Game) RunOffline() (err error) {
	// load game objects
	game.reload()

	// listen keys event
	game.keycodech, err = keys.ListenEvent()
	if err != nil {
		return
	}
	defer keys.StopEventListen()

	// create ticker for auto move
	game.autoMoveTicker = time.NewTicker(
		time.Duration(game.options.SnakeSpeedMS) * time.Millisecond,
	)
	defer game.autoMoveTicker.Stop()

	// create ticker for render
	game.renderTicker = time.NewTicker(
		30 * time.Millisecond,
	)
	defer game.renderTicker.Stop()

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
					game.snake.Move(dir)
				}
			}
		case <-game.autoMoveTicker.C:
			if game.status.isRunning() {
				game.snake.Move(game.snake.GetDir())
			}
		}

		if game.status.isRunning() {
			if game.food.IsTaken(game.snake.head.pos) {
				game.snake.Grow()
				game.food.UpdatePos()
				// bell
				fmt.Print("\a")
			}
			if game.borders.IsTaken(game.snake.head.pos) ||
				game.snake.IsTouchSelf() {
				fmt.Print("\r\033[K\033[3m* \033[0m\033[3;7mGAME OVER\033[0m")
				game.status.setOver()
				continue
			}
			result := game.ground.Render(game.borders, game.snake, game.food)
			returnCursor(100)
			fmt.Print(result)
			fmt.Printf("\r==================================================\n")
			fmt.Printf("\r\033[3m* Copyright 2022 Steve Zhang. All rights reserved.\033[0m\n")
			fmt.Printf("\r\033[3m* p) Pause; r) Replay; q) Quit\033[0m\n")
			fmt.Printf("\r\033[3m* Score: %03d\033[0m\n", game.snake.Len()-1)
			fmt.Print("\r\033[K\033[3m* \033[0m\033[3;7mRUN\033[0m\r")
		}
	}
}

type Scene struct {
	Borders     *Borders
	Snake       *Snake
	ClientSnake *Snake
	Food        *Food
}
