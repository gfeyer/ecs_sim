package systems

import (
	"fmt"
	"math"

	"github.com/gfeyer/ecs_sim/internal/components"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/donburi/filter"
)

var robotAIQuery = donburi.NewQuery(
	filter.Contains(components.Unit),
)

var mineQuery = donburi.NewQuery(
	filter.Contains(components.Mine, components.Position),
)

var buildingQuery = donburi.NewQuery(
	filter.Contains(components.Base, components.Position),
)

// UpdateRobotAI updates the State for gathering robots
func UpdateRobotAI(ecs *ecs.ECS) {
	robotAIQuery.Each(ecs.World, func(unitEntry *donburi.Entry) {
		unitData := components.Unit.Get(unitEntry)

		switch unitData.State {
		case components.Idle:
			handleIdle(ecs, unitEntry)
		case components.MovingToMine:
			handleMovingToMine(ecs, unitEntry)
		case components.Mining:
			handleMining(ecs, unitEntry)
		case components.ReturningToBase:
			handleReturningToBase(ecs, unitEntry)
		case components.Unloading:
			handleUnloading(unitEntry)
		case components.WaitingForBase:
			handleWaitingForBase(ecs, unitEntry)
		case components.MovingToWaitingSpot:
			handleMovingToWaitingSpot(ecs, unitEntry)
		case components.MovingToMineWaitingSpot:
			handleMovingToMineWaitingSpot(ecs, unitEntry)
		case components.WaitingForMine:
			handleWaitingForMine(ecs, unitEntry)
		}
	})
}

func handleIdle(ecs *ecs.ECS, unitEntry *donburi.Entry) {
	mineEntry := findClosestMine(ecs, unitEntry)

	if mineEntry == nil {
		// fmt.Printf("No mine found for unit: %v\n", unitEntry)
		return
	}

	startMovingToMine(ecs, unitEntry, mineEntry)
}

func handleMovingToMine(ecs *ecs.ECS, unitEntry *donburi.Entry) {
	unitData := components.Unit.Get(unitEntry)
	positionData := components.Position.Get(unitEntry)

	if !ecs.World.Valid(unitData.TargetEntry.Entity()) {
		fmt.Println("Target mine no longer exists, going idle")
		unitData.State = components.Idle
		return
	}

	mineData := components.Mine.Get(unitData.TargetEntry)
	if mineData.Busy && mineData.ClaimedBy != unitEntry {
		fmt.Println("Mine was claimed by another unit, going idle")
		unitData.State = components.Idle
		return
	}

	if positionData.GridX == unitData.TargetPos.X && positionData.GridY == unitData.TargetPos.Y {
		unitData.State = components.Mining
		fmt.Println("Arrived at mine")
	}
}

func handleMining(ecs *ecs.ECS, unitEntry *donburi.Entry) {
	unitData := components.Unit.Get(unitEntry)
	mineEntry := unitData.TargetEntry

	// Safety check: ensure the mine entity still exists.
	if !ecs.World.Valid(mineEntry.Entity()) {
		fmt.Println("Mine disappeared while mining, going idle.")
		unitData.State = components.Idle
		return
	}
	mineData := components.Mine.Get(mineEntry)

	unitData.CurrentLoad += unitData.LoadSpeed
	mineData.Resources -= unitData.LoadSpeed

	if mineData.Resources <= 0 {
		fmt.Println("Mine depleted, removing it.")
		ecs.World.Remove(mineEntry.Entity())
		// Robot is now full, so it should go to base.
		baseEntry := findClosestBase(ecs, unitEntry)
		if baseEntry == nil {
			fmt.Printf("No base found for unit: %v\n", unitEntry)
			unitData.State = components.Idle // Or some other state
			return
		}
		startReturningToBase(ecs, unitEntry, baseEntry)
		return
	}

	if unitData.CurrentLoad >= unitData.MaxLoadCapacity {
		fmt.Println("Capacity reached, leaving mine.")
		mineData.Busy = false // Release the mine
		mineData.ClaimedBy = nil
		baseEntry := findClosestBase(ecs, unitEntry)
		if baseEntry == nil {
			fmt.Printf("No base found for unit: %v\n", unitEntry)
			return
		}
		startReturningToBase(ecs, unitEntry, baseEntry)
	}
}

