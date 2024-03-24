// Universal chess interface protocol
package uci

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	_ "time"

	"github.com/j1642/chess-engine-2/board"
	"github.com/j1642/chess-engine-2/engine"
	"github.com/j1642/chess-engine-2/pieces"
)

var currentPosition = board.New()
var stop chan bool

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
		// TODO: "debug on" prints more info to the GUI. Can be sent while calculating. Off by default
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
		//go calculate(split)
		calculate(split)
	case "stop":
		// Keep the best move and stop calculating
		stop <- true
	case "ponderhit":
		// Ignore because ponder is not implemented
	case "quit":
		// TODO: will all goroutines stop upon exiting
		os.Exit(0)
	case "d":
		currentPosition.Print()
	default:
	}
}

// Return a new board.Board. Input can be in one of two formats: "position
// startpos moves e2e4 e7e5" or "position fen ... moves e2e4"
func buildPosition(split []string) *board.Board {
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
			fromSq, toSq, promoteTo, err := convertLongAlgebraicMoveToSquares(move)
			pieceType, err := identifyPieceOnSquare(fromSq, cb)
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

			pieces.MovePiece(
				board.Move{
					From:      fromSq,
					To:        toSq,
					Piece:     pieceType,
					PromoteTo: promoteTo,
				},
				cb,
			)
		}
	}

	return cb
}

func identifyPieceOnSquare(square int8, cb *board.Board) (uint8, error) {
	// uint64 casting is necessary, otherwise this returns false data
	squareBB := uint64(1 << square)
	switch {
	case cb.Pawns[cb.WToMove]&squareBB != 0:
		return pieces.PAWN, nil
	case cb.Knights[cb.WToMove]&squareBB != 0:
		return pieces.KNIGHT, nil

	case cb.Bishops[cb.WToMove]&squareBB != 0:
		return pieces.BISHOP, nil
	case cb.Rooks[cb.WToMove]&squareBB != 0:
		return pieces.ROOK, nil

	case cb.Queens[cb.WToMove]&squareBB != 0:
		return pieces.QUEEN, nil
	case cb.KingSqs[cb.WToMove] == square:
		return pieces.KING, nil
	default:
		return pieces.NO_PIECE,
			fmt.Errorf("no piece of the color to move (%d) on square %d",
				cb.WToMove, square)
	}
}

// Search the current position for the best move
func calculate(split []string) {
	options := buildGoOptions(split)
	// All move in options.searchmoves should be legal when they are appeded
	// TODO: add stop channel for STOP command and ticker to control INFO prints
	engine.IterativeDeepening(currentPosition, int(options.depth))
}

type goOptions struct {
	searchmoves                                                       []board.Move
	wtime, btime, binc, winc, movestogo, depth, nodes, mate, movetime uint64
	infinite                                                          bool
}

// Parse options following the "go ..." command. Options include depth, nodes,
// searchmoves, infinite, movestogo,
func buildGoOptions(split []string) goOptions {
	// TODO:
	//   wtime 2000 btime 1000
	//   binc 100 winc 100
	//   mate 10 (mate search 10 moves deep, maybe 20 ply?
	//   movetime 500 (search 500ms then return best move)
	options := goOptions{}
	for i, s := range split {
		switch s {
		case "infinite":
			options.infinite = true
		case "depth":
			depth, err := strconv.ParseUint(split[i+1], 10, 0)
			if err != nil {
				log.Println("buildGoOptions depth:", err)
				continue
			}
			options.depth = depth
		case "nodes":
			nodes, err := strconv.ParseUint(split[i+1], 10, 0)
			if err != nil {
				log.Println("buildGoOptions nodes:", err)
				continue
			}
			options.nodes = nodes
		case "movestogo":
			movestogo, err := strconv.ParseUint(split[i+1], 10, 0)
			if err != nil {
				log.Println("buildGoOptions movestogo:", err)
				continue
			}
			options.movestogo = movestogo
		case "searchmoves":
			// If an error occurs or an invalid move is submitted, stop adding moves
			for idx := i + 1; ; idx++ {
				fromSq, toSq, promoteTo, err := convertLongAlgebraicMoveToSquares(split[idx])
				if err != nil {
					break
				}
				currentPosition.Print()
				pieceType, err := identifyPieceOnSquare(fromSq, currentPosition)
				if err != nil {
					break
				}
				if pieces.IsValidMove(fromSq, toSq, pieceType, currentPosition) {
					options.searchmoves = append(
						options.searchmoves,
						board.Move{
							From:      fromSq,
							To:        toSq,
							Piece:     pieceType,
							PromoteTo: promoteTo,
						},
					)
				} else {
					break
				}
			}
		}
	}

	return options
}

// Return fromSqare, toSquare, promoteTo, err. Input formats: "a1h8", "a7a8q"
func convertLongAlgebraicMoveToSquares(move string) (int8, int8, uint8, error) {
	// TODO: think about a way to streamline this func
	fromSq := int8((move[1]-'1')*8 + move[0] - 'a')
	toSq := int8((move[3]-'1')*8 + move[2] - 'a')
	promoteTo := pieces.NO_PIECE

	err := fmt.Errorf("placeholder")
	if 'a' <= move[0] && move[0] <= 'h' &&
		'1' <= move[1] && move[1] <= '8' &&
		'a' <= move[2] && move[2] <= 'h' &&
		'1' <= move[3] && move[3] <= '8' {
		if len(move) == 4 {
			err = nil
		} else if len(move) == 5 {
			switch move[4] {
			case 'n':
				promoteTo = pieces.KNIGHT
				err = nil
			case 'b':
				promoteTo = pieces.BISHOP
				err = nil
			case 'r':
				promoteTo = pieces.ROOK
				err = nil
			case 'q':
				promoteTo = pieces.QUEEN
				err = nil
			default:
				err = fmt.Errorf("invalid long algebraic move: %s", move)
			}
			// If promoting, check for valid ranks
			if !((move[1] == '7' && move[3] == '8') ||
				(move[1] == '2' && move[3] == '1')) {
				err = fmt.Errorf("invalid long algebraic move: %s", move)
			}
		} else {
			err = fmt.Errorf("invalid long algebraic move: %s", move)
		}
	} else {
		err = fmt.Errorf("invalid long algebraic move: %s", move)
	}
	return fromSq, toSq, promoteTo, err
}
