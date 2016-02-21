package block

const (
	TILE_WALL     = 9
	TILE_MUSHROOM = 103
)

type Wall struct {
}

func (g Wall) Listeners() map[int]Listener {

	return map[int]Listener{}

}

func (g Wall) Solid() bool {
	return true
}
