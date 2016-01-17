package server

import (
	"../block"
	. "../chunk"
	. "../entity/entities"
	"../figex/mio"
	. "../game"
	. "../stuff"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mrand "math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	game   *Game
	ctl    EntityControl
	Progs  ProgDB
	Tokens TokenDB
)

const (
	TOKEN_SIZE = 20
)

type Bots map[User]*Control
type ProgDB map[User]*mio.Prog
type TokenDB map[User]string

type DB map[User]DBEntry

type DBEntry struct {
	Bot   *Control
	Prog  *mio.Prog
	Token string
	sync.RWMutex
}

type EntityControl struct {
	bots Bots
	sync.RWMutex
}

type User struct {
	Name string
}

func Serve(addr string) {
	Progs = make(ProgDB)
	Tokens = make(TokenDB)

	gameInit()
	httpInit(addr)
}

func httpInit(addr string) {
	http.HandleFunc("/start", start)
	http.HandleFunc("/register", register)
	http.HandleFunc("/get", get)
	http.HandleFunc("/ping", ping)
	http.ListenAndServe(addr, nil)
}

func gameInit() {
	mrand.Seed(time.Now().UnixNano())

	game = NewGame(NewMap(1), block.EventHandler{})
	ctl = EntityControl{bots: make(Bots)}

	game.World.GetChunk(Point{0, 0}).Data[2][2] = 9
	go gameLoop()

	NewEntity(game, Point{2, 2}, &Bear{})
}

func gameLoop() {
	ticks := uint64(0)
	diff := int64(0)

	for {
		select {
		case <-time.After(time.Second / 100):
			diff -= time.Now().UnixNano()
			game.Tick()
			diff += time.Now().UnixNano()
			ticks += 1
			if diff >= 1000 {
				fmt.Printf("%v\r", ticks)
				ticks = 0
				diff = 0
			}
		}
	}

}

func checkAuth(w http.ResponseWriter, r *http.Request) (*User, bool) {

	r.ParseMultipartForm(32 << 20)

	user, token := r.FormValue("User"), r.FormValue("Token")
	userobj := User{Name: user}

	if user == "" || Tokens[userobj] != token {
		fmt.Fprintf(w, `{"Result": false, "Reason": "Bad auth"}`)
		return nil, false
	}

	return &userobj, true
}

func register(w http.ResponseWriter, r *http.Request) {
	user := User{Name: r.FormValue("User")}
	fmt.Println("COCO", user.Name+"\n\n\n")

	_, registered := Tokens[user]
	if !registered {
		newtoken, err := randToken()
		if err == nil {
			fmt.Fprintf(w, `{"Result": true, "Token": "%v"}`, newtoken)
			Tokens[user] = newtoken
			return
		}
		//failed to generate token
		fmt.Fprintf(w, `{"Result": false, "Reason": "Internal server error"}`)
	} else {
		fmt.Fprintf(w, `{"Result": false, "Reason": "Already registered"}`)
	}
}

func start(w http.ResponseWriter, r *http.Request) {

	user, ok := checkAuth(w, r)
	if !ok {
		return
	}

	// check if already started
	bot, exists := ctl.get(*user)
	if exists && !bot.Destroyed {
		// setting result to true as we can still continue
		fmt.Fprintf(w, `{"Result": true, "Reason": "Already exists"}`)
		return
	}

	prog := r.FormValue("Prog")
	user.addProg(prog)

	// start
	bot = getBot(user)
	ctl.Lock()
	ctl.bots[*user] = bot
	ctl.Unlock()

	fmt.Fprintf(w, `{"Result": true}`)
}

func ping(w http.ResponseWriter, r *http.Request) {

	_, ok := checkAuth(w, r)
	if !ok {
		fmt.Fprintf(w, `{"Result": false}`)
		return
	}

	fmt.Fprintf(w, `{"Result": true}`)
}

func (c *EntityControl) get(u User) (*Control, bool) {
	ctl.RLock()
	bot, exists := ctl.bots[u]
	ctl.RUnlock()
	return bot, exists
}

func get(w http.ResponseWriter, r *http.Request) {
	user, ok := checkAuth(w, r)
	if !ok {
		return
	}

	bot, exists := ctl.get(*user)
	if bot == nil || !exists {
		fmt.Fprintf(w, `{"Result": false, "Reason": "NOT FOUND"}`)
		return
	}

	fmt.Fprintf(w, "%s\n", bot.Entity.(*Bot).JSON(bot))
}

func getBot(user *User) *Control {
	prog, found := Progs[*user]
	if !found {
		panic("Prog not found")
	}
	fmt.Fprintf(os.Stderr, "Spawned bot for %v\n", user.Name)

	//TODO: find safe spawn location and use it
	control, _ := NewEntity(game, Point{1, 1}, NewBot(prog))
	return control

}

func (u *User) addProg(p string) error {
	prog, err := mio.ProgFromString(p)
	Progs[*u] = prog
	return err
}

func randToken() (string, error) {
	var buf [TOKEN_SIZE]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(buf[:]), nil
}
