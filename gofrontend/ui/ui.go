package ui

import (
	. "../grocessing"
	"../req"
	"fmt"
	"os"
	"path/filepath"
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
)

const (
	LOGIN_SCREEN = Formid(iota)
	GAME_SCREEN
	HIGHSCORES_SCREEN
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
	Start()
	Stop()
}

type Formid uint

func Screen(new Formid) {
	Forms[screen].Stop()
	screen = new
	Forms[new].Start()
}

func Sz(a int) int {
	return a * scale
}

func Start() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Please specify prog file\n")
		return
	}

	Prog = os.Args[1]

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	os.Chdir(dir)

	GrocessingStart(sketch{})
}

func ConfigFile() string {
	return fmt.Sprintf("%s/.config/perechnicy.creds", os.Getenv("HOME"))
}

func (s sketch) KeyPressed() {
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