package ui

import (
	. "client/grocessing"
	"client/request"
	"fmt"
	"os"
	"path"

	"server/game"
)

var (
	SERVER_URL = "http://127.0.0.1:4242/"

	Dark   Color = Hc(0x204631)
	Bright Color = Hc(0xD7E894)
	Border Color = Hc(0x527F39)
	scale  int   = 32

	Prog  = "./examples/meow.per"
	token = ""

	screen Formid
	Forms  map[Formid]Form = make(map[Formid]Form)

	Server *request.Server

	ImageTable map[byte]*Image

	Blocks []byte
)

const (
	LOGIN_SCREEN = Formid(iota)
	GAME_SCREEN
	HIGHSCORES_SCREEN
	EXPLORER_SCREEN
	MAINMENU_SCREEN
	EDITOR_SCREEN
	EXIT

	DIM_X = 25
	DIM_Y = 9

	FONT_NAME = "PIXEL.TTF"
)

const (
	OK = iota
	GAME_OVER
	CONNECTING
	NICK_LEN = 8
)

type sketch struct {
}

type Form interface {
	Draw()
	KeyDown(Key) bool
	Setup()
	Start()
	Stop()
}

type Formid uint

func Screen(new Formid) {
	if new == EXIT {
		os.Exit(0)
	}

	Forms[screen].Stop()
	screen = new
	Forms[new].Start()
}

func Sz(a int) int {
	return a * scale
}

func Start() {

	GrocessingStart(sketch{})

}

func ConfigFile() string {
	return path.Join(os.Getenv("HOME"), ".config", "perechnicy.creds")
}

func (s sketch) KeyPressed() {
	if !Forms[screen].KeyDown(KeyCode) {
		if KeyCode == KEY_ESC {
			Screen(MAINMENU_SCREEN)
		}
	}
}

func (s sketch) Setup() {
	scale = 48
	Size(Sz(25), Sz(9))
	TextAlign(ALIGN_CENTER)

	font, err := CreateFont(path.Join("resources", FONT_NAME), Sz(1)/2)
	if err != nil {
		panic(err)
	}

	ImageTable = make(map[byte]*Image)

	Blocks = make([]byte, 0, 0)

	// block; it's a map
	for id := range game.Blocks {
		AddTile(id)
		Blocks = append(Blocks, id)
	}

	// entity; it's a map too
	for id := range game.Entities {
		AddTile(id)
	}

	// local stuff; 255 is our cursor
	for _, id := range []byte{255} {
		AddTile(id)
	}

	for _, v := range Forms {
		v.Setup()
	}

	SetFont(font)
	Fill(Dark)
	Stroke(Border)

}

func (s sketch) Draw() {
	Background(Dark)
	Forms[screen].Draw()
}

func DrawTable(table [][]string, colw int) {

	PushMatrix()

	for i := range table {
		PushMatrix()
		for _, txt := range table[i] {
			Fill(Dark)
			Rect(-1, -1, Sz(colw)/2, Sz(1)/2+1)
			Fill(Bright)
			Text(txt, 0, 0, Sz(colw)/2, Sz(1)/2)
			Translate(0, Sz(1)/2)
		}
		PopMatrix()
		Translate(Sz(colw)/2-1, 0)

	}

	PopMatrix()
}

func AddTile(id byte) {
	file := path.Join("resources", "tile", fmt.Sprintf("%v.png", id))
	img, err := LoadImage(file)
	if err != nil {
		panic(err)
	}
	ImageTable[id] = img
}
