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

func (game *Game) NewEntity(l Point, entity Entity) (*Control, bool) {
	ctl := &Control{
		Game:     game,
		Location: Point{},
		Entity:   entity,
	}

	game.Entities = append(game.Entities, ctl)

	queried := make(map[Point]bool)
	toquery := make(map[Point]bool)

	toquery[l] = true

	for i := 0; i < 4; i++ {
		next := make(map[Point]bool)

		for ptx := range toquery {
			if !queried[ptx] {
				queried[ptx] = true
				if ctl.Move(ptx) {
					fmt.Println("SPAWN K")
					return ctl, true
				} else {
					for _, ptn := range []Point{
						ptx.Add(Point{+1, 0}),
						ptx.Add(Point{-1, 0}),
						ptx.Add(Point{0, +1}),
						ptx.Add(Point{0, -1}),
					} {
						next[ptn] = true
					}
				}
			}

		}
		fmt.Println("NEXTING")
		toquery = next

	}

	return nil, false
}

func (g *Game) IsBlock(quality string, pt Point) bool {
	return Blocks.Is(g.World.At(pt), quality)
}

func (g *Game) at(pt Point) (*Control, byte) {
	g.RLock()
	for _, entity := range g.Entities {
		if entity.Location == pt {
			g.RUnlock()
			return entity, entity.Byte()
		}
	}
	g.RUnlock()

	block := g.World.At(pt)
	if new_entity, exists := Entities[block]; exists {
		g.World.Set(pt, TILE_GROUND)
		g.NewEntity(pt, new_entity())
	}
	return nil, g.World.At(pt)
}

func (g *Game) EntityAt(pt Point) *Control {
	entity, _ := g.at(pt)
	return entity
}

func (g *Game) ByteAt(pt Point) byte {
	_, block := g.at(pt)
	return block
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

	entityAt := c.Game.EntityAt(next)

	if !c.Game.IsBlock(Solid, next) && entityAt == nil {
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

			out += fmt.Sprintf(" %d", g.ByteAt(Point{j, i}))
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
