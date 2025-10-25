package systems

import (
	"fmt"
	"image"

	"github.com/beefsack/go-astar"
	"github.com/gfeyer/ecs_sim/internal/components"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	"github.com/yohamta/donburi/filter"
)

// AStarTile represents a tile in the A* pathfinding grid.
type AStarTile struct {
	world     *components.WorldData
	ecs       *ecs.ECS
	tileCache map[image.Point]*AStarTile
	X, Y      int
}

// PathNeighbors returns the neighbors of a tile for pathfinding.
func (t *AStarTile) PathNeighbors() []astar.Pather {
	neighbors := []astar.Pather{}
	offsets := []image.Point{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for _, offset := range offsets {
		nx, ny := t.X+offset.X, t.Y+offset.Y

		if nx < 0 || nx >= t.world.Width || ny < 0 || ny >= t.world.Height {
			continue
		}

		// Check if the target tile is walkable.
		targetTileEntities := t.world.EntityMap[ny][nx]
		if len(targetTileEntities) == 0 {
			continue
		}

		var isWalkable bool
		for _, entity := range targetTileEntities {
			entry := t.ecs.World.Entry(entity)
			if entry.HasComponent(components.Tile) {
				tileData := components.Tile.Get(entry)
				if tileData.IsWalkable {
					isWalkable = true
					break
				}
			}
		}

		if !isWalkable {
			continue
		}

		// Use the cache to ensure a unique tile for each coordinate.
		p := image.Point{X: nx, Y: ny}
		if _, ok := t.tileCache[p]; !ok {
			t.tileCache[p] = &AStarTile{
				world:     t.world,
				ecs:       t.ecs,
				tileCache: t.tileCache,
				X:         nx,
				Y:         ny,
			}
		}
		neighbors = append(neighbors, t.tileCache[p])
	}
	return neighbors
}

// PathNeighborCost returns the cost of moving to a neighbor tile.
func (t *AStarTile) PathNeighborCost(to astar.Pather) float64 {
	return 1.0 // All moves have a uniform cost.
}

// PathEstimatedCost returns the estimated cost to get to the destination.
func (t *AStarTile) PathEstimatedCost(to astar.Pather) float64 {
	toT := to.(*AStarTile)
	dx := toT.X - t.X
	dy := toT.Y - t.Y
	return float64(dx*dx + dy*dy) // Manhattan distance squared
}

var pathfindingQuery = donburi.NewQuery(
	filter.And(
		filter.Contains(components.Position, components.Destination),
		filter.Not(filter.Contains(components.Path)),
	),
)

// UpdatePathfinding finds paths for entities with a destination.
func UpdatePathfinding(ecs *ecs.ECS) {
	worldEntry, ok := components.World.First(ecs.World)
	if !ok {
		return
	}
	worldData := components.World.Get(worldEntry)

	pathfindingQuery.Each(ecs.World, func(entry *donburi.Entry) {
		pos := components.Position.Get(entry)
		dest := components.Destination.Get(entry)

		// Initialize the tile cache for this pathfinding operation.
		tileCache := make(map[image.Point]*AStarTile)

		from := &AStarTile{world: worldData, ecs: ecs, tileCache: tileCache, X: pos.GridX, Y: pos.GridY}
		to := &AStarTile{world: worldData, ecs: ecs, tileCache: tileCache, X: dest.X, Y: dest.Y}

		// Pre-populate the cache with the start and end points.
		tileCache[image.Point{X: from.X, Y: from.Y}] = from
		tileCache[image.Point{X: to.X, Y: to.Y}] = to

		fmt.Printf("Pathfinding: From (%d, %d) to (%d, %d)\n", pos.GridX, pos.GridY, dest.X, dest.Y)
		path, _, found := astar.Path(from, to)
		if !found {
			fmt.Println("Pathfinding: Path not found.")
		} else {
			fmt.Printf("Pathfinding: Path found: %t, Path length: %d\n", found, len(path))
		}

		if found {
			points := []image.Point{}
			for _, p := range path {
				tile := p.(*AStarTile)
				points = append(points, image.Point{X: tile.X, Y: tile.Y})
			}

			// The path is from destination to start, so we need to reverse it.
			for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
				points[i], points[j] = points[j], points[i]
			}

			// Set the path, excluding the starting point.
			donburi.Add(entry, components.Path, &components.PathData{Points: points[1:]})
		}

		// Always remove the destination component after attempting to find a path.
		entry.RemoveComponent(components.Destination)
	})
}
