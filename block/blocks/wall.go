package blocks

const (
	TILE_WALL = 9
)

type Wall struct {
}

func (g Wall) Listeners() map[int]Listener {

	return map[int]Listener{}

}

func (g Wall) Solid() bool {
	return true
}
