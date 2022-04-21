package gosnake

import (
	"fmt"
	"gosnake/keys"
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
	fmt.Println("Create connect...")
	game.network = NewNetWork()
	err = game.network.Start(
		game.options.LocalIP,
		game.options.LocalPort,
		game.options.DialIP,
		game.options.DialPort,
	)
	if err != nil {
		return
	}
	defer game.network.Stop()

	fmt.Println("Waiting for user...")
	for done := false; !done; {
		msgData := <-game.network.Recv
		msg := decodeMessage(msgData)
		if msg.CMD == MSGCMDPing {
			done = true
		}
	}

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

	// close the cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	var (
		gameover       bool
		clientGameover bool
		paused         bool
	)

	for {
		select {
		case keycode := <-game.keycodech:
			if keycode == keys.CodePause {
				paused = true
			} else if !gameover {
				if keycode == keys.CodeQuit {
					return
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
			msg := decodeMessage(data)
			switch msg.CMD {
			case MSGCMDMov:
				if !paused && !clientGameover {
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
			}
		case <-game.autoMoveTicker.C:
			if !paused {
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
			}
		case <-game.renderTicker.C:
			result := game.ground.Render(game.snake, game.clientSnake, game.food, game.borders)
			stateP1 := TreeStr(gameover, "OVER", "OK")
			stateP2 := TreeStr(clientGameover, "OVER", "OK")
			result = fmt.Sprintf("%s\r\033[3m* player1: %s, player2: %s\033[0m\n\r", result, stateP1, stateP2)
			renderMsg := encodeRenderMsg(result)
			game.network.Send <- renderMsg
			fmt.Print(result)
			if gameover && clientGameover {
				return
			}
		default:
		}
	}
}

func (game *Game) RunOnlineClient() (err error) {
	fmt.Println("Create connect...")
	game.network = NewNetWork()
	err = game.network.Start(
		game.options.LocalIP,
		game.options.LocalPort,
		game.options.DialIP,
		game.options.DialPort,
	)
	if err != nil {
		return
	}
	defer game.network.Stop()

	// listen keys event
	game.keycodech, err = keys.ListenEvent()
	if err != nil {
		return
	}
	defer keys.StopEventListen()

	// create ping ticker
	pingTicker := time.NewTicker(1 * time.Second)
	defer pingTicker.Stop()

	// close the cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")

	fmt.Println("start")

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
				result := decodeRenderData(msg.Data)
				fmt.Print(result)
			}
		case <-pingTicker.C:
			game.network.Send <- encodePingMsg()
		default:
		}
	}
}
