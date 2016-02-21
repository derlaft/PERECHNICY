package main

import (
	_ "client/forms"
	"client/ui"
	"fmt"
	"os"

	flag "github.com/ogier/pflag"
)

func main() {

	prog := flag.String("prog", "", "Prog to launch")
	chunked := flag.String("edit", "", "Chunk to edit")
	view := flag.Bool("explorer", false, "View map mode")

	flag.Parse()

	if *prog == "" && *chunked == "" && !*view {
		fmt.Fprintf(os.Stderr, "WAT DO YOU WANT\n")
		return
	}

	if *prog != "" {
		ui.Prog = *prog
		ui.Screen(ui.MAINMENU_SCREEN)
	} else if *chunked != "" {
		ui.Screen(ui.EDITOR_SCREEN)
	} else if *view {
		ui.Screen(ui.EXPLORER_SCREEN)
	}
	ui.Start()
}
