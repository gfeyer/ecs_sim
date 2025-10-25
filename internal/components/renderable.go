package components

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

// RenderablePart holds an image and its source rectangle, representing one visual state of an entity.
type RenderablePart struct {
	Image      *ebiten.Image
	SourceRect image.Rectangle
}

// Renderable holds the visual information for any entity.
type RenderableData struct {
	// Parts holds the different visual representations of the entity, keyed by a name (e.g., "default", "left", "right").
	Parts map[string]RenderablePart
	// CurrentPart is the key for the currently active part in the Parts map.
	CurrentPart string
	// DrawOrder determines rendering layer. Lower numbers are drawn first (further back).
	DrawOrder int
}

var Renderable = donburi.NewComponentType[RenderableData]()
