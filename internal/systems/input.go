package systems

import (
	"fmt"
	"math"

	"github.com/gfeyer/ecs_sim/internal/components"
	"github.com/gfeyer/ecs_sim/internal/constants"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// UpdateInput handles mouse clicks to set a destination for the player.
func UpdateInput(ecs *ecs.ECS) {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}

	player, ok := components.PlayerControlled.First(ecs.World)
	if !ok {
		return // No player to command.
	}

	settings, ok := components.Settings.First(ecs.World)
	if !ok {
		return
	}
	settingsData := components.Settings.Get(settings)

	camera, ok := components.Camera.First(ecs.World)
	if !ok {
		return
	}
	cameraData := components.Camera.Get(camera)

	// Convert screen coordinates to grid coordinates.
	screenX, screenY := ebiten.CursorPosition()
	// Convert screen coordinates to world coordinates using the camera.
	worldX, worldY := cameraData.ScreenToWorld(float64(screenX), float64(screenY))

	// Adjust for the map's centered origin.
	mapScreenX := worldX - float64(settingsData.ScreenWidth/2)
	mapScreenY := worldY

	tileWidth := float64(constants.TileWidth)
	tileHeight := float64(constants.TileHeight)

	// Reverse the isometric projection.
	gridX := int(math.Floor((mapScreenX/(tileWidth/2))+(mapScreenY/(tileHeight/2))) / 2)
	gridY := int(math.Floor((mapScreenY/(tileHeight/2))-(mapScreenX/(tileWidth/2))) / 2)

	fmt.Printf("Input: Clicked grid coordinates (%d, %d)\n", gridX, gridY)

	// Add or update the destination component on the player entity.
	destinationData := &components.DestinationData{X: gridX, Y: gridY}

	if player.HasComponent(components.Destination) {
		// Component exists, so update it.
		donburi.SetValue(player, components.Destination, *destinationData)
	} else {
		// Component does not exist, so create it.
		donburi.Add(player, components.Destination, destinationData)
	}
}
