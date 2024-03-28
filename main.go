package main

import (
	/*
			    "github.com/j1642/chess-engine-2/pieces"
			    "math/bits"
			    "github.com/j1642/chess-engine-2/board"
		        "fmt"
		        "time"
	*/
	"bufio"
	"github.com/j1642/chess-engine-2/uci"
	"os"
)

func main() {
	// TODO: add new func pieces.getCaptureMoves(cb) -> []board.Move
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		uci.ProcessMessage(line)
	}
}
