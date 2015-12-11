package server

import (
	"../block"
	. "../chunk"
	. "../entity/entities"
	"../figex/mio"
	. "../game"
	. "../stuff"
	"fmt"
	"io/ioutil"
	"net/http"
	//_ "net/http/pprof"
	"os"
	"path"
	"sync"
	"time"
)

var (
	game *Game
	ctl  EntityControl
)

type Bots map[User]*Control

type EntityControl struct {
	bots Bots
	sync.RWMutex
}

type User struct {
	Name string
}

func (u *User) progPath() string {
	return "./data/" + path.Base(u.Name) + ".per"
}

func Serve(addr string) {

	game = NewGame()
	ctl = EntityControl{bots: make(Bots)}

	game.World.GetChunk(Point{0, 0}).Data[2][2] = 9
	go gameLoop()

	http.Handle("/", http.FileServer(http.Dir("jsfrontend")))
	http.HandleFunc("/start", start)
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/get", get)
	http.ListenAndServe(addr, nil)
}

func gameLoop() {
	for {
		select {
		case <-time.After(time.Second / 10):

			game.Tick()
		}
	}
}

func checkAuth(w http.ResponseWriter, r *http.Request) (*User, bool) {

	user, token := r.FormValue("User"), r.FormValue("Token")

	if user == "" || token == "" {
		//TODO: implement auth
		fmt.Fprintf(w, "Bad auth")
		return nil, false
	}

	return &User{Name: user}, true
}

func start(w http.ResponseWriter, r *http.Request) {

	user, ok := checkAuth(w, r)
	if !ok {
		return
	}

	// check if already started
	bot, exists := ctl.get(*user)
	if exists && !bot.Destroyed {
		fmt.Fprintf(w, "EXISTS")
		return
	}

	// start
	bot = getBot(user)
	ctl.Lock()
	ctl.bots[*user] = bot
	ctl.Unlock()

	NewEntity(game, Point{2, 2}, &Bear{})
	NewEntity(game, Point{2, 3}, &Bear{})
	NewEntity(game, Point{2, 4}, &Bear{})
	NewEntity(game, Point{2, 5}, &Bear{})
	NewEntity(game, Point{2, 6}, &Bear{})

	fmt.Fprintf(w, "OK")
}

func (c *EntityControl) get(u User) (*Control, bool) {
	ctl.RLock()
	bot, exists := ctl.bots[u]
	ctl.RUnlock()
	return bot, exists
}

func upload(w http.ResponseWriter, r *http.Request) {
	user, ok := checkAuth(w, r)
	if !ok {
		return
	}

	prog := r.FormValue("PROG")
	err := ioutil.WriteFile(user.progPath(), []byte(prog), 0600)

	if err != nil {
		fmt.Fprintf(w, "ERROR\n")
		fmt.Fprintf(os.Stderr, "ERROR %v\n", err)
	}

	fmt.Fprintf(w, "OK, %+v\n", user)

}

func get(w http.ResponseWriter, r *http.Request) {
	user, ok := checkAuth(w, r)
	if !ok {
		return
	}

	bot, exists := ctl.get(*user)
	if bot == nil || !exists {
		fmt.Fprintf(w, "NOT FOUND\n")
		return
	} else if bot.Destroyed {
		fmt.Fprintf(w, "DESTROYED")
		return
	}

	fmt.Fprintf(w, "%s\n", bot.Entity.(*Bot).JSON(bot))
}

func NewGame() *Game {
	return &Game{World: NewMap(1), Entities: make(Entities), EvHandler: block.EventHandler{}}
}

func getBot(user *User) *Control {
	prog, err := mio.ProgFromFile(user.progPath())
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "Spawned bot for %v\n", user.Name)

	//TODO: find safe spawn location and use it
	control, _ := NewEntity(game, Point{1, 1}, NewBot(prog))
	return control

}
