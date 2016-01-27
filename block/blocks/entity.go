package blocks

import (
	. "../../game"
)

const ()

type DummyEntity struct {
}

func (g DummyEntity) SolidFor(c *Control) bool {

	return false
}

func (g DummyEntity) Listeners() map[int]Listener {

	return nil

}

func (g DummyEntity) Solid() bool {
	return true
}
