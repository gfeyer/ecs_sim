package game

import (
	"log"
	"math/rand"

	"github.com/gfeyer/ecs_sim/internal/components"
	"github.com/gfeyer/ecs_sim/internal/constants"
	"github.com/gfeyer/ecs_sim/internal/systems"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

type Game struct {
	ecs *ecs.ECS
}

func NewGame(screenWidth, screenHeight int) *Game {
	// Create a new ECS world.
	world := donburi.NewWorld()
	ecs := ecs.NewECS(world)

	// Create singleton entities.
	settingsEntity := ecs.World.Entry(ecs.World.Create(components.Settings))
	donburi.SetValue(settingsEntity, components.Settings, components.SettingsData{
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
	})

	assetsEntity := ecs.World.Entry(ecs.World.Create(components.Assets))
	donburi.SetValue(assetsEntity, components.Assets, components.AssetsData{
		Assets: make(map[string]components.AssetInfo),
	})

	debugEntity := ecs.World.Entry(ecs.World.Create(components.Debug))
	donburi.SetValue(debugEntity, components.Debug, components.DebugData{})

	eventQueueEntity := ecs.World.Entry(ecs.World.Create(components.EventQueue))
	donburi.SetValue(eventQueueEntity, components.EventQueue, components.EventQueueData{
		SpawnRequests: make([]components.SpawnRequestData, 0),
	})

	// Create the factory
	factory, err := systems.NewFactory(ecs)
	if err != nil {
		log.Fatalf("failed to create factory: %v", err)
	}

	// Load assets from the map
	if err := factory.LoadMapAssets("assets/test_map.tmx"); err != nil {
		log.Fatalf("failed to load map assets: %v", err)
	}

	// Create the world and all its tile entities.
	if err := factory.CreateWorld("assets/test_map.tmx"); err != nil {
		log.Fatalf("failed to create world: %v", err)
	}

	// Print the initial grid state for debugging.
	// systems.PrintDebugGrid(ecs)
	// systems.PrintDebugPropertyGrid(ecs, "IsWalkable")

	// Create the camera entity.
	cameraEntity := ecs.World.Entry(ecs.World.Create(components.Camera))
	// Center the camera on the base.
	baseGridX := constants.GridSizeX / 2
	baseGridY := constants.GridSizeY / 2
	isoX := (baseGridX - baseGridY) * (constants.TileWidth / 2)
	isoY := (baseGridX + baseGridY) * (constants.TileHeight / 2)

	// Adjust for the map's global offset to get the final world coordinates.
	worldX := float64(isoX)
	worldY := float64(isoY)

	donburi.SetValue(cameraEntity, components.Camera, components.CameraData{
		X:         worldX,
		Y:         worldY - 400,
		Speed:     5, // Adjust speed as needed
		Zoom:      1.0,
		ZoomSpeed: 0.1,
	})

	// Register systems.
	ecs.AddSystem(systems.UpdateInput)
	ecs.AddSystem(systems.UpdateCamera)
	ecs.AddSystem(systems.UpdateRobotAI)
	ecs.AddSystem(systems.UpdateLabels)
	ecs.AddSystem(systems.UpdatePathfinding)
	ecs.AddSystem(systems.UpdateMovement)
	eventSystem := systems.NewEventSystem(factory)
	ecs.AddSystem(eventSystem.Update)

	// Create the base
	systems.Spawn(ecs, constants.TypeBase, constants.GridSizeX/2, constants.GridSizeY/2)
	systems.Spawn(ecs, constants.TypeBase, constants.GridSizeX/2+4, constants.GridSizeY/2+4)

	// Spawn random robots
	for i := 0; i < 15; i++ {
		systems.Spawn(ecs, constants.TypeRobot, rand.Intn(constants.GridSizeX), rand.Intn(constants.GridSizeY))
	}

	// Spawn random mines, avoiding the base location.
	baseX := constants.GridSizeX / 2
	baseY := constants.GridSizeY / 2
	for i := 0; i < 25; i++ {
		var x, y int
		for {
			x = rand.Intn(constants.GridSizeX)
			y = rand.Intn(constants.GridSizeY)
			if x != baseX || y != baseY && x != baseX+4 || y != baseY+4 {
				break
			}
		}
		systems.Spawn(ecs, constants.TypeMine, x, y)
	}

	return &Game{
		ecs: ecs,
	}
}

func (g *Game) Update() error {
	g.ecs.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	systems.DrawRenderables(g.ecs, screen)
	systems.DrawLabels(g.ecs, screen)
	systems.DrawDebugText(g.ecs, screen)
	// systems.DrawDebug(g.ecs, screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	settings, ok := components.Settings.First(g.ecs.World)
	if !ok {
		// Fallback to a default resolution if settings are not found.
		return 800, 600
	}
	settingsData := components.Settings.Get(settings)
	return settingsData.ScreenWidth, settingsData.ScreenHeight
}
