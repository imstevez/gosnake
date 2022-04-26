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
	PingIntervalMs:    1000,
	ServerAddr:        "127.0.0.1:9001",
	RoomID:            0,
	SnakeSymbol:       "\033[41;1;37m[]\033[0m",
	PlayerSnakeSymbol: "\033[44;1;37m[]\033[0m",
	BorderSymbol:      "\033[46;1;37m[]\033[0m",
	FoodSymbol:        "\033[42;1;37m[]\033[0m",
	GroundSymbol:      "  ",
	FPS:               30,
}

func RunClient(ctx context.Context) error {
	client, err := NewClient(DefaultClientOptions)
	if err != nil {
		return err
	}
	client.Run(ctx)
	return nil
}

type ClientOptions struct {
	PingIntervalMs    int
	ServerAddr        string
	RoomID            int
	SnakeSymbol       string
	PlayerSnakeSymbol string
	FoodSymbol        string
	BorderSymbol      string
	GroundSymbol      string
	FPS               int
}

type Client struct {
	options      *ClientOptions
	network      *Network
	pingTicker   *time.Ticker
	renderTicker *time.Ticker
	keyEvents    <-chan keys.Code
	clearFuncs   []func()
	texts        Lines
	ground       *Ground
	border       *RecBorder
	once         *sync.Once
	cancel       context.CancelFunc
	frame        string
}

func NewClient(options *ClientOptions) (client *Client, err error) {
	client = &Client{options: options, once: &sync.Once{}}
	client.frame = "\rWaiting for server response...\033[K"
	client.pingTicker = time.NewTicker(
		time.Duration(options.PingIntervalMs) * time.Millisecond,
	)
	client.clearFuncs = append(
		client.clearFuncs, client.pingTicker.Stop,
	)
	client.renderTicker = time.NewTicker(
		time.Duration(1000/options.FPS) * time.Millisecond,
	)
	client.clearFuncs = append(
		client.clearFuncs, client.renderTicker.Stop,
	)
	client.network = NewNetWork(
		splitChildPackageSize, splitChildPackageNum,
	)
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
		"************************ GOSNAKE@v0.0.1 ************************",
		"****************************************************************",
		" * Up: w,i   Left: a,j  Down: s,k  Right: d,j",
		" * Pause: p  Replay: r  Quit: q",
		"----------------------------------------------------------------",
		" * rank   players                   score   state               ",
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
			client.update(data)
		case <-client.renderTicker.C:
			client.render()
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

func (client *Client) update(data []byte) {
	sceneData := DecodeSceneData(data)
	client.once.Do(func() {
		client.ground = NewGround(sceneData.BorderWidth, sceneData.BorderHeight, client.options.GroundSymbol)
		client.border = NewRecBorder(sceneData.BorderWidth, sceneData.BorderHeight, client.options.BorderSymbol)
	})
	sceneData.Food.SetSymbol(client.options.FoodSymbol)
	sceneData.Snakes.SetSymbol(client.options.SnakeSymbol)
	sceneData.PlayerSnake.SetSymbol(client.options.PlayerSnakeSymbol)
	layers := []Layer{client.border, sceneData.Food, sceneData.Snakes, sceneData.PlayerSnake}
	texts := client.getPlayerStatsTexts(sceneData.PlayerID, sceneData.PlayerStats)
	texts = append(client.texts[:], texts...)
	client.frame = client.ground.Render(layers...).PreAppend(
		texts[:1],
	).Append(
		texts[1:],
	).Merge()
}

func (client *Client) getPlayerStatsTexts(playerID string, stats PlayerStats) (texts Lines) {
	sort.Sort(stats)
	for i, stat := range stats {
		color := ""
		if playerID == stat.ID {
			color = "1;44;37"
		}
		state := client.getStateStr(stat.Pause, stat.Over)
		line := fmt.Sprintf(
			"  \033[%sm %d      %-21s     %03d     %-5s  \033[0m",
			color, i+1, stat.ID, stat.Score, state,
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

func (client *Client) render() {
	fmt.Print(client.frame)
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
