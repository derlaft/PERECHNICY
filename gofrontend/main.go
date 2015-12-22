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
	scale      int  = 32
	gameOver   bool = false

	prog     = "./examples/meow.per"
	username = "meow"
	token    = ""
)

const (
	SERVER_URL     = "http://127.0.0.1:4242/"
	GAME_OVER_SIGN = "GAME OVER"
)

const (
	OK = iota
	GAME_OVER
	NETWORK_ERROR
)

type sketch struct {
}

func sz(a int) int {
	return a * scale
}

func main() {
	GrocessingStart(sketch{})
}

func (s sketch) Setup() {
	scale = 48
	Size(sz(25), sz(9))

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

	go doRegister()
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

	//lock.Lock()
	drawState(sz(1)/2, sz(1)/2)
	drawRegisters(sz(5)/7, sz(8)-sz(1)/2)
	drawMap(sz(4), sz(2))
	//lock.Unlock()
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
	if gameOver { //gameover sign
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
	if err != nil {
		panic(err)
	}
	defer f.Close()

	progData, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	body, err := data("start", url.Values{
		"User":  {username},
		"Token": {token},
		"Prog":  {string(progData)},
	})

	if err != nil || string(body) != "OK" {
		fmt.Printf(`Could not upload (error is "%v", response is "%v")\n`, err, string(body))
		return
	}

	go func() {
		for {
			select {
			case <-time.After(time.Second / 10):
				switch doGetData() {
				case GAME_OVER:
					gameOver = true
					break
				}
			}
		}
	}()
}

type tokenResult struct {
	Token string
}

func doRegister() {
	body, err := data("register", url.Values{"User": {username}})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while fetching: %v", err)
		return
	}

	if string(body) == "FAILED" {
		fmt.Fprintf(os.Stderr, "Failed to register\n")
		return
	}

	tr := tokenResult{}
	err = json.Unmarshal(body, &tr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while decoding: %v", err)
		return
	}

	token = tr.Token
	fmt.Println("token is", token)
	doStart()
}

func doGetData() int {
	body, err := data("get", url.Values{"User": {username}, "Token": {token}})

	if err != nil {
		return NETWORK_ERROR
	}

	fmt.Println(string(body))

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
