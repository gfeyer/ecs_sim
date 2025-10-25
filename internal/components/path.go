package components

import (
	"image"

	"github.com/yohamta/donburi"
)

// Path is a component that stores a sequence of points for an entity to follow.
type PathData struct {
	Points []image.Point
}

var Path = donburi.NewComponentType[PathData]()
