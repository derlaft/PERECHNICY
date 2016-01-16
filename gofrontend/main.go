package main

import (
	"../block"
	"../entity"
	"../entity/entities"
	. "./grocessing"
	"./req"
	"fmt"
	"gopkg.in/gcfg.v1"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode"
)

var (
	dark       Color = Hc(0x204631)
	bright     Color = Hc(0xD7E894)
	border     Color = Hc(0x527F39)
	state      *entities.JSONOutput
	lock       sync.Mutex
	imageTable map[byte]*Image
	scale      int = 32

	prog  = "./examples/meow.per"
	token = ""

	screen = LOGIN_SCREEN

	SYMBOLS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	cursor  = 0
	nick    = [NICK_LEN]rune{
		'_', '_', '_', '_',
		'_', '_', '_', '_',
	}
	connectStatus string = PRESS_ENTER
)

const (
	SERVER_URL     = "http://127.0.0.1:4242/"
	GAME_OVER_SIGN = "GAME OVER"
	PRESS_ENTER    = "Press RETURN"
	LINE_WIDTH     = 20

	LOGIN_SCREEN = iota
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

func sz(a int) int {
	return a * scale
}

func main() {

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	os.Chdir(dir)

	GrocessingStart(sketch{})
}

func next(r rune, dx int) rune {
	i := strings.Index(SYMBOLS, string(r))
	ret := SYMBOLS[(len(SYMBOLS)+i+dx)%len(SYMBOLS)]
	return rune(ret)
}

func (s sketch) KeyPressed() {
	switch screen {
	case LOGIN_SCREEN:
		switch KeyCode {
		case KEY_UP:
			nick[cursor] = next(nick[cursor], -1)
		case KEY_DOWN:
			nick[cursor] = next(nick[cursor], +1)
		case KEY_LEFT:
			cursor = (cursor + NICK_LEN - 1) % NICK_LEN
		case KEY_RIGHT:
			cursor = (cursor + 1) % NICK_LEN
		case KEY_RETURN:
			connectStatus = "Connecting to the server..."
			go doRegister()
		default:
			if KeyCode == ' ' {
				KeyCode = '_'
			}

			if KeyCode >= 'a' && KeyCode <= 'z' ||
				KeyCode >= '0' && KeyCode <= '9' ||
				KeyCode == '_' {

				nick[cursor] = unicode.ToUpper(rune(KeyCode))
				cursor = (cursor + 1) % NICK_LEN
			}
		}

	case GAME_SCREEN:
		switch KeyCode {
		case KEY_RETURN:
			if state.Destroyed {
				connectStatus = PRESS_ENTER
				screen = LOGIN_SCREEN
			}
		}
	}
}

func (s sketch) Setup() {
	scale = 48
	Size(sz(25), sz(9))
	TextAlign(ALIGN_CENTER)

	font, err := CreateFont("pixel.ttf", sz(1)/2)
	if err != nil {
		panic(err)
	}

	//fill in img table
	imageTable = make(map[byte]*Image)
	for id := range block.Blocks {
		addTile(id)
	}
	for _, id := range entity.Entities {
		addTile(id)
	}

	SetFont(font)
	Fill(dark)
	Stroke(border)

	//go doRegister()

}

func addTile(id byte) {
	file := fmt.Sprintf("tile/%v.png", id)
	img, err := LoadImage(file)
	if err != nil {
		panic(err)
	}
	imageTable[id] = img
}

func (s sketch) Draw() {

	Background(dark)

	switch screen {
	case GAME_SCREEN:

		//lock.Lock()
		drawState(sz(1)/2, sz(1)/2)
		drawRegisters(sz(5)/7, sz(8)-sz(1)/2)
		drawMap(sz(4), sz(2))
		//lock.Unlock()

	case LOGIN_SCREEN:
		drawInput(sz(0), sz(1))
		//drawLoginStatus(sz(4), sz(4))

	}
}

func drawTable(table [][]string, colw int) {

	PushMatrix()

	for i := range table {
		PushMatrix()
		for _, txt := range table[i] {
			Fill(dark)
			Rect(-1, -1, sz(colw)/2, sz(1)/2+1)
			Fill(bright)
			Text(txt, 0, 0, sz(colw)/2, sz(1)/2)
			Translate(0, sz(1)/2)
		}
		PopMatrix()
		Translate(sz(colw)/2-1, 0)

	}

	PopMatrix()
}

func drawState(x, y int) {
	//@TODO: wat
	if state == nil {
		return
	}

	PushMatrix()
	Translate(x, y)
	Fill(dark)

	table := [][]string{
		{"INST#", fmt.Sprintf("%d", state.IP)},
		{"INST", fmt.Sprintf("%3s", state.Inst)}, //text should not be empty
		{"HEALTH", fmt.Sprintf("%d", state.Health)},
		{"ENERGY", fmt.Sprintf("%d", state.Energy)},
		{"AP", fmt.Sprintf("%d", state.AP)},
		{"_GELIFOZ", fmt.Sprintf("%08b", state.Reg[15])},
	}

	drawTable(table, 8)
	PopMatrix()

}

func getnick() (n string) {
	for _, v := range nick {
		n += string(v)
	}
	return
}

func getcursor() (n string) {
	for i := 0; i < NICK_LEN; i++ {
		if i == cursor {
			n += "^"
		} else {
			n += " "
		}
	}
	return
}

func drawInput(x, y int) {

	PushMatrix()
	Translate(x, y)
	Fill(bright)

	Text("Input your name", 0, 0, sz(25), sz(1))
	Text("Use arrows to input", 0, sz(1), sz(25), sz(1))

	Translate(0, sz(2))

	Text(getnick(), 0, 0, sz(25), sz(1))
	Text(getcursor(), 0, sz(1)/2, sz(25), sz(1))

	Translate(0, sz(1))

	for i := 0; i <= len(connectStatus)/LINE_WIDTH; i++ {
		Text(connectStatus[i*LINE_WIDTH:Min(len(connectStatus), (i+1)*LINE_WIDTH)], 0, 0, sz(25), sz(1))
		Translate(0, sz(1)/2)
	}

	PopMatrix()
}

func drawRegisters(x, y int) {

	PushMatrix()
	Translate(x, y)
	Fill(dark)

	table := make([][]string, 16)
	for i := range table {
		table[i] = []string{
			fmt.Sprintf("R%1X", i), fmt.Sprintf("&%02X", state.Reg[i]),
		}
	}

	drawTable(table, 3)
	PopMatrix()
}

func drawMap(x, y int) {

	PushMatrix()
	Fill(dark)
	Translate(x, y)

	for i := 0; i < 3; i++ {
		drawMapMain(i)
		Translate(sz(6), 0)
	}

	PopMatrix()
}

func drawMapMain(id int) {
	Rect(-1, -1, sz(5)+1, sz(5)+1)    //border
	for i, v := range state.Map[id] { //tiles
		imageTable[v].DrawRect(sz(i%5), sz(i/5), sz(1), sz(1))
	}
	if state.Destroyed { //gameover sign
		Fill(dark)
		//TextStyle(STYLE_BOLD)
		Translate(-2, -2)
		Text(GAME_OVER_SIGN, 0, 0, sz(5), sz(5))
		Translate(4, 4)
		Text(GAME_OVER_SIGN, 0, 0, sz(5), sz(5))
		Translate(-2, -2)
		Fill(bright)
		TextStyle(STYLE_NORMAL)
		Text(GAME_OVER_SIGN, 0, 0, sz(5), sz(5))
	}

}

func data(method string, params url.Values) ([]byte, error) {
	resp, err := http.PostForm(SERVER_URL+method, params)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func doStart() {

	f, err := os.Open(prog)
	defer f.Close()
	if err != nil {
		connectStatus = err.Error()
		return
	}

	progData, err := ioutil.ReadAll(f)
	if err != nil {
		connectStatus = err.Error()
		return
	}

	err = server.Start(string(progData))
	if err != nil {
		connectStatus = err.Error()
		return
	}

	screen = GAME_SCREEN
	state = &entities.JSONOutput{}

	go func() {
		for {
			select {
			case <-time.After(time.Second / 10):
				newState, err := server.GetData()
				if err != nil {
					fmt.Fprintf(os.Stderr, "%v\n", err)
				}
				lock.Lock()
				state = newState
				lock.Unlock()

				if state.Destroyed {
					return
				}
			}
		}
	}()
}

var (
	server *req.Server
)

func doRegister() {

	server = req.NewServer(SERVER_URL, getnick(), token)
	if server.Token != "" {
		doStart()
		return
	}

	err := server.Register()
	if err != nil {
		connectStatus = err.Error()
		return
	}

	doStart()
}
