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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	game *Game
	DB   *database
)

const (
	TOKEN_SIZE = 20
)

type database struct {
	DB map[User]*DBEntry
	sync.RWMutex
}

type DBEntry struct {
	Bot   *Control
	Prog  *mio.Prog
	Token string
	sync.RWMutex
}

type User struct {
	Name string
}

func newDatabase() *database {
	r := database{}

	r.DB = make(map[User]*DBEntry)

	return &r
}

func Serve(addr string) {

	DB = newDatabase()

	gameInit()
	httpInit(addr)
}

func httpInit(addr string) {
	http.HandleFunc("/start", start)
	http.HandleFunc("/register", register)
	http.HandleFunc("/get", get)
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/map", map_view)
	http.ListenAndServe(addr, nil)
}

func gameInit() {
	game = NewGame(NewMap(100500, int(time.Now().Unix())), block.EventHandler{})

	//game.World.GetChunk(Point{0, 0}).Data[2][2] = 9
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
				//fmt.Printf("%v\r", ticks)
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

	entry, exists := DB.DB[userobj]
	var good_token string
	if exists {
		entry.RLock()
		good_token = entry.Token
		entry.RUnlock()
	}

	if user == "" || !exists || token != good_token {
		fmt.Fprintf(w, `{"Result": false, "Reason": "Bad auth"}`)
		return nil, false
	}

	return &userobj, true
}

func register(w http.ResponseWriter, r *http.Request) {
	user := User{Name: r.FormValue("User")}
	fmt.Println("COCO", user.Name+"\n\n\n")

	_, registered := DB.DB[user]
	if !registered {
		newtoken, err := randToken()
		if err == nil {
			fmt.Fprintf(w, `{"Result": true, "Token": "%v"}`, newtoken)
			DB.Lock()
			DB.DB[user] = &DBEntry{Token: newtoken}
			DB.Unlock()
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
	entry, exists := DB.DB[*user]
	bot := entry.Bot
	if exists && bot != nil && !bot.Destroyed {
		// setting result to true as we can still continue
		fmt.Fprintf(w, `{"Result": true, "Reason": "Already exists"}`)
		return
	}

	prog := r.FormValue("Prog")
	user.addProg(prog)

	// start
	bot = getBot(user)
	entry.Lock()
	entry.Bot = bot
	entry.Unlock()

	fmt.Fprintf(w, `{"Result": true}`)
}

func map_view(w http.ResponseWriter, r *http.Request) {

	x, err1 := strconv.ParseInt(r.FormValue("X"), 10, 64)
	y, err2 := strconv.ParseInt(r.FormValue("Y"), 10, 64)
	width, err3 := strconv.ParseInt(r.FormValue("W"), 10, 64)
	height, err4 := strconv.ParseInt(r.FormValue("H"), 10, 64)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		fmt.Fprintf(w, `[]`)
		return
	}

	from := Point{x, y}
	to := Point{x + width - 1, y + height - 1}
	len := (width) * (height)
	out := make([]int, len, len)
	i := 0

	for pt := range EachPoint(from, to) {
		out[i] = int(game.At(*pt))
		i += 1
	}

	ret, err := json.Marshal(out)
	if err != nil {
		fmt.Fprintf(w, `[]`)
		return
	}

	fmt.Fprintf(w, "%v", string(ret))
}

func ping(w http.ResponseWriter, r *http.Request) {

	_, ok := checkAuth(w, r)
	if !ok {
		fmt.Fprintf(w, `{"Result": false}`)
		return
	}

	fmt.Fprintf(w, `{"Result": true}`)
}

func get(w http.ResponseWriter, r *http.Request) {
	user, ok := checkAuth(w, r)
	if !ok {
		return
	}

	entry, exists := DB.DB[*user]
	bot := entry.Bot
	if !exists || bot == nil || bot == nil {
		fmt.Fprintf(w, `{"Result": false, "Reason": "NOT FOUND"}`)
		return
	}

	fmt.Fprintf(w, "%s", bot.Entity.(*Bot).JSON(bot))
}

func getBot(user *User) *Control {
	entry, found := DB.DB[*user]
	entry.RLock()
	prog := entry.Prog
	entry.RUnlock()
	if !found || prog == nil {
		panic("Prog not found")
	}
	fmt.Fprintf(os.Stderr, "Spawned bot for %v\n", user.Name)

	//TODO: find safe spawn location and use it
	control, _ := NewEntity(game, Point{1, 1}, NewBot(prog))
	return control

}

func (u *User) addProg(p string) error {
	prog, err := mio.ProgFromString(p)
	entry := DB.DB[*u]
	entry.Lock()
	entry.Prog = prog
	entry.Unlock()

	fmt.Println("COCO", DB.DB[*u].Prog)
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
