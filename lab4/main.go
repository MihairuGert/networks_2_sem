package main

import "snake-game/internal/application"

func main() {
	var game application.Game
	err := game.Init()
	if err != nil {
		panic(err)
	}
	err = game.Start()
	if err != nil {
		panic(err)
	}
}
