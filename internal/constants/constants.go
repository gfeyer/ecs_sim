package constants

const (
	ScreenWidth  = 1600
	ScreenHeight = 980
)

const (
	GridSizeX = 32
	GridSizeY = 32
)

const (
	TileWidth      = 128
	TileHeight     = 64
	Tile64OffsetX  = 32 // this is 32 because the tile is a diamond and my tiles are 64x64 square
	Tile64OffsetY  = 0
	Tile128OffsetX = 0
	Tile128OffsetY = -64
)

const (
	IsWalkable = "IsWalkable"
)

type ItemType int

const (
	TypeNone ItemType = iota
	TypeRobot
	TypeMine
	TypeBase
)

const AssetRobotRight = "robot_right"
const AssetRobotLeft = "robot_left"
const AssetRobotUp = "robot_up"
const AssetRobotDown = "robot_down"

const AssetMine = "mine"
const AssetBase = "base"
