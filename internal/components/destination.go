package components

import "github.com/yohamta/donburi"

// Destination is a component that marks an entity with a target grid location.
type DestinationData struct {
	X, Y int
}

var Destination = donburi.NewComponentType[DestinationData]()
