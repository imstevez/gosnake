package gosnake

import (
	"context"
	"fmt"
	"gosnake/keys"
	"os"
	"os/exec"
	"time"
)

var DefaultClientOptions = &ClientOptions{
	PingInterval: 1 * time.Second,
	ServerAddr:   "127.0.0.1:9001",
	FPS:          30,
	RenderConfig: &RenderConfig{
		SnakesSymbol:      "\033[41;1;37m[]\033[0m",
		PlayerSnakeSymbol: "\033[41;1;37m[]\033[0m",
		WallsSymbol:       "\033[44;1;37m[]\033[0m",
		FoodsSymbol:       "\033[42;1;37m[]\033[0m",
		GroundSymbol:      "  ",
		PlayerStatColor:   "\033[41;1;37m",
		StatsColor:        "\033[0m",
	},
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
	PingInterval time.Duration
	ServerAddr   string
	FPS          int
	RenderConfig *RenderConfig
}

type Client struct {
	options      *ClientOptions
	network      *Network
	pingTicker   *time.Ticker
	renderTicker *time.Ticker
	keyEvents    <-chan keys.Code
	clearFuncs   []func()
	cancel       context.CancelFunc
	pingMS       uint64
	joined       bool
	updated      bool
	pongData     *PongData
	gameData     *GameData
}

func NewClient(options *ClientOptions) (client *Client, err error) {
	client = &Client{options: options}
	client.pingTicker = time.NewTicker(options.PingInterval)
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
	client.pongData = &PongData{}
	client.gameData = &GameData{}
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
			client.ping()
		case data := <-client.network.Recv:
			client.handleRecv(data)
		case <-client.renderTicker.C:
			client.render()
		}

	}
}

func (client *Client) handleRecv(data []byte) {
	client.updated = true
	cmd, data := DetachGameCMD(data)
	switch cmd {
	case CMDPong:
		DecodeData(data, client.pongData)
		client.pingMS = (uint64(time.Now().UnixNano()) - client.pongData.PingedAtUnixNano) / 1e6
	case CMDUpdate:
		client.joined = true
		DecodeData(data, client.gameData)
	}
}

func (client *Client) handleKeycode(keycode keys.Code) {
	cmd := GetKeyCodeCMD(keycode)
	if cmd == 0 {
		return
	}
	data := AttachPlayerCMD(cmd, nil)
	client.network.Send <- data
	if cmd == CMDQuit {
		time.Sleep(500 * time.Millisecond)
		client.cancel()
	}
}

func (client *Client) ping() {
	pingData := &PingData{
		PingedAtUnixNano: uint64(time.Now().UnixNano()),
	}
	data := EncodeData(pingData)
	data = AttachPlayerCMD(CMDPing, data)
	client.network.Send <- data
	if !client.joined {
		data := AttachPlayerCMD(CMDJoin, nil)
		client.network.Send <- data
	}
}

func (client *Client) render() {
	if !client.joined {
		fmt.Print("\033[2A\rWaiting for join game...\033[K\n")
		fmt.Printf("\rPing: %dms\033[K\n", client.pingMS)
		return
	}
	if client.updated {
		client.updated = false
		frame := client.gameData.Render(
			client.pongData.PingedAddr,
			client.options.RenderConfig,
			client.pingMS,
		)
		fmt.Print(frame)
	}
}

func (client *Client) clear() {
	for i := len(client.clearFuncs) - 1; i >= 0; i-- {
		client.clearFuncs[i]()
	}
}
