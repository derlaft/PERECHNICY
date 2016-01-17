package gameform

import (
	"../../block"
	"../../entity"
	"../../entity/entities"
	. "../grocessing"
	. "../ui"
	"fmt"
	"os"
	"sync"
	"time"
)

type GameForm struct {
}

var (
	imageTable map[byte]*Image
	state      *entities.JSONOutput
	lock       sync.Mutex
	stahp      chan bool

	connectStatus string
)

const (
	GAME_OVER_SIGN = "GAME OVER"
)

func init() {
	stahp = make(chan bool, 1)

	Forms[GAME_SCREEN] = GameForm{}
}

func (g GameForm) KeyDown(a Key) {
	switch a {
	case KEY_RETURN:
		if state.Destroyed {
			fmt.Println("JUMPING")
			Screen(LOGIN_SCREEN)
		}
	}
}

func (g GameForm) Start() {
	if imageTable == nil {
		imageTable = make(map[byte]*Image)
		for id := range block.Blocks {
			addTile(id)
		}
		for _, id := range entity.Entities {
			addTile(id)
		}

	}

	state = &entities.JSONOutput{}
	stahp = make(chan bool, 1)

	for {
		select {
		case <-time.After(time.Second / 10):
			newState, err := Server.GetData()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
			lock.Lock()
			state = newState
			lock.Unlock()

			if state.Destroyed {
				return
			}
		case <-stahp:
			return
		}

	}
}

func (g GameForm) Stop() {
	stahp <- true
}

func (g GameForm) Draw() {
	lock.Lock()
	drawState(Sz(1)/2, Sz(1)/2)
	drawRegisters(Sz(5)/7, Sz(8)-Sz(1)/2)
	drawMap(Sz(4), Sz(2))
	lock.Unlock()
}

func drawState(x, y int) {

	PushMatrix()
	Translate(x, y)
	Fill(Dark)

	table := [][]string{
		{"INST#", fmt.Sprintf("%d", state.IP)},
		{"INST", fmt.Sprintf("%3s", state.Inst)}, //text should not be empty
		{"HEALTH", fmt.Sprintf("%d", state.Health)},
		{"ENERGY", fmt.Sprintf("%d", state.Energy)},
		{"AP", fmt.Sprintf("%d", state.AP)},
		{"_GELIFOZ", fmt.Sprintf("%08b", state.Reg[15])},
	}

	DrawTable(table, 8)
	PopMatrix()

}

func drawRegisters(x, y int) {

	PushMatrix()
	Translate(x, y)
	Fill(Dark)

	table := make([][]string, 16)
	for i := range table {
		table[i] = []string{
			fmt.Sprintf("R%1X", i), fmt.Sprintf("&%02X", state.Reg[i]),
		}
	}

	DrawTable(table, 3)
	PopMatrix()
}

func drawMap(x, y int) {

	PushMatrix()
	Fill(Dark)
	Translate(x, y)

	for i := 0; i < 3; i++ {
		drawMapMain(i)
		Translate(Sz(6), 0)
	}

	PopMatrix()
}

func addTile(id byte) {
	file := fmt.Sprintf("./tile/%v.png", id)
	img, err := LoadImage(file)
	if err != nil {
		panic(err)
	}
	imageTable[id] = img
}

func drawMapMain(id int) {
	Rect(-1, -1, Sz(5)+1, Sz(5)+1)    //border
	for i, v := range state.Map[id] { //tiles
		imageTable[v].DrawRect(Sz(i%5), Sz(i/5), Sz(1), Sz(1))
	}
	if state.Destroyed { //gameover sign
		Fill(Dark)
		//TextStyle(STYLE_BOLD)
		Translate(-2, -2)
		Text(GAME_OVER_SIGN, 0, 0, Sz(5), Sz(5))
		Translate(4, 4)
		Text(GAME_OVER_SIGN, 0, 0, Sz(5), Sz(5))
		Translate(-2, -2)
		Fill(Bright)
		TextStyle(STYLE_NORMAL)
		Text(GAME_OVER_SIGN, 0, 0, Sz(5), Sz(5))
	}

}
