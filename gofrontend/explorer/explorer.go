package explorer

import (
	. "../grocessing"
	"../req"
	. "../ui"
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

var (
	Map    []int = nil
	cx           = 0
	cy           = 0
	server *req.Server
	lock   sync.Mutex
	stahp  chan bool
)

type explorerForm struct {
}

func init() {
	Forms[EXPLORER_SCREEN] = explorerForm{}
	server = req.NewServer(SERVER_URL, "", "")
}

func update() {
	lock.Lock()
	defer lock.Unlock()
	mp, err := server.GetMap(cx, cy, DIM_X, DIM_Y)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	Map = mp
}

func (e explorerForm) KeyDown(key Key) {
	switch key {
	case KEY_UP:
		cy -= 1
		update()
	case KEY_DOWN:
		cy += 1
		update()
	case KEY_LEFT:
		cx -= 1
		update()
	case KEY_RIGHT:
		cx += 1
		update()
	}

}

func (e explorerForm) Draw() {
	if Map != nil {
		lock.Lock()
		defer lock.Unlock()
		for i, v := range Map { //tiles
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

func (e explorerForm) Setup() {

}
func (e explorerForm) Start() {
	Title("EXPLORER.EXE")

	stahp = make(chan bool)

	go func() {
		for {
			select {
			case <-time.After(time.Second / 10):
				go update()
			case <-stahp:
				return
			}
		}
	}()
}
func (e explorerForm) Stop() {
	stahp <- true
}
