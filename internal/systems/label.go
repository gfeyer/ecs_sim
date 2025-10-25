package systems

import (
	"fmt"
	"image/color"

	"github.com/gfeyer/ecs_sim/internal/components"
	"github.com/gfeyer/ecs_sim/internal/constants"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/donburi/filter"
	"golang.org/x/image/font/basicfont"
)

var labelQuery = donburi.NewQuery(filter.Contains(components.Position, components.Label))

func UpdateLabels(ecs *ecs.ECS) {
	labelQuery.Each(ecs.World, func(entry *donburi.Entry) {
		label := components.Label.Get(entry)

		if entry.HasComponent(components.Mine) {
			// update mineData label
			mineData := components.Mine.Get(entry)
			label.Text = fmt.Sprintf("%.2f", mineData.Resources)
		} else if entry.HasComponent(components.Unit) {
			// update robot label
			unitData := components.Unit.Get(entry)
			label.Text = fmt.Sprintf("%s\nOre: %.2f", unitData.Name, unitData.CurrentLoad)
		} else if entry.HasComponent(components.Base) {
			// update base label
			baseData := components.Base.Get(entry)
			label.Text = fmt.Sprintf("Base Ore: %.2f", baseData.Resources)
		}
	})
}

// DrawLabels renders the text from Label components on the screen.
func DrawLabels(ecs *ecs.ECS, screen *ebiten.Image) {
	cameraEntry, ok := components.Camera.First(ecs.World)
	if !ok {
		return
	}
	camera := components.Camera.Get(cameraEntry)

	settingsEntry, ok := components.Settings.First(ecs.World)
	if !ok {
		return
	}
	settings := components.Settings.Get(settingsEntry)

	labelQuery.Each(ecs.World, func(entry *donburi.Entry) {
		pos := components.Position.Get(entry)
		label := components.Label.Get(entry)

		// Calculate the isometric position of the entity's origin.
		isoX := (pos.GridX - pos.GridY) * (constants.TileWidth / 2)
		isoY := (pos.GridX + pos.GridY) * (constants.TileHeight / 2)

		// Calculate the final world coordinates for the label, including all offsets.
		worldX := float64(isoX-(constants.TileWidth/2)+settings.ScreenWidth/2) + pos.OffsetX + float64(label.OffsetX)
		worldY := float64(isoY) + pos.OffsetY + float64(label.OffsetY)

		// Apply the camera's transformation to the world coordinates to get the final screen coordinates.
		screenX, screenY := camera.View.Apply(worldX, worldY)

		// Draw the text at the transformed screen coordinates.
		text.Draw(screen, label.Text, basicfont.Face7x13, int(screenX), int(screenY), color.White)
	})
}
