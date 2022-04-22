package main

import (
	"flag"
	"fmt"
	"gosnake"
	"os"
)

var options = gosnake.GameOptions{
	GroundWith:          30,
	GroundHeight:        30,
	GroundSymbol:        "  ",
	BordersWidth:        30,
	BordersHeight:       30,
	BordersSymbol:       "\033[47;37m  \033[0m",
	SnakeInitPosX:       14,
	SnakeInitPosY:       7,
	SnakeInitDir:        gosnake.DirLeft,
	SnakeSymbol:         "\033[44;37m  \033[0m",
	ClientSnakeInitPosX: 7,
	ClientSnakeInitPosY: 14,
	ClientSnakeInitDir:  gosnake.DirRight,
	ClientSnakeSymbol:   "\033[41;37m  \033[0m",
	SnakeSpeedMS:        300,
	FoodSymbol:          "\033[41;37m  \033[0m",
	Online:              false,
	Server:              false,
	LocalAddr:           "",
	RemoteAddr:          "",
	FPS:                 30,
}

func init() {
	flag.BoolVar(&(options.Online), "online", false, "play online")
	flag.BoolVar(&(options.Server), "s", false, "play as a server")
	flag.StringVar(&(options.LocalAddr), "l", "0.0.0.0:10001", "local addr")
	flag.StringVar(&(options.RemoteAddr), "r", "0.0.0.0:10002", "local ip")
	flag.Int64Var(&(options.SnakeSpeedMS), "sp", 200, "snake auto move speed")
}

func main() {
	flag.Parse()
	game, err := gosnake.NewGame(&options)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := game.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
