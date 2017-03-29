package main

import (
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	scene := new(Scene)
	sim := new(Sim)

	scene.RunGame(sim)
}
