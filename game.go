package gosnake

import (
	"gosnake/base"
	"net"
	"time"
)

type GameOptions struct {
	GroundWidth         uint
	GroundHeight        uint
	AutoMoveInterval    time.Duration
	ClearPlayerInterval time.Duration
	PlayerSize          uint
}

type GameInput struct {
	From *net.UDPAddr
	Data []byte
}

type GameOutput struct {
	To   *net.UDPAddr
	Data []byte
}

type Game struct {
	limit       base.Limit2D
	players     map[string]*Player
	snakes      *base.Bitmap2D
	walls       *base.Bitmap2D
	foods       *base.Bitmap2D
	foodsCount  int
	autoTicker  *time.Ticker
	clearTicker *time.Ticker
}

func NewGame(options *GameOptions) *Game {
	game := &Game{}

	// limit
	game.limit = base.Limit2D{
		Minx: 1, Maxx: options.GroundWidth - 2,
		Miny: 1, Maxy: options.GroundHeight - 2,
	}

	// players
	game.players = make(map[string]*Player, options.PlayerSize)

	// snakes
	game.snakes = &base.Bitmap2D{}

	// foods
	game.foods = &base.Bitmap2D{}
	game.setFood()

	// walls
	game.walls = &base.Bitmap2D{}
	game.setWalls(options.GroundWidth, options.GroundHeight)

	// tickers
	game.autoTicker = time.NewTicker(options.AutoMoveInterval)
	game.clearTicker = time.NewTicker(options.ClearPlayerInterval)

	return game
}

func (game *Game) Start(inputs <-chan *GameInput) <-chan *GameOutput {
	outputs := make(chan *GameOutput, 1)
	go func() {
		defer close(outputs)
		for {
			select {
			case input, ok := <-inputs:
				if !ok {
					return
				}
				game.handleInput(input, outputs)
			case <-game.autoTicker.C:
				game.autoMovePlayer(outputs)
			case <-game.clearTicker.C:
				game.clearDisconnectedPlayers(outputs)
			}
		}
	}()
	return outputs
}

func (game *Game) getPlayerID(addr *net.UDPAddr) string {
	return addr.String()
}

func (game *Game) handleInput(input *GameInput, outputs chan<- *GameOutput) {
	cmd, encData := DetachPlayerCMD(input.Data)
	switch cmd {
	case CMDPing:
		game.pongPlayer(input.From, encData, outputs)
	case CMDJoin:
		game.joinNewPlayerIfNotExist(input.From, outputs)
	case CMDPause:
		game.pausePlayer(input.From, outputs)
	case CMDReplay:
		game.replayPlayer(input.From, outputs)
	case CMDQuit:
		game.quitPlayer(input.From, outputs)
	default:
		if dir := GetMoveCMDDir(cmd); dir != 0 {
			game.movePlayer(input.From, dir, outputs)
		}
	}
}

func (game *Game) pongPlayer(addr *net.UDPAddr, encPingData []byte, outputs chan<- *GameOutput) {
	playerID := game.getPlayerID(addr)
	if player, ok := game.players[playerID]; ok {
		player.UpdateLastPingedAt()
	}
	pingData := &PingData{}
	DecodeData(encPingData, pingData)
	pongData := &PongData{
		PingedAddr:       addr,
		PingedAtUnixNano: pingData.PingedAtUnixNano,
		PongedAtUnixNano: uint64(time.Now().UnixNano()),
	}
	encPongData := AttachGameCMD(CMDPong, EncodeData(pongData))
	outputs <- &GameOutput{
		To:   addr,
		Data: encPongData,
	}
}

func (game *Game) joinNewPlayerIfNotExist(addr *net.UDPAddr, outputs chan<- *GameOutput) {
	playerID := game.getPlayerID(addr)
	player := game.players[playerID]
	if player != nil {
		return
	}
	player = NewPlayer(addr)
	game.players[playerID] = player
	player.SetSnake(game.limit.GetCenter(), base.GetRandomDir2D())
	game.snakes.Stack(player.snake.GetBitmap())
	game.broadcastGameData(outputs)
}

func (game *Game) pausePlayer(addr *net.UDPAddr, outputs chan<- *GameOutput) {
	playerID := game.getPlayerID(addr)
	player := game.players[playerID]
	if player == nil {
		return
	}
	if player.status.Or(PlayerStatusOver | PlayerStatusPause) {
		return
	}
	player.status.Set(PlayerStatusPause)
	game.broadcastGameData(outputs)
}

