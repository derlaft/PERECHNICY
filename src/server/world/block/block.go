package block

import (
	. "server/game"
	. "util"
)

var (
	Blocks = map[byte]Block{
		TILE_GROUND:   Ground{},
		TILE_WALL:     Wall{},
		TILE_BONES:    Ground{},
		TILE_MUSHROOM: Wall{},
		1:             Wall{},
		byte(99):      DummyEntity{},
		byte(98):      DummyEntity{},
	}
)

type Block interface {
	Solid() bool
	Listeners() map[int]Listener
}

type EventHandler struct {
}

func (e EventHandler) IsSolid(block byte) bool {
	handle, found := Blocks[block]
	if found {
		return handle.Solid()
	} else {
		panic("wat")
		return false
	}
}

func (e EventHandler) SendEvent(event int, pt Point, sender *Control) {

	block := sender.Game.World.At(pt)

	handle, found := Blocks[block]
	if !found {
		panic("wat")
		return
	}

	eventHandle, found := handle.Listeners()[event]
	if !found {
		return //in this case its ok
	}

	eventHandle(pt, sender)

}