package blocks

import (
	. "../../game"
)

const (
	TILE_GROUND = 0
	TILE_BONES  = 107
)

type Ground struct {
}

func (g Ground) SolidFor(c *Control) bool {

	return false
}

func (g Ground) Listeners() map[int]Listener {

	return map[int]Listener{}

}

func (g Ground) Solid() bool {
	return false
}
