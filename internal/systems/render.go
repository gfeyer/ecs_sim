package systems

import (
	"image/color"
	"slices"

	"github.com/gfeyer/ecs_sim/internal/components"
	"github.com/gfeyer/ecs_sim/internal/constants"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/donburi/filter"
)

var renderablesQuery = donburi.NewQuery(filter.Contains(components.Position, components.Renderable))

// DrawRenderables queries for all drawable entities, sorts them correctly, and renders them.
func DrawRenderables(ecs *ecs.ECS, screen *ebiten.Image) {
	// screen.Fill(color.Black)
	screen.Fill(color.RGBA{R: 200, G: 160, B: 96, A: 255})

	settingsEntry, ok := components.Settings.First(ecs.World)
	if !ok {
		return
	}
	settings := components.Settings.Get(settingsEntry)

	cameraEntry, ok := components.Camera.First(ecs.World)
	if !ok {
		return
	}
	camera := components.Camera.Get(cameraEntry)

	type RenderableEntry struct {
		Position   *components.PositionData
		Renderable *components.RenderableData
	}

	entries := []RenderableEntry{}
	renderablesQuery.Each(ecs.World, func(entry *donburi.Entry) {
		entries = append(entries, RenderableEntry{
			Position:   components.Position.Get(entry),
			Renderable: components.Renderable.Get(entry),
		})
	})

	// Sort entries for correct isometric depth.
	// 1. By DrawOrder (backgrounds first).
	// 2. Then by GridY (top-to-bottom).
	slices.SortFunc(entries, func(a, b RenderableEntry) int {
		if a.Renderable.DrawOrder != b.Renderable.DrawOrder {
			return a.Renderable.DrawOrder - b.Renderable.DrawOrder
		}
		return a.Position.GridY - b.Position.GridY
	})

	for _, entry := range entries {
		// Calculate base isometric screen position from grid coordinates.
		isoX := (entry.Position.GridX - entry.Position.GridY) * (constants.TileWidth / 2)
		isoY := (entry.Position.GridX + entry.Position.GridY) * (constants.TileHeight / 2)

		op := &ebiten.DrawImageOptions{}

		part, ok := entry.Renderable.Parts[entry.Renderable.CurrentPart]
		if !ok {
			continue // Skip rendering if the current part is not found
		}

		// 1. Anchor offset: Adjust from top-left of sprite to top corner of diamond.
		// 2. Global offset: Center the map on the screen.
		// 3. Visual offset: Apply smooth movement animation offset.
		rect := part.SourceRect
		tileOffsetX := 0
		tileOffsetY := 0
		if rect.Dx() == 64 && rect.Dy() == 64 {
			tileOffsetX = constants.Tile64OffsetX
			tileOffsetY = constants.Tile64OffsetY
		} else {
			tileOffsetX = constants.Tile128OffsetX
			tileOffsetY = constants.Tile128OffsetY
		}

		finalX := float64(isoX-(constants.TileWidth/2)+settings.ScreenWidth/2) + entry.Position.OffsetX + float64(tileOffsetX)
		finalY := float64(isoY) + entry.Position.OffsetY + float64(tileOffsetY)

		op.GeoM.Translate(finalX, finalY)

		// Apply the camera's view matrix to the final drawing options.
		op.GeoM.Concat(camera.View)

		screen.DrawImage(part.Image.SubImage(part.SourceRect).(*ebiten.Image), op)
	}
}
