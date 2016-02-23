package forms

import (
	. "client/grocessing"
	. "client/ui"
	"server/game"
)

const (
	CZ  = game.CHUNK_SIZE
	ROW = 8
)

var (
	Brushes []int
)

type editorForm struct {
	Map    []int
	cx, cy int
	cursor int
}

func init() {
	Forms[EDITOR_SCREEN] = &editorForm{}
}

func (e *editorForm) KeyDown(key Key) bool {
	switch key {
	case KEY_ESC:
		// do nothing
	case KEY_UP:
		e.cy -= 1
	case KEY_DOWN:
		e.cy += 1
	case KEY_LEFT:
		e.cx -= 1
	case KEY_RIGHT:
		e.cx += 1
	case 'z':
		e.cursor -= 1
	case 'x':
		e.cursor += 1
	case KEY_SPACE:
		fallthrough
	case KEY_RETURN:
		e.Map[e.cy*CZ+e.cx] = Brushes[e.cursor]
	default:
		return false
	}

	e.cx = (CZ + e.cx) % CZ
	e.cy = (CZ + e.cy) % CZ
	e.cursor = (len(Brushes) + e.cursor) % (len(Brushes))
	return true
}

func (e *editorForm) Draw() {
	Background(Dark)
	Fill(Bright)
	e.drawMap(Sz(4)+Sz(1)/2, Sz(1)/2)
	e.drawPanel(Sz(13), Sz(1))
}

func (e *editorForm) drawMap(x, y int) {
	PushMatrix()
	Translate(x, y)
	NoFill()
	Rect(-1, -1, Sz(ROW)+1, Sz(ROW)+1) //border
	for i, v := range e.Map {          //tiles
		x := i % CZ
		y := i / CZ

		img, in := ImageTable[byte(v)]
		if v >= 0 && in {
			img.DrawRect(Sz(x), Sz(y), Sz(1), Sz(1))
		}
		if x == e.cx && y == e.cy {
			ImageTable[255].DrawRect(Sz(x), Sz(y), Sz(1), Sz(1))
		}
	}
	PopMatrix()
}

func (e *editorForm) drawPanel(x, y int) {
	PushMatrix()
	Translate(x, y)

	Text("SELECT TILES BY Z/X", 0, 0, Sz(10), Sz(1))
	Translate(0, Sz(1))
	Text("PLACE BY SPACE/ENTER", 0, 0, Sz(10), Sz(1))

	Translate(Sz(1), Sz(2))
	for i, v := range Brushes {
		ImageTable[byte(v)].DrawRect(Sz(i%ROW), Sz(i/ROW), Sz(1), Sz(1))
		if i == e.cursor {
			ImageTable[255].DrawRect(Sz(i%ROW), Sz(i/ROW), Sz(1), Sz(1))
		}
	}

	PopMatrix()
}

func (e *editorForm) Setup() {
	Brushes = make([]int, 0, 0)
	for _, v := range Blocks {
		Brushes = append(Brushes, int(v))
	}

	for v := range game.Entities {
		Brushes = append(Brushes, int(v))
	}

	e.Map = make([]int, CZ*CZ, CZ*CZ)
	for i := range e.Map {
		e.Map[i] = -1
	}
}

func (e *editorForm) Start() {
	Title("MSPAINT.EXE")
}

func (e *editorForm) Stop() {
}
