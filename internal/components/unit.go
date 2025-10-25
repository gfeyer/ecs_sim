package components

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"
)

type State int

const (
	Idle State = iota
	MovingToMine
	Mining
	ReturningToBase
	Unloading
	WaitingForBase
	MovingToWaitingSpot
	MovingToMineWaitingSpot
	WaitingForMine
)

type Point struct {
	X, Y int
}

// UnitData holds information for a movable unit.
type UnitData struct {
	Name            string
	State           State
	TargetEntry     *donburi.Entry
	TargetPos       Point
	CurrentLoad     float64
	LoadSpeed       float64
	MaxLoadCapacity float64
}

var Unit = donburi.NewComponentType[UnitData]()
var UnitQuery = donburi.NewQuery(filter.Contains(Unit))

// PlayerControlled is a tag component to identify the entity controlled by the player.
type PlayerControlledData struct{}

var PlayerControlled = donburi.NewComponentType[PlayerControlledData]()
var PlayerQuery = donburi.NewQuery(filter.Contains(PlayerControlled))
