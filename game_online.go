package gosnake

import (
	"fmt"
	"gosnake/keys"
	"os"
	"os/exec"
	"time"
)

func (game *Game) onlineReload() {
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
	game.clientSnake = NewSnake(
		game.options.ClientSnakeInitPosX,
		game.options.ClientSnakeInitPosY,
		game.options.ClientSnakeInitDir,
		game.options.ClientSnakeSymbol,
	)
	game.food = NewFood(
		1, game.options.BordersWidth-2,
		1, game.options.BordersHeight-2,
		game.options.FoodSymbol,
	)
}

func (game *Game) RunOnlineServer() (err error) {
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
	game.renderTicker = time.NewTicker(30 * time.Millisecond)
	defer game.renderTicker.Stop()

	var (
		gameover       bool
		clientGameover bool
		paused         bool
		quit           bool
		connected      bool
	)

Loop:
	for {
		select {
		case keycode := <-game.keycodech:
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
				if dir, ok := KeyCodeToDir[keycode]; ok {
					paused = false
					game.snake.Move(dir)
					if game.clientSnake.IsTaken(game.snake.head.pos) ||
						game.borders.IsTaken(game.snake.head.pos) ||
						game.snake.IsTouchSelf() {
						gameover = true
					} else if game.food.IsTaken(game.snake.head.pos) {
						game.snake.Grow()
						game.food.UpdatePos()
					}
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
				game.clientSnake.Move(dir)
				if game.snake.IsTaken(game.clientSnake.head.pos) ||
					game.borders.IsTaken(game.clientSnake.head.pos) ||
					game.clientSnake.IsTouchSelf() {
					clientGameover = true
				} else if game.food.IsTaken(game.clientSnake.head.pos) {
					game.clientSnake.Grow()
					game.food.UpdatePos()
				}
			}
		case <-game.autoMoveTicker.C:
			if !connected || paused || quit {
				continue Loop
			}
			if !gameover {
				game.snake.Move(game.snake.GetDir())
			}
			if !clientGameover {
				game.clientSnake.Move(game.clientSnake.GetDir())
			}
			eated := false
			if !gameover {
				if game.clientSnake.IsTaken(game.snake.head.pos) ||
					game.borders.IsTaken(game.snake.head.pos) ||
					game.snake.IsTouchSelf() {
					gameover = true
				}
				if game.food.IsTaken(game.snake.head.pos) {
					game.snake.Grow()
					game.food.UpdatePos()
					eated = true
				}
			}
			if !clientGameover {
				if game.snake.IsTaken(game.clientSnake.head.pos) ||
					game.borders.IsTaken(game.clientSnake.head.pos) ||
					game.clientSnake.IsTouchSelf() {
					clientGameover = true
				}
				if !eated && game.food.IsTaken(game.clientSnake.head.pos) {
					game.clientSnake.Grow()
					game.food.UpdatePos()
				}
			}
		case <-game.renderTicker.C:
			if !connected && quit {
				return
			}
			if !connected {
				continue Loop
			}
			result := game.ground.Render(game.snake, game.clientSnake, game.food, game.borders)
			result += "\r==================================================\n"
			result += "\r\033[K\033[3m* Copyright 2022 Steve Zhang. All rights reserved.\033[0m\n"
			result += "\r\033[K\033[3m* p) Pause; r) Replay; q) Quit\033[0m\n"
			result += fmt.Sprintf("\r\033[K\033[3m* Score: 1-%04d | 2-%04d\033[0m\n", game.snake.Len()-1, game.clientSnake.Len()-1)
			p1Status := TreeStr(gameover, "OVER", "OK")
			p2Status := TreeStr(clientGameover, "OVER", "OK")
			result += fmt.Sprintf("\r\033[K\033[3m* State: 1-%-4s | 2-%-4s\033[0m\n\r", p1Status, p2Status)
			gameStatus := TreeStr(paused, "PAUSED", "RUN")
			gameStatus = TreeStr(quit, "QUIT", gameStatus)
			result += fmt.Sprintf("\r\033[K\033[3m* \033[0m\033[3;7m%s\033[0m\n\r", gameStatus)
			renderMsg := encodeRenderMsg(result)
			game.network.Send <- renderMsg
			returnCursor(100)
			player := TreeStr(game.options.Server, "1", "2")
			fmt.Printf("\r\033[K\033[3m* Player: %s\n", player)
			fmt.Print(result)
			if quit {
				return
			}
		}
	}
}

func returnCursor(line int) {
	for i := 0; i < line; i++ {
		fmt.Printf("\033[A")
	}
}

func (game *Game) RunOnlineClient() (err error) {
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
	game.keycodech, err = keys.ListenEvent()
	if err != nil {
		return
	}
	defer keys.StopEventListen()

	// create ping ticker
	pingTicker := time.NewTicker(1 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case keycode := <-game.keycodech:
			if keycode == keys.CodeQuit {
				return
			}
			if dir, ok := KeyCodeToDir[keycode]; ok {
				game.network.Send <- encodeMovMsg(dir)
			}
		case data := <-game.network.Recv:
			msg := decodeMessage(data)
			switch msg.CMD {
			case MSGCMDRender:
				returnCursor(100)
				player := TreeStr(game.options.Server, "1", "2")
				fmt.Printf("\r\033[K\033[3m* Player: %s\n", player)
				result := decodeRenderData(msg.Data)
				fmt.Print(result)
			}
		case <-pingTicker.C:
			game.network.Send <- encodePingMsg()
		}
	}
}
