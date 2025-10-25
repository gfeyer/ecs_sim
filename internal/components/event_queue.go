package components

import (
	"github.com/gfeyer/ecs_sim/internal/constants"
	"github.com/yohamta/donburi"
)

type SpawnRequestData struct {
	ItemType constants.ItemType
	GridX    int
	GridY    int
}

// EventQueue is used to queue events for processing.
type EventQueueData struct {
	SpawnRequests []SpawnRequestData
}

var EventQueue = donburi.NewComponentType[EventQueueData]()
