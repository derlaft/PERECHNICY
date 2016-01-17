package main

import (
	_ "./gameform/"
	_ "./loginform/"
	"./ui"
	"fmt"
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

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Please specify prog file\n")
		return
	}

	ui.Prog = os.Args[1]
	ui.Screen(ui.LOGIN_SCREEN)
	ui.Start()
}
