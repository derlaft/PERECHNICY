package main

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
	EXPLORER_FORM = iota
	DIM_X         = 25
	DIM_Y         = 9
)

var (
	Map    []int = nil
	cx           = 0
	cy           = 0
	server *req.Server
	lock   sync.Mutex
)

type explorerForm struct {
}

func main() {

	server = &req.Server{URL: "http://127.0.0.1:4242/"}

	Forms[EXPLORER_FORM] = explorerForm{}
	Screen(EXPLORER_FORM)
	Start()

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
			ImageTable[byte(v)].DrawRect(Sz(i%DIM_X), Sz(i/DIM_X), Sz(1), Sz(1))
		}
	}
}

func (e explorerForm) Setup() {
	go func() {
		for {
			select {
			case <-time.After(time.Second / 10):
				go update()
			}
		}
	}()
}
func (e explorerForm) Start() {
}
func (e explorerForm) Stop() {
}
