package components

import "github.com/yohamta/donburi"

// LabelData holds the text to be displayed above an entity.
type LabelData struct {
	Text    string
	OffsetX int
	OffsetY int
}

var Label = donburi.NewComponentType[LabelData]()
