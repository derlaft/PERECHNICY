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
	Game     *Game
	Location Point
	Entity   Entity
}

type Entities map[key]*Control

type Game struct {
	World    *Map
	Entities Entities
	sync.RWMutex
	EvHandler EventInterface
}

type key struct {
	Chunk    *Chunk
	Location Point
}

type EventInterface interface {
	IsSolid(byte) bool
	SendEvent(int, Point, *Control)
}

func NewEntity(g *Game, l Point, e Entity) (*Control, bool) {
	c := &Control{g, Point{}, e}

	//I suggest to use special spawner block for creating entities
	if !c.Move(l) {
		return nil, false
	}

	return c, true
}

func (g *Game) IsBlockSolid(pt Point) bool {
	return g.EvHandler.IsSolid(g.World.At(pt))
}

func (g *Game) EntityAt(pt Point) *Control {
	g.RLock()
	ret := g.Entities[key{g.World.GetChunk(pt), pt}]
	defer g.RUnlock()
	return ret
}

func (g *Game) At(pt Point) byte {
	g.RLock()
	entity := g.Entities[key{g.World.GetChunk(pt), pt}]
	g.RUnlock()
	if entity != nil {
		return entity.Byte()
	}
	return g.World.At(pt)
}

func (g *Game) Tick() {
	//TODO: do smth with dat locks
	//TODO: parallel tick handling
	//TODO: and make it parallel
	g.RLock()
	for _, entity := range g.Entities {
		g.RUnlock()
		entity.Tick()
		g.RLock()
	}
	g.RUnlock()
}

// Yes, the only supported movement type is teleporting, lol
func (c *Control) Move(next Point) bool {

	movedFrom := c.Game.World.GetChunk(c.Location)
	movedTo := c.Game.World.GetChunk(next)

	c.Game.RLock()
	entityAt := c.Game.EntityAt(next)
	c.Game.RUnlock()

	if !c.Game.IsBlockSolid(next) && entityAt == nil {
		c.Game.Lock()
		delete(c.Game.Entities, key{movedFrom, c.Location})
		c.Game.Entities[key{movedTo, next}] = c
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

func (e *Control) Byte() byte {
	return e.Entity.Byte(e)
}

func (e *Control) Tick() {
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
