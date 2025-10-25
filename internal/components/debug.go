package components

import (
	"time"

	"github.com/yohamta/donburi"
)

// DebugData holds performance metrics for display.
type DebugData struct {
	LastUpdated time.Time
	CachedMsg   string
}

var Debug = donburi.NewComponentType[DebugData]()