func (game *Game) replayPlayer(addr *net.UDPAddr, outputs chan<- *GameOutput) {
	playerID := game.getPlayerID(addr)
	player := game.players[playerID]
	if player == nil {
		return
	}
	if !player.status.Is(PlayerStatusOver) {
		return
	}
	player.status.UnSet(PlayerStatusOver)
	player.SetSnake(game.limit.GetCenter(), base.GetRandomDir2D())
	player.score = 0
	game.snakes.Stack(player.GetSnakeBitmap())
	game.broadcastGameData(outputs)
}

func (game *Game) quitPlayer(addr *net.UDPAddr, outputs chan<- *GameOutput) {
	playerID := game.getPlayerID(addr)
	player := game.players[playerID]
	if player == nil {
		return
	}
	game.snakes.Cull(player.GetSnakeBitmap())
	delete(game.players, playerID)
	game.broadcastGameData(outputs)
}

func (game *Game) movePlayer(addr *net.UDPAddr, dir base.Direction2D, outputs chan<- *GameOutput) {
	playerID := game.getPlayerID(addr)
	player := game.players[playerID]
	if player == nil {
		return
	}
	if player.status.Is(PlayerStatusOver) {
		return
	}
	player.status.UnSet(PlayerStatusPause)
	game.movePlayerSnake(player, dir)
	game.setFood()
	game.broadcastGameData(outputs)
}

func (game *Game) autoMovePlayer(outputs chan<- *GameOutput) {
	for _, player := range game.players {
		if !player.status.Or(PlayerStatusPause | PlayerStatusOver) {
			game.movePlayerSnake(player, player.GetSnakeDir())
		}
	}
	game.setFood()
	game.broadcastGameData(outputs)
}

func (game *Game) clearDisconnectedPlayers(outputs chan<- *GameOutput) {
	for id, player := range game.players {
		if player.lastPingAt.Add(20 * time.Second).Before(time.Now()) {
			game.snakes.Cull(player.GetSnakeBitmap())
			delete(game.players, id)
		}
	}
	game.broadcastGameData(outputs)
}

func (game *Game) setFood() {
	if game.foodsCount == 0 {
		game.foods.Set(game.limit.GetRandom(), true)
		game.foodsCount += 1
	}
}

func (game *Game) setWalls(width, height uint) {
	for x := uint(0); x < width; x++ {
		game.walls.Set(base.Position2D{X: x, Y: 0}, true)
		game.walls.Set(base.Position2D{X: x, Y: height - 1}, true)
	}
	for y := uint(0); y < height; y++ {
		game.walls.Set(base.Position2D{X: 0, Y: y}, true)
		game.walls.Set(base.Position2D{X: width - 1, Y: y}, true)
	}
}

func (game *Game) movePlayerSnake(player *Player, dir base.Direction2D) {
	nextHeadPos := player.GetNextSnakeHeadPos(dir)
	if nextHeadPos == nil {
		return
	}

	// game over
	if game.walls.Get(*nextHeadPos) ||
		(game.snakes.Get(*nextHeadPos) &&
			*nextHeadPos != player.GetSnakeTailPos()) {
		game.foods.Stack(player.GetSnakeBitmap())
		game.foodsCount += player.GetSnakeLen()
		game.snakes.Cull(player.GetSnakeBitmap())
		player.UnsetSnake()
		player.status.Set(PlayerStatusOver)
		return
	}

	// update snakes bitmap
	game.snakes.Cull(player.GetSnakeBitmap())
	defer func() {
		game.snakes.Stack(player.GetSnakeBitmap())
	}()

	// move
	player.MoveSnake(dir)

	// eat food
	if game.foods.Get(player.GetSnakeHeadPos()) {
		game.foods.Set(player.GetSnakeHeadPos(), false)
		game.foodsCount -= 1
		player.GrowSnake()
		player.IncreaseScore()
	}
}

func (game *Game) broadcastGameData(outputs chan<- *GameOutput) {
	gameData := game.getGameData()
	enc := EncodeData(gameData)
	enc = AttachGameCMD(CMDUpdate, enc)
	for _, player := range game.players {
		outputs <- &GameOutput{
			To:   player.GetAddr(),
			Data: enc,
		}
	}
}
