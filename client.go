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
	ClientAddr:     "127.0.0.1:9002",
	RoomID:         0,
}

const mySnakeSymbol = "\033[44;37m[]\033[0m"

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
	ClientAddr     string
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
		" \033[3m ===================================================\033[0m",
		" \033[3m >>>GoSnake@v0.0.1\033[0m",
		" \033[3m * Copyright 2022 Steve Zhang. All rights reserved.\033[0m",
		" \033[3m ===================================================\033[0m",
		"",
		" \033[3m >>>Keys\033[0m",
		" \033[3m * w,i)Up  | a,j)Left | s,k)Down | d,j)Right\033[0m",
		" \033[3m * p)Pause | r)Replay | q)Quit\033[0m",
		"",
		" \033[3m >>>Players\033[0m",
		"   +---------------------------+-------+-------+",
		"   | Player                    | Score | State |",
		"   +---------------------------+-------+-------+",
	}
	return
}

func (client *Client) Run(ctx context.Context) {
	defer client.clear()
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h\r")
	// clear screen
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
	fmt.Println("\rWaiting for server...")
	for {
		select {
		case <-ctx.Done():
			return
		case keycode := <-client.keyEvents:
			switch keycode {
			case keys.CodeQuit:
				client.SendCMD(CMDQuit)
				time.Sleep(1 * time.Second)
				return
			case keys.CodePause:
				client.SendCMD(CMDPause)
			case keys.CodeReplay:
				client.SendCMD(CMDReplay)
			case keys.CodeUp, keys.CodeUp2:
				client.SendCMD(CMDMovUp)
			case keys.CodeDown, keys.CodeDown2:
				client.SendCMD(CMDMovDown)
			case keys.CodeLeft, keys.CodeLeft2:
				client.SendCMD(CMDMovLeft)
			case keys.CodeRight, keys.CodeRight2:
				client.SendCMD(CMDMovRight)
			}
		case <-client.pingTicker.C:
			client.SendCMD(CMDPing)
		case data := <-client.network.Recv:
			if len(data) == 0 {
				continue
			}
			texts := client.texts[:]
			var layers []Layer
			gameData, err := client.DecodeGameData(data)
			if err == nil {
				ground := NewGround(gameData.Options.GroundWith, gameData.Options.GroundHeight, gameData.Options.GroundSymbol)
				border := NewRecBorder(gameData.Options.BorderWidth, gameData.Options.GroundHeight, gameData.Options.BorderSymbol)
				food := NewCommonLayer(map[Position]struct{}{gameData.FoodPos: struct{}{}}, gameData.Options.FoodSymbol)
				layers = append(layers, border, food)
				sort.Sort(gameData.Players)
				for _, p := range gameData.Players {
					colors := ""
					colore := ""
					snakeSymbol := gameData.Options.PlayerOptions.SnakeSymbol
					if p.Name == client.network.GetLocalAddr() {
						snakeSymbol = mySnakeSymbol
						colors = "\033[44m"
						colore = "\033[0m"
					}
					layers = append(layers, NewCommonLayer(p.SnakeTakes, snakeSymbol))
					state := IfStr(p.Pause, "Pause", "Run")
					state = IfStr(p.Over, "Over", state)

					line := fmt.Sprintf("   %s| %-25s | %03d   | %-4s  |%s", colors, p.Name, p.Score, state, colore)
					hr := "   +---------------------------+-------+-------+"
					texts = append(texts, line, hr)
				}
				frame := ground.Render(layers...).HozJoin(texts, gameData.Options.GroundWith*2).Merge()
				fmt.Print(frame)
			}
		}
	}
}

func (client *Client) DecodeGameData(data []byte) (*GameData, error) {
	var gameData GameData
	buf := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&gameData)
	return &gameData, err
}

func (client *Client) SendCMD(cmd string) {
	data := client.EncodeData(cmd)
	client.network.Send <- data
}

func (client *Client) EncodeData(cmd string) []byte {
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
