package blocks

import (
	//	. "../../event"
	. "../../game"
	. "../../stuff"
)

const (
	TILE_VODKA = 1
)

type Vodka struct {
}

func (g Vodka) SolidFor(c *Control) bool {

	return false
}

func (g Vodka) Listeners() map[int]Listener {

	return nil //map[int]Listener{
	//		EVENT_ENTITY_ON: drainVodka,
	//}

}

func (g Vodka) Solid() bool {
	return false
}

func drainVodka(pt Point, c *Control) {
	c.Game.World.Set(c.Location, TILE_GROUND)
}
