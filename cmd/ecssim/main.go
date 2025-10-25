package main

import (
	"log"
	_ "net/http/pprof" // Import for side-effect: registers pprof handlers

	"github.com/gfeyer/ecs_sim/internal/constants"
	"github.com/gfeyer/ecs_sim/internal/game"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	g := game.NewGame(constants.ScreenWidth, constants.ScreenHeight)
	// The screen size is now managed by the Layout function, which reads from the ECS.
	// We can set a default here, but the Layout function is the source of truth.
	ebiten.SetWindowSize(constants.ScreenWidth, constants.ScreenHeight)
	ebiten.SetWindowTitle("ECS Isometric Tile System")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
