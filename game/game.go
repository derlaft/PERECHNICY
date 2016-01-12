package game

import (
	. "../chunk"
	. "../event"
	. "../stuff"
	"fmt"
	"sync"
)

type Entity interface {
	Health() uint
	OnDamage(*Control, uint)
	Tick(*Control)
	Byte(*Control) byte
	OnColission(*Control, *Control)
}

type Control struct {
	Game      *Game
	Location  Point
	Entity    Entity
	Destroyed bool
}

type Entities []*Control

type Game struct {
	World    *Map
	Entities Entities
	sync.RWMutex
	EvHandler EventInterface
}

type EventInterface interface {
	IsSolid(byte) bool
	SendEvent(int, Point, *Control)
}

func NewEntity(game *Game, l Point, entity Entity) (*Control, bool) {
	ctl := &Control{
		Game:     game,
		Location: Point{},
		Entity:   entity,
	}

	game.Entities = append(game.Entities, ctl)

	if !ctl.Move(l) {
		return nil, false
	}

	return ctl, true
}

func (g *Game) IsBlockSolid(pt Point) bool {
	return g.EvHandler.IsSolid(g.World.At(pt))
}

func (g *Game) EntityAt(pt Point) *Control {
	g.RLock()
	defer g.RUnlock()
	for _, e := range g.Entities {
		if e.Location == pt {
			return e
		}
	}

	return nil
}

func (g *Game) At(pt Point) byte {
	g.RLock()
	entity := g.EntityAt(pt)
	g.RUnlock()
	if entity != nil {
		return entity.Byte()
	}
	return g.World.At(pt)
}

func (g *Game) Tick() {
	g.RLock()
	for _, entity := range g.Entities {
		g.RUnlock()
		entity.Tick()
		g.RLock()
	}
	g.RUnlock()
}

func (g *Game) deleteEntity(e *Control) {
	i := 0
	a := g.Entities

	for i < len(a) && a[i] != e {
		i++
	}

	if i < len(a) && a[i] == e {

		if len(a) > 1 {
			a[i] = a[len(a)-1]
			a[len(a)-1] = nil
			a = a[:len(a)-1]
		} else {
			a = make([]*Control, 0)
		}
	}

	g.Entities = a
}

// Yes, the only supported movement type is teleporting, lol
func (c *Control) Move(next Point) bool {

	c.Game.RLock()
	entityAt := c.Game.EntityAt(next)
	c.Game.RUnlock()

	if !c.Game.IsBlockSolid(next) && entityAt == nil {
		c.Game.Lock()
		c.Game.Unlock()
		c.Location = next

		c.Game.EvHandler.SendEvent(EVENT_ENTITY_ON, next, c)

		return true
	} else if entityAt != nil {
		//collision
		c.Entity.OnColission(c, entityAt)
		return false
	} else {
		//solid block
		return false
	}
}

func (c *Control) Destroy() {
	c.Destroyed = true
	c.Game.Lock()
	c.Game.Delete(c)
	c.Game.Unlock()
}

func (g *Game) Delete(c *Control) {
	g.deleteEntity(c)
	c.Destroyed = true
}

func (e *Control) Byte() byte {
	return e.Entity.Byte(e)
}

func (e *Control) Tick() {
	//TODO: check why it's needed
	//howto reproduce: die near another bot
	if e == nil || e.Destroyed {
		fmt.Println("WAAT")
		return
	}
	e.Entity.Tick(e)
}

func (g *Game) Dump(from, to Point) (out string) {

	from, to = MinMaxPoint(from, to)

	for i := from.Y; i <= to.Y; i++ {
		for j := from.X; j <= to.X; j++ {

			pt := Point{j, i}

			entity := g.EntityAt(pt)
			tile := g.World.At(pt)

			if entity != nil {
				out += " " + string(entity.Byte())
			} else {
				out += fmt.Sprintf(" %d", tile)
			}
		}
		out += "\n"
	}

	return
}

func NewGame(w *Map, h EventInterface) *Game {
	g := &Game{World: w, Entities: make(Entities, 0), EvHandler: h}
	g.World.At(Point{0, 0})
	return g
}
