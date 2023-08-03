package pieces

import (
	"bufio"
	"engine2/board"
	"fmt"
	"log"
	_ "math"
	"math/bits"
	"os"
	"strings"
)

/* bb = bit board, cb = chessboard

Positive direction move generation uses bitscan-forward or bits.TrailingZeros(),
and the opposite for negative directions.
*/

// TODO: Investigate performance impact of branching in move gen.

func movePiece(from, to int, cb *board.Board, promoteTo ...string) {
	// TODO: Refactor to remove switch. Maybe make a parent array board.Occupied.
	movingType, err := getPieceType(from, cb)
	if err != nil {
		// TODO: Improve? Error is because the square does not exist (too high/low).
		fmt.Printf("invalid square to move from. square=%d", from)
		return
	}
	isValid := isValidMove(from, to, movingType, cb)
	if !isValid {
		if movingType == "" {
			fmt.Printf("no piece of the proper color on square %d", from)
		} else {
			fmt.Printf("invalid move for %v: from=%d, to=%d\n", movingType, from, to)
		}
		return
	}
	// TODO: Remove reliance on isValid from rest of func (find piece var in
	// a new func).
	// TODO: Test losing castling rights

	cb.EpSquare = 100

	fromBB := uint64(1 << from)
	toBB := uint64(1 << to)

	cb.BwPieces[cb.WToMove] ^= fromBB + toBB
	switch movingType {
	case "p":
		cb.BwPawns[cb.WToMove] ^= fromBB + toBB
		if to-from == 16 || to-from == -16 {
			cb.EpSquare = (to + from) / 2
		}
	case "n":
		cb.BwKnights[cb.WToMove] ^= fromBB + toBB
	case "b":
		cb.BwBishops[cb.WToMove] ^= fromBB + toBB
	case "r":
		cb.BwRooks[cb.WToMove] ^= fromBB + toBB
		if fromBB == 0 || fromBB == 1<<63 {
			cb.CastleRights[cb.WToMove][0] = false
		} else if fromBB == 1<<7 || fromBB == 1<<63 {
			cb.CastleRights[cb.WToMove][1] = false
		}
	case "q":
		cb.BwQueens[cb.WToMove] ^= fromBB + toBB
	case "k":
		cb.BwKing[cb.WToMove] ^= fromBB + toBB
		cb.KingSquare[cb.WToMove] = to
		cb.CastleRights[cb.WToMove][0] = false
		cb.CastleRights[cb.WToMove][1] = false
	default:
		// This branch should never execute.
		panic("empty or invalid piece type")
	}

	// Is this a capture (of a non-king piece)?
	opponent := 1 ^ cb.WToMove
	if toBB&(cb.BwPieces[opponent]^cb.BwKing[opponent]) != 0 {
		cb.BwPieces[opponent] ^= toBB

		switch {
		case toBB&cb.BwPawns[opponent] != 0:
			cb.BwPawns[opponent] ^= toBB
		case toBB&cb.BwKnights[opponent] != 0:
			cb.BwKnights[opponent] ^= toBB
		case toBB&cb.BwBishops[opponent] != 0:
			cb.BwBishops[opponent] ^= toBB
		case toBB&cb.BwRooks[opponent] != 0:
			// TODO: move castling checks to a less-frequented function
			// int type mixing here seems ok based on investigation
			if opponent == 0 && toBB == 1<<56 {
				cb.CastleRights[opponent][0] = false
			} else if opponent == 0 && toBB == 1<<63 {
				cb.CastleRights[opponent][1] = false
			} else if opponent == 1 && toBB == 0 {
				cb.CastleRights[opponent][0] = false
			} else if opponent == 1 && toBB == 1<<7 {
				cb.CastleRights[opponent][1] = false
			}
			cb.BwRooks[opponent] ^= toBB
		case toBB&cb.BwQueens[opponent] != 0:
			cb.BwQueens[opponent] ^= toBB
		default:
			panic("no captured piece bitboard matches")
		}
	}

	if len(promoteTo) == 1 || (movingType == "p" && (to < 8 || to > 55)) {
		promotePawn(toBB, cb, promoteTo[0])
	}

	cb.WToMove ^= 1
}

