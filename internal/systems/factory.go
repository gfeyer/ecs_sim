package systems

import (
	"fmt"
	"image"
	"math/rand"
	"net/http"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gfeyer/ecs_sim/internal/components"
	"github.com/gfeyer/ecs_sim/internal/constants"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/lafriks/go-tiled"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/donburi/filter"
	"github.com/yohamta/donburi/query"
)

// Factory is responsible for creating all game entities and resources.
// It encapsulates dependencies like the ECS world and asset manager to simplify
// entity creation and ensure consistency.
type Factory struct {
	ecs            *ecs.ECS
	assets         *components.AssetsData
	robotNames     []string
	availableNames []string
}

// NewFactory creates a new Factory instance.
// It retrieves the Assets resource from the ECS world, which is a required dependency.
func NewFactory(ecs *ecs.ECS) (*Factory, error) {
	assetsEntry, ok := query.NewQuery(filter.Contains(components.Assets)).First(ecs.World)
	if !ok {
		return nil, fmt.Errorf("assets resource not found")
	}
	assets := components.Assets.Get(assetsEntry)

	f := &Factory{
		ecs:    ecs,
		assets: assets,
		robotNames: []string{
			"Glitch", "404", "sudo", "Ping", "Pong",
			"Root", "Null", "NaN", "Segfault", "Kernel",
			"Byte", "Bit", "Cache", "Cookie", "Cron",
			"Daemon", "Fuzz", "Hex", "Jolt", "Zorp",
		},
	}

	f.availableNames = make([]string, len(f.robotNames))
	copy(f.availableNames, f.robotNames)
	rand.Shuffle(len(f.availableNames), func(i, j int) {
		f.availableNames[i], f.availableNames[j] = f.availableNames[j], f.availableNames[i]
	})

	return f, nil
}

// CreateRobot creates the player entity with all its components.
func (f *Factory) CreateRobot(x, y int) (*donburi.Entry, error) {
	robot := f.ecs.World.Entry(f.ecs.World.Create(
		components.Unit,
		components.Position,
		components.Renderable,
		components.Movement,
		components.Label,
		// components.PlayerControlled,
	))

	var robotName string
	if len(f.availableNames) > 0 {
		robotName = f.availableNames[0]
		f.availableNames = f.availableNames[1:]
	} else {
		robotName = fmt.Sprintf("Bot %d", len(f.robotNames)+1)
	}

	donburi.Add(robot, components.Unit, &components.UnitData{
		Name:            robotName,
		State:           components.Idle,
		LoadSpeed:       0.002 + rand.Float64()*0.009,
		MaxLoadCapacity: 1 + rand.Float64()*2.0,
	})
	donburi.Add(robot, components.Position, &components.PositionData{
		GridX: x,
		GridY: y,
	})

	donburi.Add(robot, components.Movement, &components.MovementData{
		MoveSpeed: time.Second,
	})

	// Get the robot assets from the pre-loaded assets.
	assetRobotRight, ok := f.assets.Assets[constants.AssetRobotRight]
	if !ok {
		return robot, fmt.Errorf("%s asset not found", constants.AssetRobotRight)
	}

	assetRobotLeft, ok := f.assets.Assets[constants.AssetRobotLeft]
	if !ok {
		return robot, fmt.Errorf("%s asset not found", constants.AssetRobotLeft)
	}

	assetRobotUp, ok := f.assets.Assets[constants.AssetRobotUp]
	if !ok {
		return robot, fmt.Errorf("%s asset not found", constants.AssetRobotUp)
	}

	assetRobotDown, ok := f.assets.Assets[constants.AssetRobotDown]
	if !ok {
		return robot, fmt.Errorf("%s asset not found", constants.AssetRobotDown)
	}

	donburi.Add(robot, components.Renderable, &components.RenderableData{
		Parts: map[string]components.RenderablePart{
			constants.AssetRobotRight: {Image: assetRobotRight.Image, SourceRect: assetRobotRight.SourceRect},
			constants.AssetRobotLeft:  {Image: assetRobotLeft.Image, SourceRect: assetRobotLeft.SourceRect},
			constants.AssetRobotUp:    {Image: assetRobotUp.Image, SourceRect: assetRobotUp.SourceRect},
			constants.AssetRobotDown:  {Image: assetRobotDown.Image, SourceRect: assetRobotDown.SourceRect},
		},
		CurrentPart: constants.AssetRobotRight,
		DrawOrder:   99, // Render units on top of everything
	})

	donburi.Add(robot, components.Label, &components.LabelData{
		Text:    fmt.Sprintf("%s\nOre: %.2f", robotName, 0.0),
		OffsetX: 0,
		OffsetY: -5,
	})

	return robot, nil
}

