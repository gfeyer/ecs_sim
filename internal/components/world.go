package components

import "github.com/yohamta/donburi"

// WorldData holds the overall map structure and acts as a singleton resource.
type WorldData struct {
	Width     int
	Height    int
	// A 2D grid for quick lookup of tile entities by their coordinate.
	EntityMap [][][]donburi.Entity
}

var World = donburi.NewComponentType[WorldData]()