func promotePawn(toBB uint64, cb *board.Board, promoteTo ...string) {
	if len(promoteTo) == 1 {
		switch {
		case promoteTo[0] == "q":
			cb.BwQueens[cb.WToMove] ^= toBB
		case promoteTo[0] == "n":
			cb.BwKnights[cb.WToMove] ^= toBB
		case promoteTo[0] == "b":
			cb.BwBishops[cb.WToMove] ^= toBB
		case promoteTo[0] == "r":
			cb.BwRooks[cb.WToMove] ^= toBB
		default:
			panic("invalid promoteTo")
		}
	} else {
		fmt.Print("promote pawn to N, B, R, or Q: ")
		userPromote := getUserInput()

		if userPromote == "q" || userPromote == "n" || userPromote == "b" || userPromote == "r" {
			promotePawn(toBB, cb, userPromote)
		} else {
			fmt.Println("invalid promotion type, try again")
			promotePawn(toBB, cb)
		}
	}

	cb.BwPawns[cb.WToMove] ^= toBB
}

func getUserInput() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	err := scanner.Err()
	if err != nil {
		log.Println("failed to get input:", err)
		return getUserInput()
	}
	return strings.ToLower(scanner.Text())
}

func getPieceType(square int, cb *board.Board) (string, error) {
	// Return the piece type on a given square, or "" if the square is empty.
	// Only works for pieces of the moving side, cb.WToMove.
	if square < 0 || square > 63 {
		return "", fmt.Errorf("square %d does not exist", square)
	}
	squareBB := uint64(1 << square)

	switch {
	case squareBB&cb.BwPawns[cb.WToMove] != 0:
		return "p", nil
	case squareBB&cb.BwKnights[cb.WToMove] != 0:
		return "n", nil
	case squareBB&cb.BwBishops[cb.WToMove] != 0:
		return "b", nil
	case squareBB&cb.BwRooks[cb.WToMove] != 0:
		return "r", nil
	case squareBB&cb.BwQueens[cb.WToMove] != 0:
		return "q", nil
	case squareBB&cb.BwKing[cb.WToMove] != 0:
		return "k", nil
	default:
		return "", nil
	}
}

func isValidMove(from, to int, pieceType string, cb *board.Board) bool {
	// Use for user-submitted moves only?
	// Checks for blocking pieces and disallows captures of friendly pieces.
	// Does not consider check, pins, or legality of a pawn movement direction.
	if from < 0 || from > 63 || to < 0 || to > 63 || to == from {
		return false
	}
	toBB := uint64(1 << to)
	diff := to - from
	// to == from already excluded, no 0 move bugs from pawnDirections.
	pawnDirections := [2][4]int{{-7, -8, -9, -16},
		{7, 8, 9, 16},
	}

	switch pieceType {
	case "p":
		if !board.ContainsN(diff, pawnDirections[cb.WToMove]) {
			return false
		}
	case "n":
		if toBB&cb.NAttacks[from] == 0 {
			return false
		}
	case "b":
		if toBB&getBishopMoves(from, cb) == 0 {
			return false
		}
	case "r":
		if toBB&getRookMoves(from, cb) == 0 {
			return false
		}
	case "q":
		// Combined bishop and rook checks.
		if toBB&(getRookMoves(from, cb)|getBishopMoves(from, cb)) == 0 {
			return false
		}
	case "k":
		if toBB&getKingMoves(from, cb) == 0 {
			return false
		}
		// pieceType is not valid
	default:
		return false
	}

	// Friendly piece collision
	if toBB&cb.BwPieces[cb.WToMove] != 0 {
		return false
	}

	return true
}

// Captures and protection are included in move gen.
func getRookMoves(square int, cb *board.Board) uint64 {
	// North
	moves := cb.SlidingAttacks[0][square]
	blockers := cb.SlidingAttacks[0][square] & (cb.BwPieces[0] | cb.BwPieces[1])
	if blockers != 0 {
		blockerSq := bits.TrailingZeros64(blockers)
		moves ^= cb.SlidingAttacks[0][blockerSq]
	}
	// East
	moves |= cb.SlidingAttacks[2][square]
	blockers = cb.SlidingAttacks[2][square] & (cb.BwPieces[1] | cb.BwPieces[0])
	if blockers != 0 {
		blockerSq := bits.TrailingZeros64(blockers)
		moves ^= cb.SlidingAttacks[2][blockerSq]
	}
	// South
	moves |= cb.SlidingAttacks[4][square]
	blockers = cb.SlidingAttacks[4][square] & (cb.BwPieces[1] | cb.BwPieces[0])
	if blockers != 0 {
		blockerSq := 63 - bits.LeadingZeros64(blockers)
		moves ^= cb.SlidingAttacks[4][blockerSq]
	}
	// West
	moves |= cb.SlidingAttacks[6][square]
	blockers = cb.SlidingAttacks[6][square] & (cb.BwPieces[1] | cb.BwPieces[0])
	if blockers != 0 {
		blockerSq := 63 - bits.LeadingZeros64(blockers)
		moves ^= cb.SlidingAttacks[6][blockerSq]
	}

	return moves
}