func handleReturningToBase(ecs *ecs.ECS, unitEntry *donburi.Entry) {
	unitData := components.Unit.Get(unitEntry)
	pos := components.Position.Get(unitEntry)
	baseEntry := unitData.TargetEntry
	base := components.Base.Get(baseEntry)
	basePos := components.Position.Get(baseEntry)

	// Check if we have arrived at the base
	if pos.GridX == basePos.GridX && pos.GridY == basePos.GridY {
		// Ensure we have the claim
		if base.ClaimedBy == unitEntry {
			unitData.State = components.Unloading
			fmt.Printf("%s arrived at base for unloading.\n", unitData.Name)
		} else {
			// This should not happen if logic is correct, but as a safeguard
			fmt.Printf("%s arrived at a busy base it didn't claim. Returning to idle.\n", unitData.Name)
			unitData.State = components.Idle
		}
		return
	}

	// Check if we are adjacent to the base
	dx := math.Abs(float64(pos.GridX - basePos.GridX))
	dy := math.Abs(float64(pos.GridY - basePos.GridY))
	isAdjacent := (dx == 1 && dy == 0) || (dx == 0 && dy == 1)

	if isAdjacent {
		if !base.Busy {
			fmt.Printf("%s is adjacent and claims the base.\n", unitData.Name)
			base.Busy = true
			base.ClaimedBy = unitEntry
			// Allow movement to continue to the base itself
			movement := components.Movement.Get(unitEntry)
			movement.Paused = false
		} else if base.ClaimedBy != unitEntry {
			fmt.Printf("%s is adjacent to a busy base. Waiting...\n", unitData.Name)
			unitData.State = components.WaitingForBase
			// Pause movement
			movement := components.Movement.Get(unitEntry)
			movement.Paused = true
		}
	}
}

func handleMovingToWaitingSpot(ecs *ecs.ECS, unitEntry *donburi.Entry) {
	unitData := components.Unit.Get(unitEntry)
	positionData := components.Position.Get(unitEntry)

	if positionData.GridX == unitData.TargetPos.X && positionData.GridY == unitData.TargetPos.Y {
		unitData.State = components.WaitingForBase
		fmt.Println("Arrived at base waiting spot.")
	}
}

func handleMovingToMineWaitingSpot(ecs *ecs.ECS, unitEntry *donburi.Entry) {
	unitData := components.Unit.Get(unitEntry)
	positionData := components.Position.Get(unitEntry)

	if !ecs.World.Valid(unitData.TargetEntry.Entity()) {
		fmt.Println("Target mine no longer exists, going idle")
		unitData.State = components.Idle
		return
	}

	if positionData.GridX == unitData.TargetPos.X && positionData.GridY == unitData.TargetPos.Y {
		unitData.State = components.WaitingForMine
		fmt.Println("Arrived at mine waiting spot.")
	}
}

func handleWaitingForMine(ecs *ecs.ECS, unitEntry *donburi.Entry) {
	unitData := components.Unit.Get(unitEntry)

	if !ecs.World.Valid(unitData.TargetEntry.Entity()) {
		fmt.Println("Target mine no longer exists, going idle")
		unitData.State = components.Idle
		return
	}

	mineData := components.Mine.Get(unitData.TargetEntry)

	if !mineData.Busy {
		fmt.Println("Mine is now free. Claiming it and moving to mine.")
		mineData.Busy = true // Claim the mine
		mineData.ClaimedBy = unitEntry
		minePosition := components.Position.Get(unitData.TargetEntry)
		setNewDestination(unitEntry, minePosition.GridX, minePosition.GridY)
		unitData.State = components.MovingToMine
		unitData.TargetPos = components.Point{X: minePosition.GridX, Y: minePosition.GridY}
	}
}

func handleUnloading(unitEntry *donburi.Entry) {
	unitData := components.Unit.Get(unitEntry)
	baseData := components.Base.Get(unitData.TargetEntry)

	unloadSpeed := unitData.LoadSpeed * 10
	unitData.CurrentLoad = unitData.CurrentLoad - unloadSpeed
	baseData.Resources += unloadSpeed
	if unitData.CurrentLoad <= 0 {
		baseData.Busy = false
		baseData.ClaimedBy = nil
		unitData.State = components.Idle
		unitData.CurrentLoad = 0
	}
}

func handleWaitingForBase(ecs *ecs.ECS, unitEntry *donburi.Entry) {
	unitData := components.Unit.Get(unitEntry)
	baseEntry := unitData.TargetEntry
	base := components.Base.Get(baseEntry)

	if !base.Busy {
		fmt.Printf("%s sees base is free, claiming and resuming path.\n", unitData.Name)
		base.Busy = true
		base.ClaimedBy = unitEntry
		unitData.State = components.ReturningToBase

		// Resume movement
		movement := components.Movement.Get(unitEntry)
		movement.Paused = false
	}
}

