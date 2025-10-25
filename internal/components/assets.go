package components

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

// AssetInfo holds the rendering information for a tile that has been tagged as an asset.
type AssetInfo struct {
	Image      *ebiten.Image
	SourceRect image.Rectangle
}

// AssetsData holds all the loaded game assets, acting as a singleton resource.
type AssetsData struct {
	Assets map[string]AssetInfo
}

// Assets is the component type for the asset storage.
var Assets = donburi.NewComponentType[AssetsData]()
