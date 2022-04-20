package main

import (
	"fmt"
	"gosnake"
	"os"
)

var options = gosnake.GameOptions{
	GroundWith:    50,
	GroundHeight:  30,
	GroundSymbol:  " ",
	BordersWidth:  50,
	BordersHeight: 30,
	BordersSymbol: "\033[47;30m \033[0m",
	SnakeInitPosX: 24,
	SnakeInitPosY: 14,
	SnakeInitDir:  gosnake.DirRight,
	SnakeSymbol:   "\033[47;30m \033[0m",
	SnakeSpeedMS:  300,
	FoodSymbol:    "\033[47;30m \033[0m",
}

func main() {
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
