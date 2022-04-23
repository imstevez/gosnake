package gosnake

import (
	"fmt"
	"gosnake/keys"
	"os"
	"os/exec"
	"time"
)

// initOFLSPS init offline single player server game
func (game *Game) initInOFLSPS() (err error) {
	// new ground
	game.ground = NewGround(
		game.options.GroundWith, game.options.GroundHeight,
		game.options.GroundSymbol,
	)

	// new border
	game.border = NewRecBorder(
		game.options.BorderWidth, game.options.BorderHeight,
		game.options.BorderSymbol,
	)

	// new snake
	game.snake1 = NewSnake(
		game.options.Snake1InitPosX, game.options.Snake1InitPosY,
		game.options.Snake1InitDir, game.options.Snake1Symbol,
	)

	// new food
	game.food = NewFood(
		game.options.FoodSymbol, Limit{
			MinX: 1, MaxX: game.options.BorderWidth - 2,
			MinY: 1, MaxY: game.options.BorderHeight - 2,
		},
	)

	// new texts
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

	// listen keyboard events
	game.keyEvents, err = keys.ListenEvent()
	if err != nil {
		return err
	}
	game.clears = append(
		game.clears, keys.StopEventListen,
	)

	// create ticker for auto move
	game.autoMoveTicker = time.NewTicker(
		time.Duration(game.options.SnakeSpeedMS) * time.Millisecond,
	)
	game.clears = append(
		game.clears, game.autoMoveTicker.Stop,
	)

	// create ticker for render
	game.renderTicker = time.NewTicker(
		time.Duration(1000/game.options.FPS) * time.Millisecond,
	)
	game.clears = append(
		game.clears, game.renderTicker.Stop,
	)

	// close the cursor
	fmt.Print("\033[?25l")
	game.clears = append(game.clears, func() {
		fmt.Print("\033[?25h")
	})

	// clear screen
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	return
}

func (game *Game) reloadInOFLSPS() {
	// new snake
	game.snake1 = NewSnake(
		game.options.Snake1InitPosX, game.options.Snake1InitPosY,
		game.options.Snake1InitDir, game.options.Snake1Symbol,
	)

	// update food
	game.food.UpdatePos()

	// reset status
	game.over1 = false
}

// runOffline run offline game
func (game *Game) runInOFLSPS() (err error) {
	// init game
	err = game.initInOFLSPS()
	if err != nil {
		return
	}

	// define control keycode funcs
	var keycodesFuncs = map[keys.Code]func(){
		keys.CodeQuit: func() {
			game.quit1 = true
		},
		keys.CodePause: func() {
			if !game.quit1 && !game.over1 {
				game.pause1 = true
			}
		},
		keys.CodeReplay: func() {
			if !game.quit1 && game.over1 {
				game.reloadInOFLSPS()
			}
		},
	}

	// define direction key func
	for keycode, dir := range keyCodeToDir {
		idir := dir
		keycodesFuncs[keycode] = func() {
			game.snake1Mov(idir)
		}
	}

loop:
	// listen events, update objects, calc status, render frame.
	for {
		select {
		case keycode := <-game.keyEvents: // handle keyboard events
			if keyfunc, ok := keycodesFuncs[keycode]; ok {
				keyfunc()
			}
		case <-game.autoMoveTicker.C: // auto move snake
			game.offlineAutoMove()
		case <-game.renderTicker.C: // render frame
			state := IfStr(game.over1, "Over", "Run")
			state = IfStr(game.pause1, "Pause", state)
			state = IfStr(game.quit1, "Quit", state)
			score := game.snake1.Len() - 1
			text := game.texts.Sprintlines(score, state)
			frame := game.ground.Render(
				game.food, game.border, game.snake1,
			).HozJoin(
				text, game.ground.GetWidth()*2,
			).Merge()
			fmt.Print(frame)
			if game.quit1 {
				fmt.Print("\r\n\r")
				return
			}
		}
	}
}

func (game *Game) autoMoveInOFLSPS() {
	if game.quit1 || game.pause1 || game.over1 {
		return
	}
	dir := game.snake1.GetDir()
	game.snake1Mov(dir)
}

func (game *Game) offlineHandleKey() {

}
