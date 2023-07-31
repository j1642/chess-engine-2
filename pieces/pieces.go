package pieces

import (
	"engine2/board"
	"fmt"
	_ "log"
	_ "math"
	"math/bits"
)

/* bb = bit board, cb = chessboard

Positive direction move generation uses bitscan-forward or bits.TrailingZeros(),
and the opposite for negative directions.
*/

// TODO: Investigate performance impact of branching in move gen.

func movePiece(from, to int, cb *board.Board) {
	// TODO: Refactor. Maybe by making a parent array board.occupied.
	isValid, piece := isValidMove(from, to, cb)
	if !isValid {
		fmt.Printf("invalid move for %v: from=%d, to=%d\n", piece, from, to)
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

	cb.WToMove ^= 1
}

func isValidMove(from, to int, cb *board.Board) (bool, string) {
	// Does not check for check, pins, blocking pieces, or legality of a double
	// pawn push. Castling is currently always invalid.
	// Use for user-submitted moves only?
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
			fmt.Println("invalid move: pawns cannot move like that")
			return false, "p"
		}
		piece = "p"
	case fromBB&cb.BwKnights[cb.WToMove] > 0:
		if toBB&cb.NAttacks[from] == 0 {
			fmt.Println("invalid move: knights cannot move like that")
			return false, "n"
		}
		piece = "n"
	case fromBB&cb.BwBishops[cb.WToMove] > 0:
		if toBB&(cb.SlidingAttacks[1][from]|cb.SlidingAttacks[3][from]|
			cb.SlidingAttacks[5][from]|cb.SlidingAttacks[7][from]) == 0 {
			fmt.Println("invalid move: bishops cannot move like that")
			return false, "b"
		}
		piece = "b"
	case fromBB&cb.BwRooks[cb.WToMove] > 0:
		if toBB&(cb.SlidingAttacks[0][from]|cb.SlidingAttacks[2][from]|
			cb.SlidingAttacks[4][from]|cb.SlidingAttacks[6][from]) == 0 {

			fmt.Println("invalid move: rooks cannot move like that")
			return false, "r"
		}
		piece = "r"
	case fromBB&cb.BwQueens[cb.WToMove] > 0:
		// Combined bishop and rook checks.
		if toBB&(cb.SlidingAttacks[1][from]|cb.SlidingAttacks[3][from]|
			cb.SlidingAttacks[5][from]|cb.SlidingAttacks[7][from]|

			cb.SlidingAttacks[0][from]|cb.SlidingAttacks[2][from]|
			cb.SlidingAttacks[4][from]|cb.SlidingAttacks[6][from]) == 0 {

			fmt.Println("invalid move: queens cannot move like that")
			return false, "q"
		}
		piece = "q"
	case fromBB&cb.BwKing[cb.WToMove] > 0:
		if toBB&cb.KAttacks[from] == 0 {
			fmt.Println("invalid move: kings cannot move like that")
			return false, "k"
		}
		piece = "k"
	default:
		fmt.Printf("invalid move: no piece of color %v on square %d\n", cb.WToMove, from)
		return false, ""
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
