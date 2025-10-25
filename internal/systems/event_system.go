package systems

import (
	"fmt"

	"github.com/gfeyer/ecs_sim/internal/components"
	"github.com/gfeyer/ecs_sim/internal/constants"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/donburi/filter"
)

var eventQueueQuery = donburi.NewQuery(
	filter.Contains(components.EventQueue),
)

type EventSystem struct {
	factory *Factory
}

func NewEventSystem(factory *Factory) *EventSystem {
	return &EventSystem{
		factory: factory,
	}
}

func (s *EventSystem) Update(ecs *ecs.ECS) {
	// Find the singleton event queue entity.
	eventQueueEntry, ok := eventQueueQuery.First(ecs.World)
	if !ok {
		fmt.Println("No event queue found")
		return
	}

	eventQueue := components.EventQueue.Get(eventQueueEntry)
	for _, spawnRequest := range eventQueue.SpawnRequests {
		switch spawnRequest.ItemType {
		case constants.TypeRobot:
			s.factory.CreateRobot(spawnRequest.GridX, spawnRequest.GridY)
		case constants.TypeMine:
			s.factory.CreateMine(spawnRequest.GridX, spawnRequest.GridY)
		case constants.TypeBase:
			s.factory.CreateBase(spawnRequest.GridX, spawnRequest.GridY)
		default:
			fmt.Printf("Unknown spawn request: %d\n", spawnRequest.ItemType)
		}
	}

	// Clear the queue after processing.
	eventQueue.SpawnRequests = eventQueue.SpawnRequests[:0]
}
