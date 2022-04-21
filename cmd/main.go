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
	SnakeSymbol:         "\033[41;30mS\033[0m",
	ClientSnakeInitPosX: 34,
	ClientSnakeInitPosY: 14,
	ClientSnakeInitDir:  gosnake.DirRight,
	ClientSnakeSymbol:   "\033[41;30mC\033[0m",
	SnakeSpeedMS:        1000,
	FoodSymbol:          "\033[42;30m \033[0m",
	Online:              false,
	Server:              false,
	LocalIP:             "127.0.0.1",
	LocalPort:           9001,
	DialIP:              "127.0.0.1",
	DialPort:            9002,
}

func init() {
	flag.BoolVar(&(options.Online), "online", false, "play on network")
	flag.BoolVar(&(options.Server), "svr", false, "run as server")
	flag.StringVar(&(options.DialIP), "li", "127.0.0.1", "dial ip")
	flag.IntVar(&(options.LocalPort), "lp", 9001, "local port")
	flag.StringVar(&(options.LocalIP), "di", "127.0.0.1", "local ip")
	flag.IntVar(&(options.DialPort), "dp", 9002, "dial port")
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
