package components

import (
	"time"

	"github.com/yohamta/donburi"
)

// Movement is a component that manages the timing of an entity's movement.
type MovementData struct {
	MoveSpeed   time.Duration
	LastMove    time.Time
	TargetX     int
	TargetY     int
	IsMoving    bool
	Paused      bool
	PausedTime  time.Time
}

var Movement = donburi.NewComponentType[MovementData]()