func findClosestMine(ecs *ecs.ECS, unit *donburi.Entry) *donburi.Entry {

	var closestMine *donburi.Entry = nil

	unitPosition := components.Position.Get(unit)
	minDistance := math.MaxFloat64

	mineQuery.Each(ecs.World, func(mineEntry *donburi.Entry) {
		minePosition := components.Position.Get(mineEntry)
		mineData := components.Mine.Get(mineEntry)
		if mineData.Busy {
			return
		}
		distance := math.Abs(float64(minePosition.GridX-unitPosition.GridX)) + math.Abs(float64(minePosition.GridY-unitPosition.GridY))
		if distance < minDistance {
			minDistance = distance
			closestMine = mineEntry
		}
	})
	return closestMine
}

func startMovingToMine(ecs *ecs.ECS, unitEntry, mineEntry *donburi.Entry) {
	unitData := components.Unit.Get(unitEntry)
	mineData := components.Mine.Get(mineEntry)
	minePosition := components.Position.Get(mineEntry)

	unitData.TargetEntry = mineEntry

	if mineData.Busy {
		fmt.Println("Mine is busy, finding a waiting spot...")
		waitingPos := findFreeAdjacentTile(ecs, minePosition)
		if waitingPos != nil {
			fmt.Printf("Found waiting spot at (%d, %d)\n", waitingPos.X, waitingPos.Y)
			setNewDestination(unitEntry, waitingPos.X, waitingPos.Y)
			unitData.TargetPos = *waitingPos
			unitData.State = components.MovingToMineWaitingSpot
		} else {
			fmt.Println("No free adjacent tile found, staying idle.")
			unitData.State = components.Idle
		}
	} else {
		fmt.Println("Mine is free, moving directly to mine.")
		mineData.Busy = true // Claim the mine
		mineData.ClaimedBy = unitEntry
		setNewDestination(unitEntry, minePosition.GridX, minePosition.GridY)
		unitData.TargetPos = components.Point{X: minePosition.GridX, Y: minePosition.GridY}
		unitData.State = components.MovingToMine
	}
}

func startReturningToBase(ecs *ecs.ECS, unitEntry, baseEntry *donburi.Entry) {
	unitData := components.Unit.Get(unitEntry)
	basePosition := components.Position.Get(baseEntry)

	fmt.Printf("%s starting return to base.\n", unitData.Name)

	setNewDestination(unitEntry, basePosition.GridX, basePosition.GridY)

	unitData.State = components.ReturningToBase
	unitData.TargetEntry = baseEntry
	unitData.TargetPos = components.Point{X: basePosition.GridX, Y: basePosition.GridY}
}

func findFreeAdjacentTile(ecs *ecs.ECS, basePosition *components.PositionData) *components.Point {
	offsets := []components.Point{{X: 0, Y: -1}, {X: 1, Y: 0}, {X: 0, Y: 1}, {X: -1, Y: 0}} // Up, Right, Down, Left

	for _, offset := range offsets {
		tryPos := components.Point{X: basePosition.GridX + offset.X, Y: basePosition.GridY + offset.Y}
		isOccupied := false

		// Check if any unit is at or heading to this position
		robotAIQuery.Each(ecs.World, func(entry *donburi.Entry) {
			pos := components.Position.Get(entry)
			unit := components.Unit.Get(entry)

			// Check current position
			if pos.GridX == tryPos.X && pos.GridY == tryPos.Y {
				isOccupied = true
				return
			}

			// Check target position for other robots heading to any waiting spot
			isWaiting := unit.State == components.MovingToWaitingSpot ||
				unit.State == components.WaitingForBase ||
				unit.State == components.MovingToMineWaitingSpot ||
				unit.State == components.WaitingForMine

			if isWaiting && unit.TargetPos.X == tryPos.X && unit.TargetPos.Y == tryPos.Y {
				isOccupied = true
				return
			}
		})

		if !isOccupied {
			return &tryPos
		}
	}

	return nil // No free tile found
}

func findClosestBase(ecs *ecs.ECS, unit *donburi.Entry) *donburi.Entry {
	var closestBuilding *donburi.Entry = nil

	unitPosition := components.Position.Get(unit)
	minDistance := math.MaxFloat64

	buildingQuery.Each(ecs.World, func(buildingEntry *donburi.Entry) {
		buildingPosition := components.Position.Get(buildingEntry)
		distance := math.Abs(float64(buildingPosition.GridX-unitPosition.GridX)) + math.Abs(float64(buildingPosition.GridY-unitPosition.GridY))
		if distance < minDistance {
			minDistance = distance
			closestBuilding = buildingEntry
		}
	})
	return closestBuilding
}
