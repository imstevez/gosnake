package gosnake

import (
	"fmt"
	"gosnake/keys"
	"os"
	"os/exec"
	"time"
)

func (game *Game) runOnline() error {
	if !game.options.Server {
		return game.runOnlineClient()
	}
	return game.runOnlineServer()
}

func (game *Game) onlineReload() {
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
	game.clientSnake = NewSnake(
		game.options.ClientSnakeInitPosX,
		game.options.ClientSnakeInitPosY,
		game.options.ClientSnakeInitDir,
		game.options.ClientSnakeSymbol,
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
		"\033[3m* Player: %s\033[0m",
		"=========",
		"\033[3m* Copyright 2022 Steve Zhang. All rights reserved.\033[0m",
		"\033[3m* w,i) Up; a,j) Left; s,k) Down; d,l) Right;\033[0m",
		"\033[3m* p) Pause; r) Replay; q) Quit\033[0m",
		"\033[3m* Score: player1-%04d-%-4s | player2-%04d%-4s\033[0m",
		"\033[3m* \033[0m\033[3;7m%s\033[0m",
	}
}

func (game *Game) runOnlineServer() (err error) {
	game.network = NewNetWork()
	err = game.network.Start(
		game.options.LocalAddr,
		game.options.RemoteAddr,
	)
	if err != nil {
		return
	}
	defer game.network.Stop()

	// close the cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	// clear screen
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	fmt.Print("\rWaiting for player2...\n\r")

	// reload game
	game.onlineReload()

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

	var (
		gameover       bool
		clientGameover bool
		paused         bool
		quit           bool
		connected      bool
	)

	mov := func(dir Direction) {
		if !quit && !gameover {
			paused = false
			err := game.snake.Move(dir)
			if err != nil && err != ErrSnakeMovGoOppsite {
				gameover = true
				return
			}
			if game.clientSnake.GetSymbolAt(game.snake.GetHeadPos()) != "" {
				gameover = true
				return
			}
			if game.IsEeatFood() {
				game.snake.Grow()
				game.food.UpdatePos()
			}
		}
	}

	clientMov := func(dir Direction) {
		if !quit && !clientGameover {
			err := game.clientSnake.Move(dir)
			if err != nil && err != ErrSnakeMovGoOppsite {
				clientGameover = true
				return
			}
			if game.snake.GetSymbolAt(game.clientSnake.GetHeadPos()) != "" {
				clientGameover = true
				return
			}
			if game.IsEeatFood() {
				game.clientSnake.Grow()
				game.food.UpdatePos()
			}
		}
	}
Loop:
	for {
		select {
		case keycode := <-keycodech:
			if quit {
				continue Loop
			}
			switch keycode {
			case keys.CodePause:
				if connected && !gameover {
					paused = true
				}
				continue Loop
			case keys.CodeQuit:
				quit = true
				continue Loop
			case keys.CodeReplay:
				if connected && gameover && clientGameover {
					game.onlineReload()
					gameover = false
					clientGameover = false
				}
				continue Loop
			default:
				if !connected || gameover {
					continue Loop
				}
				if dir, ok := keyCodeToDir[keycode]; ok {
					mov(dir)
				}
			}
		case data := <-game.network.Recv:
			if quit {
				continue Loop
			}
			msg := decodeMessage(data)
			switch msg.CMD {
			case MSGCMDPing:
				connected = true
				continue Loop
			case MSGCMDMov:
				if !connected || paused || clientGameover {
					continue Loop
				}
				dir := decodeDirData(msg.Data)
				clientMov(dir)
				game.clientSnake.Move(dir)
			}
		case <-autoMoveTicker.C:
			if !connected || paused || quit {
				continue Loop
			}
			if !gameover {
				dir := game.snake.GetDir()
				mov(dir)
			}
			if !clientGameover {
				dir := game.clientSnake.GetDir()
				clientMov(dir)
			}
		case <-renderTicker.C:
			if !connected && quit {
				return
			}
			if !connected {
				continue Loop
			}
			p1Score := game.snake.Len() - 1
			p1State := NoitherStr(gameover, "OVER", "OK")
			p2Score := game.clientSnake.Len() - 1
			p2State := NoitherStr(clientGameover, "OVER", "OK")
			gameStatus := NoitherStr(paused, "PAUSED", "RUN")
			gameStatus = NoitherStr(quit, "QUIT", gameStatus)
			texts := game.texts.Sprintlines(
				"%d", p1Score, p1State, p2Score, p2State, gameStatus,
			)
			frame := game.ground.Render(
				game.food, game.snake, game.clientSnake, game.border,
			).HozJoin(
				texts, game.ground.GetWidth()*2,
			).Merge()
			renderMsg := encodeRenderMsg(fmt.Sprintf(frame, 2))
			game.network.Send <- renderMsg
			fmt.Printf(frame, 1)
			if quit {
				fmt.Print("\r\n")
				return
			}
		}
	}
}

func (game *Game) runOnlineClient() (err error) {
	game.network = NewNetWork()
	err = game.network.Start(
		game.options.LocalAddr,
		game.options.RemoteAddr,
	)
	if err != nil {
		return
	}
	defer game.network.Stop()

	// close the cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	// clear screen
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	fmt.Print("\r Waiting for player1...\n\r")

	// listen keys event
	keycodech, err := keys.ListenEvent()
	if err != nil {
		return
	}
	keys.StopEventListen()

	// create ping ticker
	pingTicker := time.NewTicker(1 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case keycode := <-keycodech:
			if keycode == keys.CodeQuit {
				return
			}
			if dir, ok := keyCodeToDir[keycode]; ok {
				game.network.Send <- encodeMovMsg(dir)
			}
		case data := <-game.network.Recv:
			msg := decodeMessage(data)
			switch msg.CMD {
			case MSGCMDRender:
				result := decodeRenderData(msg.Data)
				fmt.Print(result)
			}
		case <-pingTicker.C:
			game.network.Send <- encodePingMsg()
		}
	}
}
