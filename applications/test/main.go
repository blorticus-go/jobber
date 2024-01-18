package main

import (
	"flag"
	"fmt"
)

func main() {
	var s string

	flag.StringVar(&s, "s", "", "s")
	flag.Parse()

	fmt.Printf("s = (%s)\n", s)
	fmt.Println("args = ", flag.Args())
}
