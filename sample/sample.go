package main

import (
	"fmt"
	"os"
)

func main() {
	s := "world"
	if len(os.Args) > 1 {
		s = os.Args[1]
	}
	fmt.Println("hello", s)
}
