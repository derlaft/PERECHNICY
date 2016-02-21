package forms

import (
	. "client/grocessing"
	. "client/ui"
	"fmt"
	"math/rand"
)

const ()

type entry struct {
	id    Formid
	Title string
}

var (
	MENU_ENTRIES []entry = []entry{
		{LOGIN_SCREEN, "LOG IN"},
		{EXPLORER_SCREEN, "VIEW MAP"},
		{EXIT, "EXIT"},
	}

	SLOGANS []string = []string{
		"The great and powerful!",
		"2.0.1-git",
	}

	slogan string

	cursor int = 0
)

type mainmenuForm struct {
}

func init() {
	Forms[MAINMENU_SCREEN] = mainmenuForm{}
}

func (e mainmenuForm) KeyDown(key Key) {
	switch key {
	case KEY_DOWN:
		cursor = (cursor + 1) % len(MENU_ENTRIES)
	case KEY_UP:
		cursor = (len(MENU_ENTRIES) + cursor - 1) % len(MENU_ENTRIES)
	case KEY_RETURN:
		Screen(MENU_ENTRIES[cursor].id)
	}
}

func (e mainmenuForm) Draw() {
	PushMatrix()
	Background(Dark)
	Fill(Bright)
	Translate(0, Sz(1))
	Text("PERECHNICY", 0, 0, Sz(25), Sz(1))
	Translate(0, Sz(1))
	Text(slogan, 0, 0, Sz(25), Sz(1))
	Translate(0, Sz(2))

	for i, v := range MENU_ENTRIES {
		var todraw string
		if cursor == i {
			todraw = fmt.Sprintf("> %v <", v.Title)
		} else {
			todraw = v.Title
		}

		Text(todraw, 0, 0, Sz(25), Sz(1))
		Translate(0, Sz(1))
	}

	PopMatrix()
}

func (e mainmenuForm) Setup() {
	Title("PERECHNICY")
}

func (e mainmenuForm) Start() {
	slogan = SLOGANS[rand.Intn(len(SLOGANS))]
}
func (e mainmenuForm) Stop() {
}