func (f *Factory) CreateMine(x, y int) (*donburi.Entry, error) {
	mine := f.ecs.World.Entry(f.ecs.World.Create(
		components.Mine,
		components.Position,
		components.Renderable,
		components.Label,
	))

	resources := 10 + rand.Float64()*10

	donburi.Add(mine, components.Mine, &components.MineData{Resources: resources})
	donburi.Add(mine, components.Position, &components.PositionData{
		GridX: x,
		GridY: y,
	})
	donburi.Add(mine, components.Label, &components.LabelData{
		Text:    fmt.Sprintf("%.2f", resources),
		OffsetX: 64,
		OffsetY: 0,
	})

	// determine asset for this mine
	assetName := fmt.Sprintf("mine%d", rand.Intn(5)+1)
	asset, ok := f.assets.Assets[assetName]
	if !ok {
		return mine, fmt.Errorf("%s asset not found", assetName)
	}

	donburi.Add(mine, components.Renderable, &components.RenderableData{
		Parts: map[string]components.RenderablePart{
			assetName: {Image: asset.Image, SourceRect: asset.SourceRect},
		},
		CurrentPart: assetName,
		DrawOrder:   10, // Render on top of tiles
	})

	return mine, nil
}

func (f *Factory) CreateBase(x, y int) (*donburi.Entry, error) {
	base := f.ecs.World.Entry(f.ecs.World.Create(
		components.Base,
		components.Position,
		components.Renderable,
		components.Label,
	))

	donburi.Add(base, components.Base, &components.BaseData{})
	donburi.Add(base, components.Position, &components.PositionData{
		GridX: x,
		GridY: y,
	})
	donburi.Add(base, components.Label, &components.LabelData{
		Text:    "Base\nOre: 0.0",
		OffsetX: 64,
		OffsetY: 0,
	})

	asset := f.assets.Assets[constants.AssetBase]

	donburi.Add(base, components.Renderable, &components.RenderableData{
		Parts: map[string]components.RenderablePart{
			constants.AssetBase: {Image: asset.Image, SourceRect: asset.SourceRect},
		},
		CurrentPart: constants.AssetBase,
		DrawOrder:   10, // Render on top of tiles
	})

	return base, nil
}

// LoadMapAssets parses a TMX file and loads all associated tileset images and
// asset information into the Assets resource.
func (f *Factory) LoadMapAssets(path string) error {
	var gameMap *tiled.Map
	var err error
	if runtime.GOOS == "js" {
		resp, err := http.Get(path)
		if err != nil {
			return fmt.Errorf("failed to fetch map file: %w", err)
		}
		defer resp.Body.Close()
		gameMap, err = tiled.LoadReader(filepath.Dir(path), resp.Body)
	} else {
		gameMap, err = tiled.LoadFile(path)
	}
	if err != nil {
		return fmt.Errorf("failed to load map file: %w", err)
	}

	dir := filepath.Dir(path)
	for _, tileset := range gameMap.Tilesets {
		imgPath := filepath.Join(dir, tileset.Image.Source)
		var img *ebiten.Image
		var err error
		if runtime.GOOS == "js" {
			resp, err := http.Get(imgPath)
			if err != nil {
				return fmt.Errorf("failed to fetch tileset image '%s': %w", imgPath, err)
			}
			defer resp.Body.Close()
			img, _, err = ebitenutil.NewImageFromReader(resp.Body)
		} else {
			img, _, err = ebitenutil.NewImageFromFile(imgPath)
		}
		if err != nil {
			return fmt.Errorf("failed to load tileset image '%s': %w", imgPath, err)
		}

		for _, tileDef := range tileset.Tiles {
			assetName := tileDef.Properties.GetString("Asset")
			if assetName != "" {
				tileX := (int(tileDef.ID) % tileset.Columns) * tileset.TileWidth
				tileY := (int(tileDef.ID) / tileset.Columns) * tileset.TileHeight
				srcRect := image.Rect(tileX, tileY, tileX+tileset.TileWidth, tileY+tileset.TileHeight)

				f.assets.Assets[assetName] = components.AssetInfo{
					Image:      img,
					SourceRect: srcRect,
				}
			}
		}
	}
	return nil
}

