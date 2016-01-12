package main

import (
	"../block"
	"../entity"
	"../entity/entities"
	. "./grocessing"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

var (
	dark       Color = Hc(0x204631)
	bright     Color = Hc(0xD7E894)
	border     Color = Hc(0x527F39)
	state      entities.JSONOutput
	lock       sync.Mutex
	imageTable map[byte]*Image
	scale      int = 32

	prog  = "./examples/meow.per"
	token = ""

	screen = LOGIN_SCREEN

	SYMBOLS = []rune{
		'_',
		'A', 'D', 'E',
		'F', 'G', 'I',
		'K', 'L', 'O',
		'P', 'R', 'S',
		'T', 'U', 'X',
		'2', '4', '7',
	}
	cursor            = 0
	nick              = [NICK_LEN]int{0}
	connectStatus int = NOT_YET
)

const (
	SERVER_URL     = "http://127.0.0.1:4242/"
	GAME_OVER_SIGN = "GAME OVER"

	LOGIN_SCREEN = iota
	GAME_SCREEN
	HIGHSCORES_SCREEN
)

const (
	OK = iota
	GAME_OVER
	CONNECTING
	NETWORK_ERROR
	NOT_YET
	BUSY
	PROG_UPLOAD_FAILED
	NICK_LEN = 8
)

type sketch struct {
}

func sz(a int) int {
	return a * scale
}

func main() {
	GrocessingStart(sketch{})
}

func (s sketch) KeyPressed() {
	switch screen {
	case LOGIN_SCREEN:
		switch KeyCode {
		case KEY_UP:
			nick[cursor] = (nick[cursor] + 1) % len(SYMBOLS)
		case KEY_DOWN:
			nick[cursor] = (nick[cursor] + len(SYMBOLS) - 1) % len(SYMBOLS)
		case KEY_LEFT:
			cursor = (cursor + NICK_LEN - 1) % NICK_LEN
		case KEY_RIGHT:
			cursor = (cursor + 1) % NICK_LEN
		case KEY_RETURN:
			connectStatus = CONNECTING
			go doRegister()
		}

	case GAME_SCREEN:
		switch KeyCode {
		case KEY_RETURN:
			if connectStatus == GAME_OVER {
				connectStatus = NOT_YET
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
		drawInput(sz(0), sz(3))
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
		n += string(SYMBOLS[v])
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
	var t string
	switch connectStatus {
	case NOT_YET:
		t = "Press ENTER"
	case BUSY:
		t = "Nickname BUSY"
	case NETWORK_ERROR:
		t = "Network ERR"
	case PROG_UPLOAD_FAILED:
		t = "Prog problemo"
	case OK:
		t = "Starting game..."
	default:
		t = "WAT"
	}

	Text(t, 0, sz(0), sz(25), sz(1))

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
	if connectStatus == GAME_OVER { //gameover sign
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
		fmt.Fprintf(os.Stderr, `Could not open prog (%v")\n`, err)
		connectStatus = PROG_UPLOAD_FAILED
		return
	}

	progData, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, `Could not read prog (%v")\n`, err)
		connectStatus = PROG_UPLOAD_FAILED
		return
	}

	body, err := data("start", url.Values{
		"User":  {getnick()},
		"Token": {token},
		"Prog":  {string(progData)},
	})

	if err != nil || string(body) != `{"result": "OK"}` {
		fmt.Fprintf(os.Stderr, `Could not upload (error is "%v", response is "%v")\n`, err, string(body))
		connectStatus = PROG_UPLOAD_FAILED
		return
	}

	connectStatus = OK
	screen = GAME_SCREEN

	go func() {
		for {
			select {
			case <-time.After(time.Second / 10):
				switch doGetData() {
				case GAME_OVER:
					connectStatus = GAME_OVER
					return
				}
			}
		}
	}()
}

type tokenResult struct {
	Token string
}

func doRegister() {
	if token != "" {
		doStart()
		return
	}

	body, err := data("register", url.Values{"User": {getnick()}})

	if err != nil {
		connectStatus = NETWORK_ERROR
		return
	}

	if string(body) == "FAILED" {
		connectStatus = BUSY
		return
	}

	tr := tokenResult{}
	err = json.Unmarshal(body, &tr)
	if err != nil {
		connectStatus = NETWORK_ERROR
		return
	}

	token = tr.Token
	doStart()
}

func doGetData() int {
	body, err := data("get", url.Values{"User": {getnick()}, "Token": {token}})

	if err != nil {
		return NETWORK_ERROR
	}

	if string(body) == "DESTROYED" {
		return GAME_OVER
	}

	newState := entities.JSONOutput{}
	err = json.Unmarshal(body, &newState)

	if err != nil {
		return NETWORK_ERROR
	}

	lock.Lock()
	state = newState
	lock.Unlock()

	//fmt.Printf("Got data %+v\n", state)

	return OK
}
