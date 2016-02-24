package game

import (
	"fmt"
	"sync"
	. "util"
)

const (
	CHUNK_SIZE = 8
)

var (
	BIOMES = map[byte]Generator{
		BIOME_1: doRoomChunk,
		BIOME_2: doRoomChunk,
	}
)

const (
	BIOME_1 = iota
	BIOME_2
)

type Generator func(*Map, *Chunk, Point)

type Map struct {
	Chunks map[Point]*Chunk
	Size   int
	Seed   int
	sync.RWMutex
}

type Chunk struct {
	Data [CHUNK_SIZE][CHUNK_SIZE]byte
	sync.RWMutex
}

func NewMap(size int, seed int) (m *Map) {
	m = &Map{Size: size, Seed: seed}
	m.Chunks = make(map[Point]*Chunk)
	return m
}

func ChunkAbs(a int64) int64 {
	a = a % CHUNK_SIZE
	if a < 0 {
		return CHUNK_SIZE + a
	}
	return a
}

func ChunkCrd(pt Point) Point {
	ret := Point{pt.X / CHUNK_SIZE, pt.Y / CHUNK_SIZE}
	if pt.X < 0 {
		ret = ret.Add(Point{-1, 0})
	}
	if pt.Y < 0 {
		ret = ret.Add(Point{0, -1})
	}
	return ret
}

// get tile by coordinates
func (m *Map) At(pt Point) (b byte) {
	chunk := m.GetChunk(pt)
	chunk.RLock()
	b = chunk.Data[ChunkAbs(pt.X)][ChunkAbs(pt.Y)]
	chunk.RUnlock()
	return
}

// get tile by coordinates
func (m *Map) Set(pt Point, b byte) {
	chunk := m.GetChunk(pt)
	chunk.Lock()
	chunk.Data[ChunkAbs(pt.X)][ChunkAbs(pt.Y)] = b
	chunk.Unlock()
}

// get chunk by coordinates
func (m *Map) GetChunk(pt Point) *Chunk {
	return m.getChunk(ChunkCrd(pt))
}

// get chunk by chunk coordinates
func (m *Map) getChunk(pt Point) *Chunk {
	m.RLock()
	chunk, ok := m.Chunks[pt]
	m.RUnlock()
	if !ok {
		return GenChunk(m, pt)
	}

	return chunk
}

func (m *Map) OutOfBorder(pt Point) bool {
	return pt.Dist(Point{0, 0}) >= m.Size
}

// generate chunk
func GenChunk(m *Map, pt Point) *Chunk {

	chunk := &Chunk{}
	if !m.OutOfBorder(pt) {
		n2d := NewNoise2DContext(m.Seed)
		t := n2d.GetByte(pt.X, pt.Y, len(BIOMES))
		if t > 1 {
			fmt.Println("WAAT")
			fmt.Println("WAAT")
			fmt.Println("WAAT")
			fmt.Println("WAAT")
		}
		BIOMES[t](m, chunk, pt)
	} else {
		doWallChunk(chunk)
	}

	m.Lock()
	m.Chunks[pt] = chunk
	m.Unlock()

	return chunk
}

func doRoomChunk(m *Map, c *Chunk, pt Point) {
	n2d := NewNoise2DContext(m.Seed)

	for i, row := range c.Data {
		for j := range row {
			b := n2d.GetByte(pt.X+int64(i), pt.Y+int64(j), 64)
			r := byte(0)

			if b >= 30 && b <= 40 || b >= 50 && b <= 60 {
				r = TILE_WALL
			}

			switch b {
			case 1, 3, 9, 10, 11:
				r = TILE_BONES
			case 12, 15, 17, 19:
				r = TILE_MUSHROOM
			case 13:
				r = ENTITY_BEAR
			case 20:
				r = ENTITY_KTULHU
			}

			c.Data[i][j] = r
		}
	}

}

func emptyChunk(m *Map, c *Chunk, pt Point) {
}

func doWallChunk(chunk *Chunk) {
	for i, row := range chunk.Data {
		for j := range row {
			chunk.Data[i][j] = 9
		}
	}
}
