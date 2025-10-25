package components

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"
)

type MineData struct {
	Resources float64
	Busy      bool
	ClaimedBy *donburi.Entry
}

var Mine = donburi.NewComponentType[MineData]()
var MineQuery = donburi.NewQuery(filter.Contains(Mine))
