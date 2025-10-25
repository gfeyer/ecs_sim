package components

import "github.com/yohamta/donburi"

// TileData holds information about a specific tile.
type TileData struct {
	ID         uint32
	IsWalkable bool
}

var Tile = donburi.NewComponentType[TileData]()
