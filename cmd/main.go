package main

import (
	"flag"
	"fmt"
	"gosnake"
	"os"
)

var options = gosnake.GameOptions{
	GroundWith:          50,
	GroundHeight:        30,
	GroundSymbol:        " ",
	BordersWidth:        50,
	BordersHeight:       30,
	BordersSymbol:       "\033[46;30m \033[0m",
	SnakeInitPosX:       14,
	SnakeInitPosY:       7,
	SnakeInitDir:        gosnake.DirLeft,
	SnakeSymbol:         "\033[44;30m \033[0m",
	ClientSnakeInitPosX: 34,
	ClientSnakeInitPosY: 14,
	ClientSnakeInitDir:  gosnake.DirRight,
	ClientSnakeSymbol:   "\033[41;30m \033[0m",
	SnakeSpeedMS:        300,
	FoodSymbol:          "\033[42;30m \033[0m",
	Online:              false,
	Server:              false,
	LocalAddr:           "",
	RemoteAddr:          "",
}

func init() {
	flag.BoolVar(&(options.Online), "online", false, "play online")
	flag.BoolVar(&(options.Server), "s", false, "play as a server")
	flag.StringVar(&(options.LocalAddr), "l", "0.0.0.0:10001", "local addr")
	flag.StringVar(&(options.RemoteAddr), "r", "0.0.0.0:10002", "local ip")
	flag.Int64Var(&(options.SnakeSpeedMS), "sp", 300, "snake auto move speed (millsecond per mov)")
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
