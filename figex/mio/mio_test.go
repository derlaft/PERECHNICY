package mio

import (
	"fmt"
	"testing"
)

func TestOpen(t *testing.T) {

	prog, err := ProgFromFile("./test.txt")
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}

	for _, v := range prog.Prog {
		fmt.Println(v.InstName)
	}

}

func TestTokenize(t *testing.T) {

	s := "\t\tMOV  AL \tAL "
	a := tokenize(s)

	if a[0] != "MOV" || a[1] != "AL" || a[2] != "AL" {
		t.Fail()
	}
}
