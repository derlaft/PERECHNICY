package forms

import (
	. "client/grocessing"
	"client/request"
	. "client/ui"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"unicode"
)

var (
	SYMBOLS = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
)

const (
	PRESS_ENTER = "Press RETURN"
	LINE_WIDTH  = 20
)

func init() {
	Forms[LOGIN_SCREEN] = &LoginForm{}
}

type LoginForm struct {
	cursor        uint
	nick          []rune
	connectStatus string
}

func (e *LoginForm) Setup() {
	var err error

	Server, err = request.Load(SERVER_URL, ConfigFile())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config file (%v)\n", err)
		Server = request.NewServer(SERVER_URL, "", "")
	}

	//@TODO: clean solution
	if len(Server.User) != 8 {
		Server.User = "________"
	}

	e.nick = []rune(Server.User)
}

func (e *LoginForm) Start() {
	e.connectStatus = PRESS_ENTER
}

func (e *LoginForm) Stop() {
}

func (e *LoginForm) KeyDown(key Key) {
	switch key {
	case KEY_UP:
		e.nick[e.cursor] = next(e.nick[e.cursor], -1)
	case KEY_DOWN:
		e.nick[e.cursor] = next(e.nick[e.cursor], +1)
	case KEY_LEFT:
		e.cursor = (e.cursor + NICK_LEN - 1) % NICK_LEN
	case KEY_RIGHT:
		e.cursor = (e.cursor + 1) % NICK_LEN
	case KEY_RETURN:
		e.connectStatus = "Connecting to the server..."
		go e.doRegister()
	default:
		if key == ' ' {
			key = '_'
		}

		if key >= 'a' && key <= 'z' ||
			key >= '0' && key <= '9' ||
			key == '_' {

			e.nick[e.cursor] = unicode.ToUpper(rune(key))
			e.cursor = (e.cursor + 1) % NICK_LEN
		}
	}
}

func (e *LoginForm) Draw() {
	e.drawInput(Sz(0), Sz(1))
}

func (e *LoginForm) drawInput(x, y int) {

	PushMatrix()
	Translate(x, y)
	Fill(Bright)

	Text("Input your name", 0, 0, Sz(25), Sz(1))
	Text("Use arrows to input", 0, Sz(1), Sz(25), Sz(1))

	Translate(0, Sz(2))

	Text(e.getnick(), 0, 0, Sz(25), Sz(1))
	Text(e.getcursor(), 0, Sz(1)/2, Sz(25), Sz(1))

	Translate(0, Sz(1))

	for i := 0; i <= len(e.connectStatus)/LINE_WIDTH; i++ {
		Text(e.connectStatus[i*LINE_WIDTH:Min(len(e.connectStatus), (i+1)*LINE_WIDTH)], 0, 0, Sz(25), Sz(1))
		Translate(0, Sz(1)/2)
	}

	PopMatrix()
}

func next(r rune, dx int) rune {
	i := strings.Index(SYMBOLS, string(r))
	ret := SYMBOLS[(len(SYMBOLS)+i+dx)%len(SYMBOLS)]
	return rune(ret)
}

func (e *LoginForm) getnick() (n string) {
	for _, v := range e.nick {
		n += string(v)
	}
	return
}

func (e *LoginForm) getcursor() (n string) {
	for i := uint(0); i < NICK_LEN; i++ {
		if i == e.cursor {
			n += "^"
		} else {
			n += " "
		}
	}
	return
}

func (e *LoginForm) doRegister() {

	nick := e.getnick()

	if Server.User != nick {
		Server = request.NewServer(SERVER_URL, nick, "")
	} else {
		Server.Check()
	}

	err := Server.Register()
	if err != nil {
		e.connectStatus = err.Error()
		return
	}

	e.doStart()
}

func (e *LoginForm) doStart() {

	Server.Credentials.Save(ConfigFile())

	f, err := os.Open(Prog)
	defer f.Close()
	if err != nil {
		e.connectStatus = err.Error()
		return
	}

	progData, err := ioutil.ReadAll(f)
	if err != nil {
		e.connectStatus = err.Error()
		return
	}

	err = Server.Start(string(progData))
	if err != nil {
		e.connectStatus = err.Error()
		return
	}

	fmt.Println("OKAY")
	Screen(GAME_SCREEN)

}
