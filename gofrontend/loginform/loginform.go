package loginform

import (
	. "../grocessing"
	"../req"
	. "../ui"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"unicode"
)

var (
	SYMBOLS              = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	cursor               = 0
	nick                 = []rune("________")
	connectStatus string = PRESS_ENTER
)

const (
	PRESS_ENTER = "Press RETURN"
	LINE_WIDTH  = 20
)

func init() {
	var err error

	Server, err = req.Load(SERVER_URL, ConfigFile())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config file (%v)\n", err)
	}
	nick = []rune(Server.User)

	Forms[LOGIN_SCREEN] = LoginForm{}
}

type LoginForm struct {
}

func (f LoginForm) Setup() {
}
func (f LoginForm) Start() {
	connectStatus = PRESS_ENTER
}
func (f LoginForm) Stop() {
}

func (f LoginForm) KeyDown(key Key) {
	switch key {
	case KEY_UP:
		nick[cursor] = next(nick[cursor], -1)
	case KEY_DOWN:
		nick[cursor] = next(nick[cursor], +1)
	case KEY_LEFT:
		cursor = (cursor + NICK_LEN - 1) % NICK_LEN
	case KEY_RIGHT:
		cursor = (cursor + 1) % NICK_LEN
	case KEY_RETURN:
		connectStatus = "Connecting to the server..."
		go doRegister()
	default:
		if key == ' ' {
			key = '_'
		}

		if key >= 'a' && key <= 'z' ||
			key >= '0' && key <= '9' ||
			key == '_' {

			nick[cursor] = unicode.ToUpper(rune(key))
			cursor = (cursor + 1) % NICK_LEN
		}
	}
}

func (f LoginForm) Draw() {
	drawInput(Sz(0), Sz(1))
}

func drawInput(x, y int) {

	PushMatrix()
	Translate(x, y)
	Fill(Bright)

	Text("Input your name", 0, 0, Sz(25), Sz(1))
	Text("Use arrows to input", 0, Sz(1), Sz(25), Sz(1))

	Translate(0, Sz(2))

	Text(getnick(), 0, 0, Sz(25), Sz(1))
	Text(getcursor(), 0, Sz(1)/2, Sz(25), Sz(1))

	Translate(0, Sz(1))

	for i := 0; i <= len(connectStatus)/LINE_WIDTH; i++ {
		Text(connectStatus[i*LINE_WIDTH:Min(len(connectStatus), (i+1)*LINE_WIDTH)], 0, 0, Sz(25), Sz(1))
		Translate(0, Sz(1)/2)
	}

	PopMatrix()
}

func next(r rune, dx int) rune {
	i := strings.Index(SYMBOLS, string(r))
	ret := SYMBOLS[(len(SYMBOLS)+i+dx)%len(SYMBOLS)]
	return rune(ret)
}

func getnick() (n string) {
	for _, v := range nick {
		n += string(v)
	}
	return
}

func getcursor() (n string) {
	for i := 0; i < NICK_LEN; i++ {
		if i == cursor {
			n += "^"
		} else {
			n += " "
		}
	}
	return
}

func doRegister() {

	if Server.User != getnick() {
		Server = req.NewServer(SERVER_URL, getnick(), "")
	} else {
		Server.Check()
	}

	err := Server.Register()
	if err != nil {
		doStart()
		connectStatus = err.Error()
	}

	doStart()
}

func doStart() {

	Server.Credentials.Save(ConfigFile())

	f, err := os.Open(Prog)
	defer f.Close()
	if err != nil {
		connectStatus = err.Error()
		return
	}

	progData, err := ioutil.ReadAll(f)
	if err != nil {
		connectStatus = err.Error()
		return
	}

	err = Server.Start(string(progData))
	if err != nil {
		connectStatus = err.Error()
		return
	}

	fmt.Println("OKAY")
	Screen(GAME_SCREEN)

}
