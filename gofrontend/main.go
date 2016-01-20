package main

import (
	_ "./explorer/"
	_ "./gameform/"
	_ "./loginform/"
	_ "./mainmenu/"
	"./ui"
	"fmt"
	flag "github.com/ogier/pflag"
	"os"
	"path/filepath"
)

func main() {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	os.Chdir(dir)

	prog := flag.String("prog", "", "Prog to launch")
	chunked := flag.String("chunked", "", "Chunk to edit")
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
		//@TODO
	} else if *view {
		ui.Screen(ui.EXPLORER_SCREEN)
	}
	ui.Start()
}
