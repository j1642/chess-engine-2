package pieces

import (
	"bufio"
	"engine2/board"
	"fmt"
	_ "log"
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
	isValid, piece := isValidMove(from, to, cb)
	if !isValid {
		if piece == "" {
			fmt.Printf("no piece of the proper color on square %d", from)
		} else {
			fmt.Printf("invalid move for %v: from=%d, to=%d\n", piece, from, to)
		}
		return
	}

	fromBB := uint64(1 << from)
	toBB := uint64(1 << to)

	cb.BwPieces[cb.WToMove] ^= fromBB + toBB
	switch {
	case piece == "p":
		cb.BwPawns[cb.WToMove] ^= fromBB + toBB
	case piece == "n":
		cb.BwKnights[cb.WToMove] ^= fromBB + toBB
	case piece == "b":
		cb.BwBishops[cb.WToMove] ^= fromBB + toBB
	case piece == "r":
		cb.BwRooks[cb.WToMove] ^= fromBB + toBB
	case piece == "q":
		cb.BwQueens[cb.WToMove] ^= fromBB + toBB
	case piece == "k":
		cb.BwKing[cb.WToMove] ^= fromBB + toBB
		cb.KingSquare[cb.WToMove] = to
	default:
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
			cb.BwRooks[opponent] ^= toBB
		case toBB&cb.BwQueens[opponent] != 0:
			cb.BwQueens[opponent] ^= toBB
		default:
			panic("no captured piece bitboard matches")
		}
	}

	if len(promoteTo) == 1 || (piece == "p" && (to < 8 || to > 55)) {
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
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		err := scanner.Err()
		if err != nil {
			panic(err)
		}
		userPromote := strings.ToLower(scanner.Text())

		if userPromote == "q" || userPromote == "n" || userPromote == "b" || userPromote == "r" {
			promotePawn(toBB, cb, userPromote)
		} else {
			fmt.Println("invalid promotion type, try again")
			promotePawn(toBB, cb)
		}
	}

	cb.BwPawns[cb.WToMove] ^= toBB
}

func isValidMove(from, to int, cb *board.Board) (bool, string) {
	// Use for user-submitted moves only?
	// Checks for blocking pieces and disallows captures of friendly pieces.
	// Does not consider check, pins, or legality of a pawn movement direction.
	// Castling is currently always invalid.
	if from < 0 || from > 63 || to < 0 || to > 63 || to == from {
		return false, ""
	}
	fromBB := uint64(1 << from)
	toBB := uint64(1 << to)
	diff := to - from
	var piece string
	// to == from already excluded, no 0 move bugs from pawnDirections.
	pawnDirections := [2][8]int{{-7, -8, -9, -16, 0, 0, 0, 0},
		{7, 8, 9, 16, 0, 0, 0, 0},
	}

	switch {
	case fromBB&cb.BwPawns[cb.WToMove] > 0:
		if !board.ContainsN(diff, pawnDirections[cb.WToMove]) {
			return false, "p"
		}
		piece = "p"
	case fromBB&cb.BwKnights[cb.WToMove] > 0:
		if toBB&cb.NAttacks[from] == 0 {
			return false, "n"
		}
		piece = "n"
	case fromBB&cb.BwBishops[cb.WToMove] > 0:
		if toBB&getBishopMoves(from, cb) == 0 {
			return false, "b"
		}
		piece = "b"
	case fromBB&cb.BwRooks[cb.WToMove] > 0:
		if toBB&getRookMoves(from, cb) == 0 {
			return false, "r"
		}
		piece = "r"
	case fromBB&cb.BwQueens[cb.WToMove] > 0:
		// Combined bishop and rook checks.
		if toBB&(getRookMoves(from, cb)|getBishopMoves(from, cb)) == 0 {
			return false, "q"
		}
		piece = "q"
	case fromBB&cb.BwKing[cb.WToMove] > 0:
		if toBB&cb.KAttacks[from] == 0 {
			return false, "k"
		}
		piece = "k"
	default:
		return false, ""
	}

	// Friendly piece collision
	if toBB&cb.BwPieces[cb.WToMove] != 0 {
		return false, piece
	}

	return true, piece
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
	// TODO: Castling.
	return cb.KAttacks[square]
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