func getPawnMoves(square int, cb *board.Board) uint64 {
	var moves uint64
	opponent := 1 ^ cb.WToMove

	if square < 8 || square > 55 {
		panic("pawns can't be on the first or last rank")
	}

	moves = cb.PAttacks[cb.WToMove][square] & (cb.BwPieces[opponent] | uint64(1<<cb.EpSquare))

	var dir, low, high int
	if cb.WToMove == 1 {
		dir = 8
		low = 7
		high = 16
	} else {
		dir = -8
		low = 47
		high = 56
	}
	if low < square && square < high {
		moves |= (1<<(square+dir) | 1<<(square+2*dir)) & ^(cb.BwPieces[0] | cb.BwPieces[1])
	} else {
		moves |= 1 << (square + dir) & ^(cb.BwPieces[0] | cb.BwPieces[1])
	}

	return moves
}

func getKnightMoves(square int, cb *board.Board) uint64 {
	return cb.NAttacks[square]
}

func getBishopMoves(square int, cb *board.Board) uint64 {
	// Northeast
	moves := cb.SlidingAttacks[1][square]
	blockers := cb.SlidingAttacks[1][square] & (cb.BwPieces[1] | cb.BwPieces[0])
	if blockers != 0 {
		blockerSq := bits.TrailingZeros64(blockers)
		moves ^= cb.SlidingAttacks[1][blockerSq]
	}
	// Southeast
	moves |= cb.SlidingAttacks[3][square]
	blockers = cb.SlidingAttacks[3][square] & (cb.BwPieces[1] | cb.BwPieces[0])
	if blockers != 0 {
		blockerSq := 63 - bits.LeadingZeros64(blockers)
		moves ^= cb.SlidingAttacks[3][blockerSq]
	}
	// Southwest
	moves |= cb.SlidingAttacks[5][square]
	blockers = cb.SlidingAttacks[5][square] & (cb.BwPieces[1] | cb.BwPieces[0])
	if blockers != 0 {
		blockerSq := 63 - bits.LeadingZeros64(blockers)
		moves ^= cb.SlidingAttacks[5][blockerSq]
	}
	// Northwest
	moves |= cb.SlidingAttacks[7][square]
	blockers = cb.SlidingAttacks[7][square] & (cb.BwPieces[1] | cb.BwPieces[0])
	if blockers != 0 {
		blockerSq := bits.TrailingZeros64(blockers)
		moves ^= cb.SlidingAttacks[7][blockerSq]
	}

	return moves
}

func getQueenMoves(square int, cb *board.Board) uint64 {
	return getRookMoves(square, cb) | getBishopMoves(square, cb)
}

func getKingMoves(square int, cb *board.Board) uint64 {
	// TODO: King cannot move to squares attacked by opponent.
	moves := cb.KAttacks[square]

	// TODO: A king cannot castle out of, through, or into an attacked square.
	// CastleRights true when king and rook(s) have not moved or been captured.
	if cb.WToMove == 0 {
		if cb.CastleRights[0][0] && (1<<57+1<<58+1<<59)&(cb.BwPieces[0]|cb.BwPieces[1]) == 0 {
			moves += 1 << 58
		}
		if cb.CastleRights[0][1] && (1<<61+1<<62)&(cb.BwPieces[0]|cb.BwPieces[1]) == 0 {
			moves += 1 << 62
		}
	} else {
		if cb.CastleRights[1][0] && (1+1<<2+1<<3)&(cb.BwPieces[0]|cb.BwPieces[1]) == 0 {
			moves += 1 << 2
		}
		if cb.CastleRights[1][1] && (1<<5+1<<6)&(cb.BwPieces[0]|cb.BwPieces[1]) == 0 {
			moves += 1 << 6
		}
	}

	return moves
}

func read1Bits(bb uint64) []int {
	// Using TrailingZeros64() seems as fast as bitshifting right while bb>0.
	squares := []int{}
	for bb > 0 {
		squares = append(squares, bits.TrailingZeros64(bb))
		bb &= bb - 1
	}
	return squares
}

func binSearch(n int, nums [8]int) bool {
	l := 0
	r := len(nums) - 1
	var mid int
	for l <= r {
		mid = (l + r) / 2
		switch {
		case n > nums[mid]:
			l = mid + 1
		case n < nums[mid]:
			r = mid - 1
		default:
			return true
		}
	}

	return false
}
