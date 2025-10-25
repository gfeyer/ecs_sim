package systems

import (
	"fmt"

	"time"

	"github.com/gfeyer/ecs_sim/internal/components"
	"github.com/gfeyer/ecs_sim/internal/constants"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// Spawn requests that a new entity be created with the given name and position.
// This is a helper function that provides a clean API for other systems to request spawning.
func Spawn(ecs *ecs.ECS, itemType constants.ItemType, gridX, gridY int) {
	// Find the singleton event queue entity.
	entry, ok := eventQueueQuery.First(ecs.World)
	if !ok {
		// This should not happen if the queue is created at startup.
		// For a real game, you might want to log a more serious error here.
		fmt.Println("event queue not found")
		return
	}

	eventQueue := components.EventQueue.Get(entry)
	eventQueue.SpawnRequests = append(eventQueue.SpawnRequests, components.SpawnRequestData{
		ItemType: itemType,
		GridX:    gridX,
		GridY:    gridY,
	})
}

// PauseMovement pauses or unpauses an entity's movement.
func PauseMovement(entry *donburi.Entry, paused bool) {
	if !entry.HasComponent(components.Movement) {
		return
	}
	movement := components.Movement.Get(entry)

	if paused && !movement.Paused {
		// Pause the movement
		movement.Paused = true
		movement.PausedTime = time.Now()
	} else if !paused && movement.Paused {
		// Unpause the movement
		movement.Paused = false
		pauseDuration := time.Since(movement.PausedTime)
		movement.LastMove = movement.LastMove.Add(pauseDuration)
	}
}

func setNewDestination(entry *donburi.Entry, gridX, gridY int) {
	destinationData := &components.DestinationData{X: gridX, Y: gridY}

	if entry.HasComponent(components.Destination) {
		// Component exists, so update it.
		donburi.SetValue(entry, components.Destination, *destinationData)
	} else {
		// Component does not exist, so create it.
		donburi.Add(entry, components.Destination, destinationData)
	}
}
