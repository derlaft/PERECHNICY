package game

import (
	"math/rand"
	. "util"
)

type Ktulhu struct {
	hp   uint
	tick uint
}

func NewKtulhu() Entity {
	return &Ktulhu{
		hp: 4200,
	}
}

func (b *Ktulhu) Health() uint {
	return b.hp
}

func (b *Ktulhu) OnDamage(c *Control, dmg uint) {
	b.hp = uint(Max(0, int(b.hp)-int(dmg)))
}

func (b *Ktulhu) Tick(c *Control) {
	b.tick = (b.tick + 1) % 250

	if b.tick == 0 {
		c.Move(c.Location.Add(Point{
			1 - rand.Int63n(3),
			1 - rand.Int63n(3),
		}))
	}
}

func (b *Ktulhu) Byte(c *Control) byte {
	return ENTITY_KTULHU
}

func (b *Ktulhu) OnColission(c1 *Control, c2 *Control) {
}
