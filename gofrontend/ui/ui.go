package ui

import (
	. "../grocessing"
	"../req"
	"fmt"
	"os"

	"../../block"
	"../../entity"
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

	Server *req.Server

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
	KeyDown(Key)
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
	return fmt.Sprintf("%s/.config/perechnicy.creds", os.Getenv("HOME"))
}

func (s sketch) KeyPressed() {
	if KeyCode == KEY_ESC {
		Screen(MAINMENU_SCREEN)
	}
	Forms[screen].KeyDown(KeyCode)
}

func (s sketch) Setup() {
	scale = 48
	Size(Sz(25), Sz(9))
	TextAlign(ALIGN_CENTER)

	font, err := CreateFont("pixel.ttf", Sz(1)/2)
	if err != nil {
		panic(err)
	}

	ImageTable = make(map[byte]*Image)

	Blocks = make([]byte, 0, 0)

	// block
	for id := range block.Blocks {
		AddTile(id)
		Blocks = append(Blocks, id)
	}

	// entity
	for _, id := range entity.Entities {
		AddTile(id)
	}

	// local stuff
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
	file := fmt.Sprintf("./tile/%v.png", id)
	img, err := LoadImage(file)
	if err != nil {
		panic(err)
	}
	ImageTable[id] = img
}
