package systems

import (
	"fmt"
	"image/color"
	"time"

	"reflect"

	"github.com/gfeyer/ecs_sim/internal/components"
	"github.com/gfeyer/ecs_sim/internal/constants"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/donburi/filter"
)

var (
	getDebugCompQuery = donburi.NewQuery(filter.Contains(components.Debug))
	settingsQuery     = donburi.NewQuery(filter.Contains(components.Settings))
)

var (
	renderableQuery = donburi.NewQuery(filter.Contains(components.Position, components.Renderable))
)

// DrawDebugText renders debug text like FPS to the screen.
func DrawDebugText(ecs *ecs.ECS, screen *ebiten.Image) {
	entry, ok := getDebugCompQuery.First(ecs.World)
	if !ok {
		return
	}
	debug := components.Debug.Get(entry)

	// Only update the debug message every 250ms.
	if time.Since(debug.LastUpdated) > 250*time.Millisecond {
		debug.CachedMsg = fmt.Sprintf("FPS: %.2f", ebiten.ActualFPS())
		debug.LastUpdated = time.Now()
	}

	// Draw the cached message.
	ebitenutil.DebugPrintAt(screen, debug.CachedMsg, 10, screen.Bounds().Dy()-20)

	// Draw game instructions at bottom
	ebitenutil.DebugPrintAt(screen, "Use Arrows or Middle Mouse Button to pan around the map\nUse Mouse Wheel to zoom in or out\nRobot simulation:\n\tWatch the robots collect resources\n\tand return them to base", 10, 10)
}

// DrawDebug renders performance metrics and tile coordinates to the screen.
func DrawDebug(ecs *ecs.ECS, screen *ebiten.Image) {

	settingsEntry, ok := settingsQuery.First(ecs.World)
	if !ok {
		return
	}
	settings := components.Settings.Get(settingsEntry)

	// Draw tile coordinates
	renderableQuery.Each(ecs.World, func(entry *donburi.Entry) {
		position := components.Position.Get(entry)

		// Calculate isometric screen position from grid coordinates.
		isoX := (position.GridX - position.GridY) * (constants.TileWidth / 2)
		isoY := (position.GridX + position.GridY) * (constants.TileHeight / 2)

		// Adjust for tile anchor and center the map on the screen.
		finalX := float64(isoX-(constants.TileWidth/2)+settings.ScreenWidth/2) + position.OffsetX
		finalY := float64(isoY) + position.OffsetY

		// Draw debug diamond
		topX, topY := finalX+constants.TileWidth/2, finalY
		rightX, rightY := finalX+constants.TileWidth, finalY+constants.TileHeight/2
		bottomX, bottomY := finalX+constants.TileWidth/2, finalY+constants.TileHeight
		leftX, leftY := finalX, finalY+constants.TileHeight/2

		// Draw the diamond outline
		lineColor := color.RGBA{R: 255, G: 0, B: 0, A: 255} // Red
		ebitenutil.DrawLine(screen, topX, topY, rightX, rightY, lineColor)
		ebitenutil.DrawLine(screen, rightX, rightY, bottomX, bottomY, lineColor)
		ebitenutil.DrawLine(screen, bottomX, bottomY, leftX, leftY, lineColor)
		ebitenutil.DrawLine(screen, leftX, leftY, topX, topY, lineColor)

		// Draw coordinates text
		textX := finalX + constants.TileWidth/2
		textY := finalY + constants.TileHeight/2
		coords := fmt.Sprintf("(%d,%d)", position.GridX, position.GridY)
		ebitenutil.DebugPrintAt(screen, coords, int(textX), int(textY))
	})
}

// PrintDebugGrid prints a text representation of the world grid to the console.
func PrintDebugGrid(ecs *ecs.ECS) {
	world, ok := components.World.First(ecs.World)
	if !ok {
		fmt.Println("Debug Grid: World component not found.")
		return
	}
	worldData := components.World.Get(world)

	fmt.Println("--- Initial Grid State ---")
	for y := 0; y < worldData.Height; y++ {
		line := ""
		for x := 0; x < worldData.Width; x++ {
			entities := worldData.EntityMap[y][x]
			if len(entities) > 0 {
				var tileFound bool
				for _, entity := range entities {
					entry := ecs.World.Entry(entity)
					if entry.HasComponent(components.Tile) {
						tileData := components.Tile.Get(entry)
						line += fmt.Sprintf("[%3d]", tileData.ID)
						tileFound = true
						break
					}
				}
				if !tileFound {
					line += "[ P ]" // Assume it's the player if not a tile
				}
			} else {
				line += "[ . ]"
			}
		}
		fmt.Println(line)
	}
	fmt.Println("------------------------")
}

// PrintDebugPropertyGrid prints a text representation of a boolean property for each tile.
func PrintDebugPropertyGrid(ecs *ecs.ECS, propertyName string) {
	world, ok := components.World.First(ecs.World)
	if !ok {
		fmt.Println("Debug Grid: World component not found.")
		return
	}
	worldData := components.World.Get(world)

	fmt.Printf("--- Grid Property State: %s ---\n", propertyName)
	for y := 0; y < worldData.Height; y++ {
		line := ""
		for x := 0; x < worldData.Width; x++ {
			entities := worldData.EntityMap[y][x]
			if len(entities) > 0 {
				var tileFound bool
				for _, entity := range entities {
					entry := ecs.World.Entry(entity)
					if entry.HasComponent(components.Tile) {
						tileData := components.Tile.Get(entry)
						// Use reflection to get the property value by name.
						r := reflect.ValueOf(tileData)
						f := reflect.Indirect(r).FieldByName(propertyName)
						if f.IsValid() && f.Kind() == reflect.Bool {
							if f.Bool() {
								line += "[X]"
							} else {
								line += "[ ]"
							}
						} else {
							line += "[?]"
						}
						tileFound = true
						break
					}
				}
				if !tileFound {
					line += "[P]"
				}
			} else {
				line += "[.]"
			}
		}
		fmt.Println(line)
	}
	fmt.Println("---------------------------------")
}
