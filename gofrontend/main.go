package main

import (
	_ "./gameform/"
	_ "./loginform/"
	"./ui"
)

func main() {

	ui.Screen(ui.LOGIN_SCREEN)
	ui.Start()
}