// CreateWorld initializes the tile entities and the World resource from a TMX file.
// It should be called after LoadMapAssets has populated the necessary assets.
func (f *Factory) CreateWorld(path string) error {
	var gameMap *tiled.Map
	var err error
	if runtime.GOOS == "js" {
		resp, err := http.Get(path)
		if err != nil {
			return fmt.Errorf("failed to fetch map file: %w", err)
		}
		defer resp.Body.Close()
		gameMap, err = tiled.LoadReader(filepath.Dir(path), resp.Body)
	} else {
		gameMap, err = tiled.LoadFile(path)
	}
	if err != nil {
		return fmt.Errorf("failed to load map file: %w", err)
	}

	tilesetImages, propertiesMap, err := f.loadTilesets(gameMap, path)
	if err != nil {
		return err
	}

	entityMap := f.createTileEntities(gameMap, tilesetImages, propertiesMap)

	// Create the world resource entity that holds the lookup grid.
	worldEntity := f.ecs.World.Entry(f.ecs.World.Create(components.World))
	donburi.SetValue(worldEntity, components.World, components.WorldData{
		Width:     gameMap.Width,
		Height:    gameMap.Height,
		EntityMap: entityMap,
	})

	return nil
}

// loadTilesets pre-processes all tilesets in the map to load their images and properties.
func (f *Factory) loadTilesets(gameMap *tiled.Map, mapPath string) (map[string]*ebiten.Image, map[uint32]tileProperties, error) {
	tilesetImages := make(map[string]*ebiten.Image)
	propertiesMap := make(map[uint32]tileProperties)
	dir := filepath.Dir(mapPath)

	for _, tileset := range gameMap.Tilesets {
		imgPath := filepath.Join(dir, tileset.Image.Source)
		var img *ebiten.Image
		var err error
		if runtime.GOOS == "js" {
			resp, err := http.Get(imgPath)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to fetch tileset image '%s': %w", imgPath, err)
			}
			defer resp.Body.Close()
			img, _, err = ebitenutil.NewImageFromReader(resp.Body)
		} else {
			img, _, err = ebitenutil.NewImageFromFile(imgPath)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("failed to load tileset image '%s': %w", imgPath, err)
		}
		tilesetImages[tileset.Name] = img

		for _, tileDef := range tileset.Tiles {
			gid := tileset.FirstGID + tileDef.ID
			propertiesMap[gid] = tileProperties{
				IsWalkable: tileDef.Properties.GetBool(constants.IsWalkable),
			}
		}
	}

	return tilesetImages, propertiesMap, nil
}

// createTileEntities iterates through map layers and creates an entity for each tile.
func (f *Factory) createTileEntities(gameMap *tiled.Map, tilesetImages map[string]*ebiten.Image, propertiesMap map[uint32]tileProperties) [][][]donburi.Entity {
	entityMap := make([][][]donburi.Entity, gameMap.Height)
	for i := range entityMap {
		entityMap[i] = make([][]donburi.Entity, gameMap.Width)
	}

	for layerIndex, layer := range gameMap.Layers {
		for i, tile := range layer.Tiles {
			if tile.Nil {
				continue
			}

			x := i % gameMap.Width
			y := i / gameMap.Width

			layerTile, err := gameMap.TileGIDToTile(tile.ID + tile.Tileset.FirstGID)
			if err != nil || layerTile == nil || layerTile.Tileset == nil {
				continue
			}

			tilesetImage := tilesetImages[layerTile.Tileset.Name]
			if tilesetImage == nil {
				continue
			}

			tileX := (int(tile.ID) % layerTile.Tileset.Columns) * layerTile.Tileset.TileWidth
			tileY := (int(tile.ID) / layerTile.Tileset.Columns) * layerTile.Tileset.TileHeight
			srcRect := image.Rect(tileX, tileY, tileX+layerTile.Tileset.TileWidth, tileY+layerTile.Tileset.TileHeight)

			entity := f.ecs.World.Entry(f.ecs.World.Create(components.Position, components.Renderable, components.Tile))
			donburi.Add(entity, components.Position, &components.PositionData{GridX: x, GridY: y})
			donburi.Add(entity, components.Renderable, &components.RenderableData{
				Parts: map[string]components.RenderablePart{
					"default": {Image: tilesetImage, SourceRect: srcRect},
				},
				CurrentPart: "default",
				DrawOrder:   layerIndex,
			})

			props := propertiesMap[tile.ID+tile.Tileset.FirstGID]
			donburi.Add(entity, components.Tile, &components.TileData{
				ID:         tile.ID,
				IsWalkable: props.IsWalkable,
			})

			entityMap[y][x] = append(entityMap[y][x], entity.Entity())
		}
	}

	return entityMap
}

type tileProperties struct {
	IsWalkable bool
}
