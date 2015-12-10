package chunk

import (
	. "../stuff"
	"math/rand"
	"sync"
)

const (
	CHUNK_SIZE = 16
)

type Map struct {
	Chunks map[Point]*Chunk
	Size   int
	sync.RWMutex
}

type Chunk struct {
	Data [CHUNK_SIZE][CHUNK_SIZE]byte
	sync.RWMutex
}

func NewMap(size int) (m *Map) {
	m = &Map{Size: size}
	m.Chunks = make(map[Point]*Chunk)
	return m
}

func ChunkAbs(a int64) int64 {
	a = a % CHUNK_SIZE
	if a < 0 {
		return -a
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
		return m.GenChunk(pt)
	}

	return chunk
}

// generate chunk
func (m *Map) GenChunk(pt Point) *Chunk {
	chunk := &Chunk{}

	border := pt.Dist(Point{0, 0}) >= m.Size

	for i, row := range chunk.Data {
		for j := range row {
			if !border {
				// no need to lock chunk here: it's not linked to Chunks
				chunk.Data[i][j] = (1 - byte(float64(rand.Intn(100))*0.011))
			} else {
				chunk.Data[i][j] = 9
			}
		}
	}
	m.Lock()
	m.Chunks[pt] = chunk
	m.Unlock()
	return chunk
}
