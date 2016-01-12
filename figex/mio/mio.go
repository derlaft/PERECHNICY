package mio

import (
	. "../asm"
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Prog struct {
	Prog  []Command
	Const map[string]int

	State *State
}

const MAXLEN = 2048

const (
	OP_OP       = 0
	OP_LABEL    = 1
	OP_CONSTANT = 2
	OP_NOP      = -1
)

func readLines(file *os.File) (str []string, err error) {
	str = make([]string, 512, MAXLEN)

	scanner := bufio.NewScanner(file)

	n := 0
	for ; scanner.Scan(); n++ {
		str[n] = scanner.Text()
	}

	return str[:n], nil
}

func ProgFromFile(path string) (prog *Prog, err error) {
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return nil, err
	}

	str, e := readLines(file)

	if e != nil {
		return &Prog{}, e
	}

	return preprocess(str), nil
}

func ProgFromFD(fd *os.File) (prog *Prog, err error) {
	str, e := readLines(fd)

	if e != nil {
		return &Prog{}, e
	}

	return preprocess(str), nil
}

func ProgFromString(s string) (prog *Prog, err error) {
	return preprocess(strings.Split(s, "\n")), nil
}

func preprocess(str []string) *Prog {
	prog := &Prog{}

	prog.Const = make(map[string]int)
	prog.Prog = make([]Command, 64, MAXLEN)

	n := 0

	for _, s := range str {
		tokens := tokenize(s)
		switch argType(tokens) {

		//@TODO: check conv error

		case OP_LABEL:
			name := tokens[0][0 : len(tokens[0])-1]
			prog.Const[name] = n - 1
		case OP_CONSTANT:
			value, _ := strconv.Atoi(tokens[1])
			prog.Const[tokens[0][1:]] = value
		case OP_OP:
			if isRealOp(tokens[0]) {
				n += 1
			}
		}
	}

	n = 0

	//TODO: use one for for everything?
	for _, s := range str {
		tokens := tokenize(s)
		if argType(tokens) == OP_OP && isRealOp(tokens[0]) {
			prog.Prog[n] = parseOp(prog, tokens)
			n += 1
		}
	}

	prog.Prog = prog.Prog[:n]

	prog.State = &State{}

	return prog
}

func isRealOp(name string) bool {
	_, contains := Handlers[name]

	return contains
}

func parseOp(prog *Prog, tokens []string) (args Command) {

	args.InstName = tokens[0]
	for _, arg := range tokens[1:] {
		err := pushArg(&args, prog, arg)
		if err != nil {
			fmt.Println("ERR", err)
			args.InstName = "FLT"
			return
		}
	}

	return args

}

func argType(t []string) int {

	if len(t) < 1 || len(t[0]) < 1 {
		return OP_NOP
	}

	first, last := t[0][0], t[0][len(t[0])-1]
	switch {
	case first == '#' && len(t) == 2:
		return OP_CONSTANT
	case last == ':' && len(t) == 1:
		return OP_LABEL
	case first == '%' || len(t[0]) == 0:
		return OP_NOP
	default:
		return OP_OP
	}
}

func tokenize(str string) []string {
	return strings.Fields(str)
}

func pushArg(org *Command, prog *Prog, arg string) error {

	var res byte

	switch arg[0] {
	//hex constant
	case '&':
		var err error
		res, err = getInt(arg)

		if err != nil {
			return err
		}

		push(org, ARG_CONST, int(res))
	case 'R':
		number, err := getInt(arg)
		if err != nil {
			return err
		}
		if number >= 16 {
			return errors.New("Nonexistent register usage")
		}

		push(org, ARG_REG, int(number))
	case '_', ':':
		val, ok := prog.Const[arg[1:]]
		if ok && int(byte(val)) == val && arg[0] == '_' {
			push(org, ARG_CONST, int(byte(val)))
		} else if ok && arg[0] == ':' {
			push(org, ARG_LABEL, val)
		} else {
			return errors.New("Nonexistent constant usage")
		}
	}

	return nil
}

func push(a *Command, t byte, val int) {
	a.Args[a.Used] = Argument{t, val}
	a.Used += 1
}

func getInt(arg string) (byte, error) {
	i, err := strconv.ParseInt(arg[1:], 16, 16)
	return byte(i), err
}

func (prog *Prog) Tick() *byte {

	if len(prog.Prog) == 0 {
		return nil
	}

	args := prog.Prog[prog.State.IP]
	prog.State.Cycle(args)

	prog.State.IP += 1

	// cycle code
	if prog.State.IP > len(prog.Prog)-1 {
		prog.State.IP = 0
	}

	hwint := prog.State.Interrupt
	prog.State.Interrupt = nil

	return hwint
}

func (prog *Prog) SoftwareInt(name string) {

	addr, found := prog.Const[name]
	if found {
		args := [2]Argument{
			{Type: ARG_LABEL, Val: addr},
			{},
		}
		prog.State.Cycle(Command{InstName: "JI", Args: args})

	} else {
		prog.Flt()
	}
}

func (prog *Prog) Flt() {
	cmd := Command{InstName: "FLT"}
	prog.State.Cycle(cmd)
}
