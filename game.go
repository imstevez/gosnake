package gosnake

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"gosnake/base"
	"gosnake/helper"
	"net"
	"sort"
	"strings"
	"time"
)

type GameOptions struct {
	GroundWidth         int
	GroundHeight        int
	AutoMoveInterval    time.Duration
	ClearPlayerInterval time.Duration
	PlayerSize          int
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
	game.foods.Set(game.limit.GetRandom(), true)
	game.foodsCount = 1

	// walls
	game.walls = &base.Bitmap2D{}
	for x := 0; x < options.GroundWidth; x++ {
		game.walls.Set(base.Position2D{x, 0}, true)
		game.walls.Set(base.Position2D{x, options.GroundHeight - 1}, true)
	}
	for y := 0; y < options.GroundHeight; y++ {
		game.walls.Set(base.Position2D{0, y}, true)
		game.walls.Set(base.Position2D{options.GroundWidth - 1, y}, true)
	}

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
	cmd, encData := DetatchPlayerCMD(input.Data)
	fmt.Println("CMD: ", cmd)
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
	encPongData := AttatchGameCMD(CMDPong, EncodeData(pongData))
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
	game.snakes.Add(player.snake.GetBitmap())
	game.broadcastGameData(outputs)
}

func (game *Game) pausePlayer(addr *net.UDPAddr, outputs chan<- *GameOutput) {
	playerID := game.getPlayerID(addr)
	player := game.players[playerID]
	if player == nil {
		return
	}
	if player.status.Is(PlayerStatusOver) ||
		player.status.Is(PlayerStatusPause) {
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
	game.snakes.Add(player.GetSnakeBitmap())
	if game.foodsCount == 0 {
		game.setRandomFood()
	}
	game.broadcastGameData(outputs)
}

func (game *Game) quitPlayer(addr *net.UDPAddr, outputs chan<- *GameOutput) {
	playerID := game.getPlayerID(addr)
	player := game.players[playerID]
	if player == nil {
		return
	}
	delete(game.players, playerID)
	game.broadcastGameData(outputs)
}

func (game *Game) movePlayer(addr *net.UDPAddr, dir base.Direction2D, outputs chan<- *GameOutput) {
	playerID := game.getPlayerID(addr)
	player := game.players[playerID]
	if player == nil {
		return
	}
	game.movePlayerSnake(player, dir)
	game.broadcastGameData(outputs)
}

func (game *Game) autoMovePlayer(outputs chan<- *GameOutput) {
	for _, player := range game.players {
		if player.status.Is(PlayerStatusPause) ||
			player.status.Is(PlayerStatusOver) ||
			player.snake == nil {
			continue
		}
		game.movePlayerSnake(player, player.GetSnakeDir())
	}
	if game.foodsCount == 0 {
		game.setRandomFood()
	}
	game.broadcastGameData(outputs)
}

func (game *Game) clearDisconnectedPlayers(outputs chan<- *GameOutput) {
	for id, player := range game.players {
		if player.lastPingAt.Add(20 * time.Second).Before(time.Now()) {
			game.snakes.Minus(player.GetSnakeBitmap())
			delete(game.players, id)
		}
	}
	game.broadcastGameData(outputs)
}

func (game *Game) setRandomFood() {
	game.foods.Set(game.limit.GetRandom(), true)
	game.foodsCount += 1
}

func (game *Game) movePlayerSnake(player *Player, dir base.Direction2D) {
	defer func() {
		fmt.Println(player.GetSnakeBitmap())
	}()
	if player.status.Is(PlayerStatusOver) || player.snake == nil {
		return
	}
	player.status.UnSet(PlayerStatusPause)
	nextHeadPos := player.GetNextSnakeHeadPos(dir)
	if nextHeadPos == nil {
		return
	}

	// game over
	if game.walls.Get(*nextHeadPos) ||
		(game.snakes.Get(*nextHeadPos) &&
			*nextHeadPos != player.GetSnakeTailPos()) {
		game.foods.Add(player.GetSnakeBitmap())
		game.foodsCount += player.GetSnakeLen()
		game.snakes.Minus(player.GetSnakeBitmap())
		player.UnsetSnake()
		player.status.Set(PlayerStatusOver)
		return
	}

	// update snakes bitmap
	game.snakes.Minus(player.GetSnakeBitmap())
	defer func() {
		game.snakes.Add(player.GetSnakeBitmap())
	}()

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
	enc = AttatchGameCMD(CMDUpdate, enc)
	for _, player := range game.players {
		outputs <- &GameOutput{
			To:   player.GetAddr(),
			Data: enc,
		}
	}
}

type GameData struct {
	Walls       *base.Bitmap2D
	Foods       *base.Bitmap2D
	PlayersData PlayersData
}

type RenderConfig struct {
	SnakesSymbol      string
	PlayerSnakeSymbol string
	WallsSymbol       string
	FoodsSymbol       string
	GroundSymbol      string
	PlayerStatColor   string
	StatsColor        string
}

func (gameData *GameData) Render(playerAddr *net.UDPAddr, renderConfig *RenderConfig, pingMS uint64) string {
	ground := []string{}
	stats := []string{}
	snakes := &base.Bitmap2D{}
	playerSnake := &base.Bitmap2D{}

	sort.Sort(gameData.PlayersData)

	addrStr := playerAddr.String()

	for i, item := range gameData.PlayersData {
		snakes.Add(item.Snake)
		color := renderConfig.StatsColor
		if item.Addr.String() == addrStr {
			playerSnake = item.Snake
			color = renderConfig.PlayerStatColor
		}
		status := helper.IfStr(item.Status.Is(PlayerStatusOver), "Over ", "Run  ")
		status = helper.IfStr(item.Status.Is(PlayerStatusPause), "Pause", status)
		line := fmt.Sprintf(
			"\r%s%d\t%s\t%d\t%s\033[0m",
			color, i, item.Addr.String(),
			item.Score, status,
		)
		stats = append(stats, line)
	}

	h := len(*gameData.Walls)
	w := len((*gameData.Walls)[0]) * base.BitsPerByte

	for y := 0; y < h; y++ {
		line := "\r"
		for x := 0; x < w; x++ {
			pos := base.Position2D{X: x, Y: y}
			symbol := renderConfig.GroundSymbol
			if snakes.Get(pos) {
				symbol = renderConfig.SnakesSymbol
			}
			if playerSnake.Get(pos) {
				symbol = renderConfig.PlayerSnakeSymbol
			}
			if gameData.Foods.Get(pos) {
				symbol = renderConfig.FoodsSymbol
			}
			if gameData.Walls.Get(pos) {
				symbol = renderConfig.WallsSymbol
			}
			line += symbol
		}
		ground = append(ground, line)
	}
	frame := fmt.Sprintf("\033[%dA", len(ground)+len(stats)+1)
	frame += strings.Join(ground, "\033[K\n")
	frame += "\033[K\n"
	frame += strings.Join(stats, "\033[K\n")
	frame += fmt.Sprintf("\n\rPing: %dms\033[K\n", pingMS)
	return frame
}

func (game *Game) getGameData() *GameData {
	data := &GameData{
		Walls:       game.walls,
		Foods:       game.foods,
		PlayersData: make(PlayersData, len(game.players)),
	}
	i := 0
	for _, player := range game.players {
		data.PlayersData[i] = player.GetPlayerData()
		i++
	}
	return data
}

func AttatchPlayerCMD(cmd PlayerCMD, data []byte) (aData []byte) {
	aData = make([]byte, 2)
	binary.BigEndian.PutUint16(aData, uint16(cmd))
	aData = append(aData, data...)
	return
}

func DetatchPlayerCMD(aData []byte) (cmd PlayerCMD, data []byte) {
	cmd = PlayerCMD(binary.BigEndian.Uint16(aData[:2]))
	data = aData[2:]
	return
}

func AttatchGameCMD(cmd GameCMD, data []byte) (aData []byte) {
	aData = make([]byte, 2)
	binary.BigEndian.PutUint16(aData, uint16(cmd))
	aData = append(aData, data...)
	return
}

func DetatchGameCMD(aData []byte) (cmd GameCMD, data []byte) {
	cmd = GameCMD(binary.BigEndian.Uint16(aData[:2]))
	data = aData[2:]
	return
}

type PingData struct {
	PingedAtUnixNano uint64
}

type PongData struct {
	PingedAddr       *net.UDPAddr
	PingedAtUnixNano uint64
	PongedAtUnixNano uint64
}

func EncodeData(data interface{}) []byte {
	enc, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return enc
}

func DecodeData(enc []byte, data interface{}) {
	err := json.Unmarshal(enc, data)
	if err != nil {
		panic(err)
	}
}
