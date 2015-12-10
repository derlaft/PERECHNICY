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
)

const (
	SERVER_URL = "http://127.0.0.1:4242/"
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

	go doStart()
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

	lock.Lock()
	drawState(sz(1)/2, sz(1)/2)
	drawRegisters(sz(5)/7, sz(8)-sz(1)/2)
	drawMap(sz(4), sz(2))
	lock.Unlock()
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
		{"AP", fmt.Sprintf("%d", 2)},
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
	Rect(-1, -1, sz(5)+1, sz(5)+1) //border
	for i, v := range state.Map[id] {
		imageTable[v].DrawRect(sz(i%5), sz(i/5), sz(1), sz(1))
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
	body, err := data("start", url.Values{"User": {"meow"}, "Token": {"coco"}})

	if err != nil || string(body) != "OK" {
		fmt.Println("Could not register (error is %v, response is %v)", err, string(body))
	}

	go func() {
		for {
			select {
			case <-time.After(time.Second / 10):
				doGetData()
			}
		}
	}()
}

func doGetData() {
	body, err := data("get", url.Values{"User": {"meow"}, "Token": {"coco"}})

	if err != nil {
		panic(fmt.Sprintf("Could not fetch data: %v", err))
	}

	newState := entities.JSONOutput{}
	err = json.Unmarshal(body, &newState)

	if err != nil {
		panic(fmt.Sprintf("Could not decode request: %v", err))
	}

	lock.Lock()
	state = newState
	lock.Unlock()

	//fmt.Printf("Got data %+v\n", state)

}
