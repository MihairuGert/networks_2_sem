package main

import "snake-game/internal/application"

func main() {
	var game application.Game
	game.Init()
	err := game.Start()
	if err != nil {
		panic(err)
	}
}
