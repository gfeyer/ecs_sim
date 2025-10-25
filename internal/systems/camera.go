package systems

import (
	"github.com/gfeyer/ecs_sim/internal/components"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi/ecs"
)

func UpdateCamera(ecs *ecs.ECS) {
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

	// Keyboard panning
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		camera.X -= camera.Speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		camera.X += camera.Speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		camera.Y -= camera.Speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		camera.Y += camera.Speed
	}

	// Mouse Panning
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle) {
		camera.IsPanning = true
		camera.PanStartX, camera.PanStartY = ebiten.CursorPosition()
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonMiddle) {
		camera.IsPanning = false
	}
	if camera.IsPanning {
		cx, cy := ebiten.CursorPosition()
		dx := float64(cx - camera.PanStartX)
		dy := float64(cy - camera.PanStartY)
		camera.X -= dx / camera.Zoom
		camera.Y -= dy / camera.Zoom
		camera.PanStartX, camera.PanStartY = cx, cy
	}

	// Mouse Zooming
	_, wheelY := ebiten.Wheel()
	if wheelY != 0 {
		// Get the cursor position on screen.
		mx, my := ebiten.CursorPosition()

		// Find the world coordinate under the cursor before the zoom.
		wx, wy := camera.ScreenToWorld(float64(mx), float64(my))

		// Apply zoom.
		if wheelY > 0 {
			camera.Zoom += camera.ZoomSpeed
		} else if wheelY < 0 {
			camera.Zoom -= camera.ZoomSpeed
		}

		// Clamp zoom level.
		if camera.Zoom < 0.5 {
			camera.Zoom = 0.5
		} else if camera.Zoom > 2.0 {
			camera.Zoom = 2.0
		}

		// Temporarily update the view matrix to get the new world coordinates.
		// We need to do this to correctly calculate the post-zoom world position.
		{
			tempView := ebiten.GeoM{}
			tempView.Translate(float64(settings.ScreenWidth)/2, float64(settings.ScreenHeight)/2)
			tempView.Scale(camera.Zoom, camera.Zoom)
			tempView.Translate(-float64(settings.ScreenWidth)/2, -float64(settings.ScreenHeight)/2)
			tempView.Translate(-camera.X, -camera.Y)

			inverseView := tempView
			inverseView.Invert()
			nwx, nwy := inverseView.Apply(float64(mx), float64(my))

			// Adjust camera position to counteract the drift.
			camera.X += wx - nwx
			camera.Y += wy - nwy
		}
	}

	// Update the camera's view matrix with the final, correct values for this frame.
	camera.View.Reset()
	camera.View.Translate(float64(settings.ScreenWidth)/2, float64(settings.ScreenHeight)/2)
	camera.View.Scale(camera.Zoom, camera.Zoom)
	camera.View.Translate(-float64(settings.ScreenWidth)/2, -float64(settings.ScreenHeight)/2)
	camera.View.Translate(-camera.X, -camera.Y)
}
