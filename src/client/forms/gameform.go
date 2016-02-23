package forms

import (
	. "client/grocessing"
	. "client/ui"
	"fmt"
	"os"
	"server/game/entity"
	"sync"
	"time"
)

type GameForm struct {
	state         *entity.JSONOutput
	lock          sync.Mutex
	stahp         chan bool
	connectStatus string
}

const (
	GAME_OVER_SIGN = "GAME OVER"
)

func init() {

	Forms[GAME_SCREEN] = &GameForm{}
}

func (g *GameForm) Setup() {
	g.stahp = make(chan bool, 1)
}

func (g *GameForm) KeyDown(a Key) bool {
	switch a {
	case KEY_RETURN:
		if g.state.Destroyed {
			fmt.Println("JUMPING")
			Screen(LOGIN_SCREEN)
		}
	default:
		return false
	}

	return true
}

func (g *GameForm) Start() {

	g.state = &entity.JSONOutput{}
	g.stahp = make(chan bool, 1)

	for {
		select {
		case <-time.After(time.Second / 10):
			newState, err := Server.GetData()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
			}
			g.lock.Lock()
			g.state = newState
			g.lock.Unlock()

			if g.state.Destroyed {
				return
			}
		case <-g.stahp:
			return
		}

	}
}

func (g *GameForm) Stop() {
	g.stahp <- true
}

func (g *GameForm) Draw() {
	g.lock.Lock()
	g.drawState(Sz(1)/2, Sz(1)/2)
	g.drawRegisters(Sz(5)/7, Sz(8)-Sz(1)/2)
	g.drawMap(Sz(4), Sz(2))
	g.lock.Unlock()
}

func (g *GameForm) drawState(x, y int) {

	PushMatrix()
	Translate(x, y)
	Fill(Dark)

	table := [][]string{
		{"INST#", fmt.Sprintf("%d", g.state.IP)},
		{"INST", fmt.Sprintf("%3s", g.state.Inst)}, //text should not be empty
		{"HEALTH", fmt.Sprintf("%d", g.state.Health)},
		{"ENERGY", fmt.Sprintf("%d", g.state.Energy)},
		{"AP", fmt.Sprintf("%d", g.state.AP)},
		{"_GELIFOZ", fmt.Sprintf("%08b", g.state.Reg[15])},
	}

	DrawTable(table, 8)
	PopMatrix()

}

func (g *GameForm) drawRegisters(x, y int) {

	PushMatrix()
	Translate(x, y)
	Fill(Dark)

	table := make([][]string, 16)
	for i := range table {
		table[i] = []string{
			fmt.Sprintf("R%1X", i), fmt.Sprintf("&%02X", g.state.Reg[i]),
		}
	}

	DrawTable(table, 3)
	PopMatrix()
}

func (g *GameForm) drawMap(x, y int) {

	PushMatrix()
	Fill(Dark)
	Translate(x, y)

	for i := 0; i < 3; i++ {
		g.drawMapMain(i)
		Translate(Sz(6), 0)
	}

	PopMatrix()
}

func (g *GameForm) drawMapMain(id int) {
	Rect(-1, -1, Sz(5)+1, Sz(5)+1)      //border
	for i, v := range g.state.Map[id] { //tiles
		ImageTable[v].DrawRect(Sz(i%5), Sz(i/5), Sz(1), Sz(1))
	}
	if g.state.Destroyed { //gameover sign
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
