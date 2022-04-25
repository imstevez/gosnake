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
}

func NewClient(options *ClientOptions) (client *Client, err error) {
	client = &Client{options: options}
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
		"                    \033[3;1mGoSnake@v0.0.1\033[0m                    ",
		" =====================================================",
		"",
		"",
		" \033[3m* Keys Map\033[0m",
		"   w,i) Up    a,j) Left   s,k) Down   d,j) Right      ",
		"     p) Pause   r) Replay   q) Quit                   ",
		"",
		"",
		" \033[3m* layers Stat\033[0m",
		"   Rank    Players                   Score   State    ",
	}
	return
}

func (client *Client) Run(ctx context.Context) {
	defer client.clear()
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h\n\r")
	// clear screen
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
	fmt.Println("\rWaiting for server response...")
	for {
		select {
		case <-ctx.Done():
			return
		case keycode := <-client.keyEvents:
			switch keycode {
			case keys.CodeQuit:
				client.sendCMD(CMDQuit)
				time.Sleep(500 * time.Millisecond)
				return
			case keys.CodePause:
				client.sendCMD(CMDPause)
			case keys.CodeReplay:
				client.sendCMD(CMDReplay)
			case keys.CodeUp, keys.CodeUp2:
				client.sendCMD(CMDMovUp)
			case keys.CodeDown, keys.CodeDown2:
				client.sendCMD(CMDMovDown)
			case keys.CodeLeft, keys.CodeLeft2:
				client.sendCMD(CMDMovLeft)
			case keys.CodeRight, keys.CodeRight2:
				client.sendCMD(CMDMovRight)
			}
		case <-client.pingTicker.C:
			client.sendCMD(CMDPing)
		case data := <-client.network.Recv:
			if len(data) == 0 {
				continue
			}
			var layers []Layer
			texts := client.texts[:]
			sceneData, err := client.decodeSceneData(data)
			if err == nil {
				ground := NewGround(
					sceneData.Options.GroundWith,
					sceneData.Options.GroundHeight,
					sceneData.Options.GroundSymbol,
				)
				border := NewRecBorder(
					sceneData.Options.BorderWidth,
					sceneData.Options.GroundHeight,
					sceneData.Options.BorderSymbol,
				)
				food := NewCommonLayer(
					map[Position]struct{}{sceneData.FoodPos: {}},
					sceneData.Options.FoodSymbol,
				)
				layers = append(layers, border, food)
				sort.Sort(sceneData.Players)
				for i, player := range sceneData.Players {
					snakeSymbol := sceneData.Options.PlayerOptions.SnakeSymbol
					color := "0"
					if sceneData.PlayerID == player.ID {
						snakeSymbol = mySnakeSymbol
						color = "1;34"
					}
					layers = append(layers, NewCommonLayer(player.SnakeTakes, snakeSymbol))
					state := IfStr(player.Pause, "Pause", "Run")
					state = IfStr(player.Over, "Over", state)
					line := fmt.Sprintf(
						" \033[%sm  %d       %-21s     %03d     %-5s    \033[0m",
						color, i+1, player.ID, player.Score,
						state,
					)
					texts = append(texts, line)
				}
				frame := ground.Render(layers...).HozJoin(
					texts,
					sceneData.Options.GroundWith*len(ground.symbol),
				).Merge()
				fmt.Print(frame)
			}
		}
	}
}

func (client *Client) decodeSceneData(data []byte) (*GameSceneData, error) {
	var sceneData GameSceneData
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&sceneData)
	return &sceneData, err
}

func (client *Client) sendCMD(cmd string) {
	data := client.encodeClientData(cmd)
	client.network.Send <- data
}

func (client *Client) encodeClientData(cmd string) []byte {
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
