package main

import (
	"./server"
	"fmt"
)

// cx, where are you?
// this is made by your design
// if you read this,
// please contact me
// <you know where I live> @ 2-47.ru

func main() {

	fmt.Println("Start server")
	server.Serve("0.0.0.0:4242")

}
