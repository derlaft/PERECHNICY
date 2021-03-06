package game

import (
	"math/rand"
	. "util"
)

type Bear struct {
	hp   uint
	tick uint
}

func NewBear() Entity {
	return &Bear{
		hp: 42,
	}
}

func (b *Bear) Health() uint {
	return b.hp
}

func (b *Bear) OnDamage(c *Control, dmg uint) {
	b.hp = uint(Max(0, int(b.hp)-int(dmg)))
}

func (b *Bear) Tick(c *Control) {
	b.tick = (b.tick + 1) % 250

	if b.tick == 0 {
		c.Move(c.Location.Add(Point{
			1 - rand.Int63n(3),
			1 - rand.Int63n(3),
		}))
	}
}

func (b *Bear) Byte(c *Control) byte {
	return ENTITY_BEAR
}

func (b *Bear) OnColission(c1 *Control, c2 *Control) {
	// bear is afraid to do anything
}
