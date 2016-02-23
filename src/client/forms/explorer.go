package forms

import (
	. "client/grocessing"
	"client/request"
	. "client/ui"
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	TILES_X = DIM_X * SCALE
	TILES_Y = DIM_Y * SCALE
	SCALE   = 1
)

type explorerForm struct {
	Map    []int
	cx, cy int
	lock   sync.Mutex
	stahp  chan bool
}

func init() {
	Forms[EXPLORER_SCREEN] = &explorerForm{}
	Server = request.NewServer(SERVER_URL, "", "")
}

func (e *explorerForm) update() {
	e.lock.Lock()
	defer e.lock.Unlock()
	mp, err := Server.GetMap(e.cx, e.cy, DIM_X, DIM_Y)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	e.Map = mp
}

func (e *explorerForm) KeyDown(key Key) bool {
	switch key {
	case KEY_UP:
		e.cy -= 1
		e.update()
	case KEY_DOWN:
		e.cy += 1
		e.update()
	case KEY_LEFT:
		e.cx -= 1
		e.update()
	case KEY_RIGHT:
		e.cx += 1
		e.update()
	default:
		return false
	}
	return true
}

func (e *explorerForm) Draw() {
	if e.Map != nil {
		e.lock.Lock()
		defer e.lock.Unlock()
		for i, v := range e.Map { //tiles
			ImageTable[byte(v)].DrawRect(sz(i%DIM_X), sz(i/DIM_X), sz(1), sz(1))
		}
	} else {
		Fill(Bright)
		Text("Connection problem", 0, 0, Sz(DIM_X), Sz(DIM_Y))
	}
}

func sz(a int) int {
	return Sz(a) / SCALE
}

func (e *explorerForm) Setup() {

}
func (e *explorerForm) Start() {
	Title("EXPLORER.EXE")

	e.stahp = make(chan bool)

	go func() {
		for {
			select {
			case <-time.After(time.Second / 10):
				go e.update()
			case <-e.stahp:
				return
			}
		}
	}()
}
func (e *explorerForm) Stop() {
	e.stahp <- true
}
