package components

import "github.com/yohamta/donburi"

// SettingsData holds global game settings.
type SettingsData struct {
	ScreenWidth  int
	ScreenHeight int
}

var Settings = donburi.NewComponentType[SettingsData]()
