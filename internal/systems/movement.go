package systems

import (
	"time"

	"github.com/gfeyer/ecs_sim/internal/components"
	"github.com/gfeyer/ecs_sim/internal/constants"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/donburi/filter"
)

var movementQuery = donburi.NewQuery(
	filter.Contains(components.Position, components.Path, components.Movement),
)

// UpdateMovement moves entities along their calculated paths with smooth animation.
func UpdateMovement(ecs *ecs.ECS) {
	movementQuery.Each(ecs.World, func(entry *donburi.Entry) {
		movement := components.Movement.Get(entry)
		if movement.Paused {
			return // Skip updates if paused.
		}
		pos := components.Position.Get(entry)

		// If the entity is not currently moving, check if it should start a new movement.
		if !movement.IsMoving {
			path := components.Path.Get(entry)
			if len(path.Points) == 0 {
				entry.RemoveComponent(components.Path) // No more path, remove component.
				return
			}

			// Start a new movement segment.
			movement.IsMoving = true
			movement.TargetX = path.Points[0].X
			movement.TargetY = path.Points[0].Y
			movement.LastMove = time.Now()
			path.Points = path.Points[1:] // Consume the next point in the path.

			// Update sprite based on direction.
			if renderable := components.Renderable.Get(entry); renderable != nil {
				if movement.TargetX < pos.GridX {
					renderable.CurrentPart = constants.AssetRobotLeft
				} else if movement.TargetX > pos.GridX {
					renderable.CurrentPart = constants.AssetRobotRight
				} else if movement.TargetY < pos.GridY {
					renderable.CurrentPart = constants.AssetRobotUp
				} else if movement.TargetY > pos.GridY {
					renderable.CurrentPart = constants.AssetRobotDown
				}
			}

			// fmt.Printf("Movement: Starting new segment. From (%d, %d) to (%d, %d)\n", pos.GridX, pos.GridY, movement.TargetX, movement.TargetY)
			return // Return immediately to start animation on the next frame.
		}

		// If the entity is moving, update its animation progress.
		if movement.IsMoving {
			progress := time.Since(movement.LastMove).Seconds() / movement.MoveSpeed.Seconds()

			if progress >= 1.0 {
				// Movement segment is complete.
				pos.GridX = movement.TargetX
				pos.GridY = movement.TargetY
				pos.OffsetX = 0
				pos.OffsetY = 0
				movement.IsMoving = false
				// fmt.Printf("Movement: Segment complete. Arrived at (%d, %d)\n", pos.GridX, pos.GridY)
			} else {
				// Continue animating towards the target.
				dx := movement.TargetX - pos.GridX
				dy := movement.TargetY - pos.GridY
				totalOffsetX := float64(dx-dy) * constants.TileWidth / 2  // Half tile width
				totalOffsetY := float64(dx+dy) * constants.TileHeight / 2 // Half tile height

				pos.OffsetX = totalOffsetX * progress
				pos.OffsetY = totalOffsetY * progress
				// fmt.Printf("Movement: Progress: %.2f, OffsetX: %.2f, OffsetY: %.2f\n", progress, pos.OffsetX, pos.OffsetY)
			}
		}
	})
}
