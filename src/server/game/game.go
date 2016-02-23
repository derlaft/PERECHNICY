package game

import (
	"fmt"
	"sync"
	. "util"
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

type entities []*Control

type Game struct {
	World    *Map
	Entities entities
	sync.RWMutex
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
	return Blocks.Is(g.World.At(pt), "Solid")
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

			bt := g.At(pt)
			out += fmt.Sprintf(" %d", bt)
		}
		out += "\n"
	}

	return
}

func NewGame(w *Map) *Game {
	g := &Game{World: w, Entities: make(entities, 0)}
	g.World.At(Point{0, 0})
	return g
}
