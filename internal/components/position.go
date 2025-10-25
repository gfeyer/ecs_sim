package components

import (
	"github.com/yohamta/donburi"
)

// Position holds the grid and visual offset data for any entity.
type PositionData struct {
	// The "Source of Truth" for game logic.
	GridX int
	GridY int

	// The "Visual Offset" for smooth animation, in pixels.
	// When an entity is perfectly centered on its tile, these are 0.
	OffsetX float64
	OffsetY float64
}

var Position = donburi.NewComponentType[PositionData]()
