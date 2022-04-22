package gosnake

import (
	"fmt"
	"gosnake/keys"
	"os"
	"os/exec"
	"time"
)

func (game *Game) load() {
	game.ground = NewGround(
		game.options.GroundWith,
		game.options.GroundHeight,
		game.options.GroundSymbol,
	)
	game.border = NewRecBorder(
		game.options.BordersWidth,
		game.options.BordersHeight,
		game.options.BordersSymbol,
	)
	game.snake = NewSnake(
		game.options.SnakeInitPosX,
		game.options.SnakeInitPosY,
		game.options.SnakeInitDir,
		game.options.SnakeSymbol,
		Limit{
			MinX: 1,
			MaxX: game.options.BordersWidth - 2,
			MinY: 1,
			MaxY: game.options.BordersHeight - 2,
		},
	)
	game.food = NewFood(
		game.options.FoodSymbol,
		Limit{
			MinX: 1,
			MaxX: game.options.BordersWidth - 2,
			MinY: 1,
			MaxY: game.options.BordersHeight - 2,
		},
	)
	game.texts = []string{
		" \033[3m===================================================\033[0m",
		" \033[7m GoSnake@v0.0.1 \033[0m",
		" \033[3m Copyright 2022 Steve Zhang. All rights reserved.\033[0m",
		" \033[3m===================================================\033[0m",
		" \033[3m* w,i) Up; a,j) Left; s,k) Down; d,l) Right;\033[0m",
		" \033[3m* p) Pause; r) Replay; q) Quit\033[0m",
		" \033[3m* Score: %04d\033[0m",
		" \033[3m* State: %s\033[0m",
	}
}

func (game *Game) runOffline() (err error) {
	// load objects
	game.load()

	// listen keys event
	keycodech, err := keys.ListenEvent()
	if err != nil {
		return
	}
	defer keys.StopEventListen()

	// create ticker for auto move
	autoMoveTicker := time.NewTicker(
		time.Duration(game.options.SnakeSpeedMS) * time.Millisecond,
	)
	defer autoMoveTicker.Stop()

	// create ticker for render
	renderTicker := time.NewTicker(20 * time.Millisecond)
	defer renderTicker.Stop()

	// close the cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	// clear screen
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	// define game state
	var (
		quit     bool
		pause    bool
		gameover bool
	)

	// define move keycodes funcs
	mov := func(dir Direction) {
		if !quit && !gameover {
			pause = false
			err := game.snake.Move(dir)
			if err != nil && err != ErrSnakeMovGoOppsite {
				gameover = true
				return
			}
			fmt.Print("\a")
			if game.IsEeatFood() {
				game.snake.Grow()
				game.food.UpdatePos()
				fmt.Print("\a")
			}
		}
	}

	// define control keycode funcs
	var keycodesFuncs = map[keys.Code]func(){
		keys.CodeQuit: func() {
			quit = true
		},
		keys.CodePause: func() {
			if !quit && !gameover {
				pause = true
			}
		},
		keys.CodeReplay: func() {
			if !quit {
				game.load()
				gameover = false
				pause = false
			}
		},
	}
	for keycode, dir := range keyCodeToDir {
		idir := dir
		keycodesFuncs[keycode] = func() {
			mov(idir)
		}
	}

loop:
	// listen events, update objects, calc status, render frame.
	for {
		select {
		case keycode := <-keycodech: // handle keyboard events
			if keyfunc, ok := keycodesFuncs[keycode]; ok {
				keyfunc()
			}
		case <-autoMoveTicker.C: // auto move snake
			if quit || pause || gameover {
				continue loop
			}
			dir := game.snake.GetDir()
			mov(dir)
		case <-renderTicker.C: // render frame
			state := NoitherStr(gameover, "Over", "Run")
			state = NoitherStr(pause, "Pause", state)
			state = NoitherStr(quit, "Quit", state)
			score := game.snake.Len() - 1
			text := game.texts.Sprintlines(score, state)
			frame := game.ground.Render(
				game.food, game.border, game.snake,
			).HozJoin(
				text, game.ground.GetWidth()*2,
			).Merge()
			fmt.Print(frame)
			if quit {
				fmt.Print("\r\n")
				return
			}
		}
	}
}