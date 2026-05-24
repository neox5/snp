package main

import (
	"fmt"
	"os"
)

const maxDepth = 1

func main() {
	if len(os.Args) < 2 {
		return
	}
	srcDir := os.Args[1]

	fmt.Println(srcDir)
}
