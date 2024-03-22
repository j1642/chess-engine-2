// Universal chess interface protocol
package uci

import (
	"github.com/j1642/chess-engine-2/board"
	"github.com/j1642/chess-engine-2/pieces"

	"fmt"
	"log"
	"os"
	"strings"
)

// Receive a message from the chess GUI and return a response
func processMessage(s string) {
	split := strings.Fields(s)
	switch split[0] {
	case "uci":
		fmt.Println("id name chess-engine-2")
		fmt.Println("id author j1642")
		// would send available options here is they existed
		fmt.Println("uciok")
	case "debug":
		// "debug on" prints more info to the GUI. Can be sent while calculating. Off by default
	case "isready":
		// respond immediately if calculating, wait to send if doing something
		// time-consuming like setting up tablebases
		fmt.Println("readyok")
	case "setoption":
		// Change engine settings. No settings are implemented at the moment
	case "register":
		// This engine does not require a username or code to work
		fmt.Println("registration ok")
	case "ucinewgame":
		// Next position and search will be a different game. Ignore but maybe clear tTable?
	case "position":
		currentPosition = buildPosition(split)
	case "go":
		// TODO
	case "stop":
		// Keep the best move and stop calculating
	case "ponderhit":
		// Ignore because ponder is not implemented
	case "quit":
		// TODO: will all goroutines stop upon exiting
		os.Exit(0)
	case "d":
		currentPosition.Print()
	}
}

var currentPosition *board.Board

// Return a new board.Board
func buildPosition(split []string) *board.Board {
	// "position startpos moves e2e4 e7e5"
	// "position fen ... moves e2e4"
	var cb *board.Board
	var movesIdx int
	for i, s := range split {
		if s == "moves" {
			movesIdx = i
			break
		}
	}

	if split[1] == "fen" {
		var err error
		if movesIdx != 0 {
			cb, err = board.FromFen(strings.Join(split[2:movesIdx], " "))
		} else {
			cb, err = board.FromFen(strings.Join(split[2:], " "))
		}
		if err != nil {
			// Invalid FEN, return empty board
			log.Println("uci calling FromFen:", err) // TODO: remove
			return cb
		}
	} else if split[1] == "startpos" {
		cb = board.New()
	} else {
		return cb
	}

	// Make moves, if provided
	if movesIdx > 1 {
		moves := split[movesIdx+1:]
		for _, move := range moves {
			chars := []rune(move)
			fromSq := int8((chars[1]-'1')*8 + chars[0] - 'a')
			toSq := int8((chars[3]-'1')*8 + chars[2] - 'a')
			pieceType, err := identifyPieceOnSquare(fromSq, cb)
			log.Println("from, to, piece:", fromSq, toSq, pieceType)

			if err != nil {
				// Invalid move, return the board as-is
				log.Println("uci identifying piece:", err) // TODO: remove
				return cb
			}
			isValidMove := pieces.IsValidMove(fromSq, toSq, pieceType, cb)
			if !isValidMove {
				log.Printf("invalid move, to:%d, from:%d, piece:%d", toSq, fromSq, pieceType) // TODO: remove
				return cb
			}

			move := board.Move{
				From:      fromSq,
				To:        toSq,
				Piece:     pieceType,
				PromoteTo: pieces.NO_PIECE,
			}
			log.Printf("move: %+v\n", move)
			pieces.MovePiece(
				move,
				cb,
			)
		}

	}

	return cb
}

func identifyPieceOnSquare(square int8, cb *board.Board) (uint8, error) {
	/*if square == 12 || square == 52 {
	      cb.Print()
	  }
	  if square == 12 && cb.Rooks[cb.WToMove] & uint64(1<<square) != 0 {
	      os.Exit(1)
	  }*/
	// uint64 casting is necessary, otherwise returns false data without a compiler error
	switch {
	case cb.Pawns[cb.WToMove]&uint64(1<<square) != 0:
		return pieces.PAWN, nil
	case cb.Knights[cb.WToMove]&uint64(1<<square) != 0:
		return pieces.KNIGHT, nil

	case cb.Bishops[cb.WToMove]&uint64(1<<square) != 0:
		return pieces.BISHOP, nil
	case cb.Rooks[cb.WToMove]&uint64(1<<square) != 0:
		return pieces.ROOK, nil

	case cb.Queens[cb.WToMove]&uint64(1<<square) != 0:
		return pieces.QUEEN, nil
	case cb.KingSqs[cb.WToMove] == square:
		return pieces.KING, nil
	default:
		return pieces.NO_PIECE,
			fmt.Errorf("no piece of the color to move (%d) on square %d",
				cb.WToMove, square)
	}
}
