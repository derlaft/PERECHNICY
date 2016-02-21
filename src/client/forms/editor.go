package forms

import (
	. "client/grocessing"
	. "client/ui"
	"server/world"
)

const (
	CZ  = world.CHUNK_SIZE
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

func (e *editorForm) KeyDown(key Key) {
	switch key {
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
		e.Map[cy*CZ+cx] = Brushes[cursor]
	}

	e.cx = (CZ + e.cx) % CZ
	e.cy = (CZ + e.cy) % CZ
	e.cursor = (len(Blocks) + 1 + e.cursor) % (len(Blocks) + 1)

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
		if v > 0 && in {
			img.DrawRect(Sz(x), Sz(y), Sz(1), Sz(1))
		}
		if x == cx && y == cy {
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
	for i, v := range Blocks {
		ImageTable[v].DrawRect(Sz(i%ROW), Sz(i/ROW), Sz(1), Sz(1))
		if i == cursor {
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
	Brushes = append(Brushes, -1)
}

func (e *editorForm) Start() {
	Title("MSPAINT.EXE")
}

func (e *editorForm) Stop() {
}
