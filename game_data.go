package gosnake

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"gosnake/base"
	"gosnake/helper"
	"math/bits"
	"net"
	"sort"
	"strings"
)

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
		snakes.Stack(item.Snake)
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

	h := uint(len(*gameData.Walls))
	w := uint(len((*gameData.Walls)[0])) * bits.UintSize

	for y := uint(0); y < h; y++ {
		line := "\r"
		for x := uint(0); x < w; x++ {
			pos := base.Position2D{X: x, Y: y}
			symbol := renderConfig.GroundSymbol
			if gameData.Foods.Get(pos) {
				symbol = renderConfig.FoodsSymbol
			}
			if snakes.Get(pos) {
				symbol = renderConfig.SnakesSymbol
			}
			if playerSnake.Get(pos) {
				symbol = renderConfig.PlayerSnakeSymbol
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

func AttachPlayerCMD(cmd PlayerCMD, data []byte) (aData []byte) {
	aData = make([]byte, 2)
	binary.BigEndian.PutUint16(aData, uint16(cmd))
	aData = append(aData, data...)
	return
}

func DetachPlayerCMD(aData []byte) (cmd PlayerCMD, data []byte) {
	cmd = PlayerCMD(binary.BigEndian.Uint16(aData[:2]))
	data = aData[2:]
	return
}

func AttachGameCMD(cmd GameCMD, data []byte) (aData []byte) {
	aData = make([]byte, 2)
	binary.BigEndian.PutUint16(aData, uint16(cmd))
	aData = append(aData, data...)
	return
}

func DetachGameCMD(aData []byte) (cmd GameCMD, data []byte) {
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
