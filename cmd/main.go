package main

import (
	"flag"
	"fmt"
	"gosnake"
	"os"
)

var options gosnake.GameOptions

func init() {
	flag.IntVar(&(options.GroundWith), "gw", 30, "ground width")
	flag.IntVar(&(options.GroundHeight), "gh", 30, "ground height")
	flag.StringVar(&(options.GroundSymbol), "gs", "  ", "ground symbol")
	flag.IntVar(&(options.BorderWidth), "bw", 30, "border width")
	flag.IntVar(&(options.BorderHeight), "bh", 30, "border height")
	flag.StringVar(&(options.BorderSymbol), "bs", "\033[47;37m  \033[0m", "border symbol")
	flag.IntVar(&(options.Snake1InitPosX), "s1x", 14, "snake1 init postion x")
	flag.IntVar(&(options.Snake1InitPosY), "s1y", 7, "snake1 init position y")
	flag.IntVar((*int)(&(options.Snake1InitDir)), "s1d", int(gosnake.DirLeft), "snake1 init move direction")
	flag.StringVar(&(options.Snake1Symbol), "s1s", "\033[44;37m  \033[0m", "snake1 symbol")
	flag.IntVar(&(options.Snake2InitPosX), "s2x", 7, "snake2 init postion x")
	flag.IntVar(&(options.Snake2InitPosY), "s2y", 14, "snake2 init position y")
	flag.IntVar((*int)(&(options.Snake2InitDir)), "s2d", int(gosnake.DirRight), "snake2 init move direction")
	flag.StringVar(&(options.Snake2Symbol), "s2s", "\033[41;37m  \033[0m", "snake2 symbol")
	flag.IntVar(&(options.SnakeSpeedMS), "sp", 300, "snake speed")
	flag.StringVar(&(options.FoodSymbol), "fs", "\033[41;37m  \033[0m", "food symbol")
	flag.BoolVar(&(options.Online), "ol", false, "play online")
	flag.BoolVar(&(options.Server), "sv", false, "play as a server")
	flag.StringVar(&(options.LocalAddr), "la", "0.0.0.0:10001", "local addr")
	flag.StringVar(&(options.RemoteAddr), "ra", "0.0.0.0:10002", "local ip")
	flag.IntVar(&(options.FPS), "fs", 30, "render fps")
}

func main() {
	flag.Parse()
	game := gosnake.NewGame(&options)
	if err := game.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}
