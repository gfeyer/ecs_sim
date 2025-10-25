package components

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"
)

type BaseData struct {
	Resources float64
	Busy      bool
	ClaimedBy *donburi.Entry
}

var Base = donburi.NewComponentType[BaseData]()

var BaseQuery = donburi.NewQuery(filter.Contains(Base))
