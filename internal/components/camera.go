package components

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

// CameraData represents the camera's state.
type CameraData struct {
	X, Y  float64
	Speed float64

	Zoom           float64
	ZoomSpeed      float64
	IsPanning      bool
	PanStartX, PanStartY int
	View           ebiten.GeoM
}

var Camera = donburi.NewComponentType[CameraData]()

// ScreenToWorld converts screen coordinates to world coordinates.
func (c *CameraData) ScreenToWorld(x, y float64) (float64, float64) {
	inverseView := c.View
	if inverseView.IsInvertible() {
		inverseView.Invert()
		return inverseView.Apply(x, y)
	}
	// Fallback if the matrix is not invertible, though this shouldn't happen with typical camera transformations.
	return x, y
}
