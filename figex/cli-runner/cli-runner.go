package main

import (
	. "../mio"
	"bufio"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <file>\n", os.Args[0])
		os.Exit(1)
	}
	fname := os.Args[1]

	prog, err := ProgFromFile(fname)

	if err != nil {
		fmt.Println(err)
	}

	scanner := bufio.NewScanner(os.Stdin)

	for i, in := range prog.Prog {
		fmt.Println(i, in.InstName)
	}

	for scanner.Scan() {
		args := prog.Prog[prog.State.IP]
		fmt.Printf("%q %d %s\n", prog.State.Reg, prog.State.IP, args.InstName)
		prog.Tick()
	}

}
