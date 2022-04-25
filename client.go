package gosnake

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"gosnake/keys"
	"os"
	"os/exec"
	"sort"
	"sync"
	"time"
)

var DefaultClientOptions = &ClientOptions{
	PingIntervalMs: 1000,
	ServerAddr:     "127.0.0.1:9001",
	RoomID:         0,
}

const mySnakeSymbol = "\033[44;1;37m[]\033[0m"

func RunClient(ctx context.Context) error {
	client, err := NewClient(DefaultClientOptions)
	if err != nil {
		return err
	}
	client.Run(ctx)
	return nil
}

type ClientOptions struct {
	PingIntervalMs int
	ServerAddr     string
	RoomID         int
}

type Client struct {
	options    *ClientOptions
	network    *Network
	pingTicker *time.Ticker
	keyEvents  <-chan keys.Code
	clearFuncs []func()
	texts      Lines
	ground     *Ground
	border     *RecBorder
	once       *sync.Once
	cancel     context.CancelFunc
}

func NewClient(options *ClientOptions) (client *Client, err error) {
	client = &Client{options: options, once: &sync.Once{}}
	client.pingTicker = time.NewTicker(
		time.Duration(options.PingIntervalMs) * time.Millisecond,
	)
	client.clearFuncs = append(
		client.clearFuncs, client.pingTicker.Stop,
	)
	client.network = NewNetWork()
	err = client.network.Start("", client.options.ServerAddr)
	if err != nil {
		return
	}
	client.clearFuncs = append(
		client.clearFuncs, client.network.Stop,
	)
	client.keyEvents, err = keys.ListenEvent()
	if err != nil {
		return
	}
	client.clearFuncs = append(
		client.clearFuncs, keys.StopEventListen,
	)
	client.texts = []string{
		" =====================================================",
		" ////////////////// GOSNAKE@v0.0.1 ///////////////////",
		" =====================================================",
		"                                                      ",
		" * KEYS:                                              ",
		" -----------------------------------------------------",
		"   w,i) Up    a,j) Left   s,k) Down   d,j) Right      ",
		"     p) Pause   r) Replay   q) Quit                   ",
		"                                                      ",
		" * PLAYERS:                                           ",
		" -----------------------------------------------------",
		"   rank    players                   score   state    ",
	}
	return
}

func (client *Client) Run(ctx context.Context) {
	defer client.clear()

	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h\n\r")
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
	fmt.Println("\rWaiting for server response...")

	ctx, client.cancel = context.WithCancel(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case keycode := <-client.keyEvents:
			client.handleKeycode(keycode)
		case <-client.pingTicker.C:
			client.sendCMD(CMDPing)
		case data := <-client.network.Recv:
			client.render(data)
		}
	}
}

func (client *Client) handleKeycode(keycode keys.Code) {
	cmd := GetKeyCodeCMD(keycode)
	if cmd == "" {
		return
	}
	client.sendCMD(cmd)
	if cmd == CMDQuit {
		time.Sleep(500 * time.Millisecond)
		client.cancel()
	}
}

func (client *Client) render(data []byte) {
	sceneData := client.decodeSceneData(data)
	if sceneData == nil {
		return
	}
	options := sceneData.Options
	client.once.Do(func() {
		client.ground = NewGround(options.GroundWith, options.GroundHeight, options.GroundSymbol)
		client.border = NewRecBorder(options.BorderWidth, options.GroundHeight, options.BorderSymbol)
	})
	food := NewCommonLayer(
		map[Position]struct{}{sceneData.FoodPos: {}},
		options.FoodSymbol,
	)

	playersLayers, playersTexts := client.getPlayersPrint(sceneData)
	layers := append([]Layer{client.border, food}, playersLayers...)
	texts := append(client.texts[:], playersTexts...)
	joinwith := sceneData.Options.GroundWith * len(client.ground.symbol)

	frame := client.ground.Render(layers...).HozJoin(
		texts, joinwith,
	).Merge()

	fmt.Print(frame)
}

func (client *Client) getPlayersPrint(sceneData *GameSceneData) (layers []Layer, texts Lines) {
	sort.Sort(sceneData.Players)
	for i, player := range sceneData.Players {
		snakeSymbol := sceneData.Options.PlayerOptions.SnakeSymbol
		color := "0"
		if sceneData.PlayerID == player.ID {
			snakeSymbol = mySnakeSymbol
			color = "1;34"
		}
		snakeLayer := NewCommonLayer(player.SnakeTakes, snakeSymbol)
		layers = append(layers, snakeLayer)
		state := client.getStateStr(player.Pause, player.Over)
		line := fmt.Sprintf(
			" \033[%sm  %d       %-21s     %03d     %-5s    \033[0m",
			color, i+1, player.ID, player.Score,
			state,
		)
		texts = append(texts, line)
	}
	return
}

func (client *Client) getStateStr(pause, over bool) (state string) {
	state = IfStr(pause, "Pause", "Run")
	state = IfStr(over, "Over", state)
	return
}

func (client *Client) decodeSceneData(data []byte) *GameSceneData {
	if len(data) == 0 {
		return nil
	}
	var sceneData GameSceneData
	buf := bytes.NewBuffer(data)
	gob.NewDecoder(buf).Decode(&sceneData)
	return &sceneData
}

func (client *Client) sendCMD(cmd CMD) {
	data := client.encodeClientData(cmd)
	client.network.Send <- data
}

func (client *Client) encodeClientData(cmd CMD) []byte {
	cliData := &ClientData{
		RoomID: client.options.RoomID,
		CMD:    cmd,
	}
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(cliData)
	if err != nil {
		return nil
	}
	return buf.Bytes()
}

func (client *Client) clear() {
	for i := len(client.clearFuncs) - 1; i >= 0; i-- {
		client.clearFuncs[i]()
	}
}
