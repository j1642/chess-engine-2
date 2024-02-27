// Move generation
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

/*
bb = bitboard, cb = chessboard
Magic numbers 0, ..., 63 and 1<<0, ..., 1<<63 are squares of the chessboard.
*/

func MovePiece(move board.Move, cb *board.Board) {
	// TODO: Refactor to remove switch. Maybe make a parent array board.Occupied
	fromBB := uint64(1 << move.From)
	toBB := uint64(1 << move.To)

	if toBB&(cb.Pieces[1^cb.WToMove]^cb.Kings[1^cb.WToMove]) != 0 {
		capturePiece(toBB, cb)
	}

	cb.Pieces[cb.WToMove] ^= fromBB + toBB
	switch move.Piece {
	case "p":
		cb.Pawns[cb.WToMove] ^= fromBB + toBB
		if move.To-move.From == 16 || move.To-move.From == -16 {
			cb.EpSquare = (move.To + move.From) / 2
		} else if move.To < 8 || move.To > 55 {
			promotePawn(toBB, cb, move.PromoteTo)
			cb.EpSquare = 100
		} else if move.To == cb.EpSquare {
			captureSq := move.To + 8
			if cb.WToMove == 1 {
				captureSq = move.To - 8
			}
			cb.Pawns[1^cb.WToMove] ^= uint64(1 << captureSq)
			cb.Pieces[1^cb.WToMove] ^= uint64(1 << captureSq)
			cb.EpSquare = 100
		} else {
			cb.EpSquare = 100
		}
	case "n":
		cb.Knights[cb.WToMove] ^= fromBB + toBB
		cb.EpSquare = 100
	case "b":
		cb.Bishops[cb.WToMove] ^= fromBB + toBB
		cb.EpSquare = 100
	case "r":
		cb.Rooks[cb.WToMove] ^= fromBB + toBB
		if move.From == 0 || move.From == 56 {
			cb.CastleRights[cb.WToMove][0] = false
		} else if move.From == 7 || move.From == 63 {
			cb.CastleRights[cb.WToMove][1] = false
		}
		cb.EpSquare = 100
	case "q":
		cb.Queens[cb.WToMove] ^= fromBB + toBB
		cb.EpSquare = 100
	case "k":
		if move.To-move.From == 2 || move.To-move.From == -2 {
			if cb.CastleRights[cb.WToMove][0] && (move.To == 2 || move.To == 58) {
				cb.Rooks[cb.WToMove] ^= uint64(1<<(move.To-2) + 1<<(move.To+1))
				cb.Pieces[cb.WToMove] ^= uint64(1<<(move.To-2) + 1<<(move.To+1))
			} else if cb.CastleRights[cb.WToMove][1] && (move.To == 6 || move.To == 62) {
				cb.Rooks[cb.WToMove] ^= uint64(1<<(move.To+1) + 1<<(move.To-1))
				cb.Pieces[cb.WToMove] ^= uint64(1<<(move.To+1) + 1<<(move.To-1))
			} else {
				panic("king moving two squares, but is not castling")
			}
		}
		cb.Kings[cb.WToMove] ^= fromBB + toBB
		cb.KingSqs[cb.WToMove] = move.To
		cb.CastleRights[cb.WToMove][0] = false
		cb.CastleRights[cb.WToMove][1] = false
		cb.EpSquare = 100
	default:
		panic("empty or invalid piece type")
	}

	cb.PrevMove = move
	cb.WToMove ^= 1
}

func capturePiece(squareBB uint64, cb *board.Board) {
	opponent := 1 ^ cb.WToMove
	cb.Pieces[opponent] ^= squareBB

	switch {
	case squareBB&cb.Pawns[opponent] != 0:
		cb.Pawns[opponent] ^= squareBB
	case squareBB&cb.Knights[opponent] != 0:
		cb.Knights[opponent] ^= squareBB
	case squareBB&cb.Bishops[opponent] != 0:
		cb.Bishops[opponent] ^= squareBB
	case squareBB&cb.Rooks[opponent] != 0:
		// int type mixing here seems ok based on investigation
		if opponent == 0 && squareBB == 1<<56 {
			cb.CastleRights[opponent][0] = false
		} else if opponent == 0 && squareBB == 1<<63 {
			cb.CastleRights[opponent][1] = false
		} else if opponent == 1 && squareBB == 0 {
			cb.CastleRights[opponent][0] = false
		} else if opponent == 1 && squareBB == 1<<7 {
			cb.CastleRights[opponent][1] = false
		}
		cb.Rooks[opponent] ^= squareBB
	case squareBB&cb.Queens[opponent] != 0:
		cb.Queens[opponent] ^= squareBB
	default:
		panic("no captured piece bitboard matches")
	}
}

func promotePawn(toBB uint64, cb *board.Board, promoteTo ...string) {
	// TODO: Else never triggers b/c move.promoteTo always has a string
	// Change to 'if promotoTo != ""'
	if len(promoteTo) == 1 {
		switch {
		case promoteTo[0] == "q":
			cb.Queens[cb.WToMove] ^= toBB
		case promoteTo[0] == "n":
			cb.Knights[cb.WToMove] ^= toBB
		case promoteTo[0] == "b":
			cb.Bishops[cb.WToMove] ^= toBB
		case promoteTo[0] == "r":
			cb.Rooks[cb.WToMove] ^= toBB
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

	cb.Pawns[cb.WToMove] ^= toBB
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

// Return the piece type on a given square, or "" if the square is empty.
// Only works for pieces of the moving side, cb.WToMove.
func getPieceType(square int, cb *board.Board) (string, error) {
	if square < 0 || square > 63 {
		return "", fmt.Errorf("square %d does not exist", square)
	}
	squareBB := uint64(1 << square)

	switch {
	case squareBB&cb.Pawns[cb.WToMove] != 0:
		return "p", nil
	case squareBB&cb.Knights[cb.WToMove] != 0:
		return "n", nil
	case squareBB&cb.Bishops[cb.WToMove] != 0:
		return "b", nil
	case squareBB&cb.Rooks[cb.WToMove] != 0:
		return "r", nil
	case squareBB&cb.Queens[cb.WToMove] != 0:
		return "q", nil
	case squareBB&cb.Kings[cb.WToMove] != 0:
		return "k", nil
	default:
		return "", nil
	}
}

// Use for user-submitted moves only?
// Checks for blocking pieces and disallows captures of friendly pieces.
// Does not consider check, pins, or legality of a pawn movement direction.
func isValidMove(from, to int, pieceType string, cb *board.Board) bool {
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
		if toBB&(getRookMoves(from, cb)|getBishopMoves(from, cb)) == 0 {
			return false
		}
	case "k":
		cb.Pieces[cb.WToMove] ^= uint64(1 << cb.KingSqs[cb.WToMove])
		cb.WToMove ^= 1
		attkSquares := GetAttackedSquares(cb)
		cb.WToMove ^= 1
		cb.Pieces[cb.WToMove] ^= uint64(1 << cb.KingSqs[cb.WToMove])
		if toBB&getKingMoves(from, attkSquares, cb) == 0 {
			return false
		}
	default:
		// pieceType is not valid
		return false
	}

	// Friendly piece collision
	if toBB&cb.Pieces[cb.WToMove] != 0 {
		return false
	}

	return true
}

// Captures and protection are included in move gen.
func getRookMoves(square int, cb *board.Board) uint64 {
	occupied := cb.Pieces[0] | cb.Pieces[1]
	// North
	moves := cb.SlidingAttacks[0][square]
	blockers := cb.SlidingAttacks[0][square] & occupied
	blockerSq := bits.TrailingZeros64(blockers | uint64(1<<63))
	moves ^= cb.SlidingAttacks[0][blockerSq]
	// East
	moves |= cb.SlidingAttacks[2][square]
	blockers = cb.SlidingAttacks[2][square] & occupied
	blockerSq = bits.TrailingZeros64(blockers | uint64(1<<63))
	moves ^= cb.SlidingAttacks[2][blockerSq]
	// South
	moves |= cb.SlidingAttacks[4][square]
	blockers = cb.SlidingAttacks[4][square] & occupied
	blockerSq = 63 - bits.LeadingZeros64(blockers|uint64(1))
	moves ^= cb.SlidingAttacks[4][blockerSq]
	// West
	moves |= cb.SlidingAttacks[6][square]
	blockers = cb.SlidingAttacks[6][square] & occupied
	blockerSq = 63 - bits.LeadingZeros64(blockers|uint64(1))
	moves ^= cb.SlidingAttacks[6][blockerSq]

	return moves
}

func getPawnMoves(square int, cb *board.Board) uint64 {
	opponent := 1 ^ cb.WToMove

	if square < 8 || square > 55 {
		panic("pawns can't be on the first or last rank")
	}

	moves := cb.PAttacks[cb.WToMove][square] & (cb.Pieces[opponent] | uint64(1<<cb.EpSquare))

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
	occupied := cb.Pieces[0] | cb.Pieces[1]

	if low < square && square < high && 1<<(square+dir)&occupied == 0 {
		moves |= (1<<(square+dir) + 1<<(square+2*dir)) & ^occupied
	} else {
		moves |= 1 << (square + dir) & ^occupied
	}

	return moves
}

func getKnightMoves(square int, cb *board.Board) uint64 {
	return cb.NAttacks[square]
}

func getBishopMoves(square int, cb *board.Board) uint64 {
	occupied := cb.Pieces[0] | cb.Pieces[1]
	// Northeast
	moves := cb.SlidingAttacks[1][square]
	blockers := cb.SlidingAttacks[1][square] & occupied
	blockerSq := bits.TrailingZeros64(blockers | uint64(1<<63))
	moves ^= cb.SlidingAttacks[1][blockerSq]
	// Southeast
	moves |= cb.SlidingAttacks[3][square]
	blockers = cb.SlidingAttacks[3][square] & occupied
	blockerSq = 63 - bits.LeadingZeros64(blockers|uint64(1))
	moves ^= cb.SlidingAttacks[3][blockerSq]
	// Southwest
	moves |= cb.SlidingAttacks[5][square]
	blockers = cb.SlidingAttacks[5][square] & occupied
	blockerSq = 63 - bits.LeadingZeros64(blockers|uint64(1))
	moves ^= cb.SlidingAttacks[5][blockerSq]
	// Northwest
	moves |= cb.SlidingAttacks[7][square]
	blockers = cb.SlidingAttacks[7][square] & occupied
	blockerSq = bits.TrailingZeros64(blockers | uint64(1<<63))
	moves ^= cb.SlidingAttacks[7][blockerSq]

	return moves
}

func getQueenMoves(square int, cb *board.Board) uint64 {
	return getRookMoves(square, cb) | getBishopMoves(square, cb)
}

// Return legal king moves.
func getKingMoves(square int, oppAttackedSquares uint64, cb *board.Board) uint64 {
	occupied := cb.Pieces[0] | cb.Pieces[1]
	moves := cb.KAttacks[square] & ^oppAttackedSquares & ^cb.Pieces[cb.WToMove]

	if cb.WToMove == 0 {
		if cb.CastleRights[0][0] && (1<<57+1<<58+1<<59)&occupied == 0 &&
			(1<<58+1<<59+1<<60)&oppAttackedSquares == 0 {
			moves += 1 << 58
		}
		if cb.CastleRights[0][1] && (1<<61+1<<62)&occupied == 0 &&
			(1<<60+1<<61+1<<62)&oppAttackedSquares == 0 {
			moves += 1 << 62
		}
	} else {
		if cb.CastleRights[1][0] && (1<<1+1<<2+1<<3)&occupied == 0 &&
			(1<<2+1<<3+1<<4)&oppAttackedSquares == 0 {
			moves += 1 << 2
		}
		if cb.CastleRights[1][1] && (1<<5+1<<6)&occupied == 0 &&
			(1<<4+1<<5+1<<6)&oppAttackedSquares == 0 {
			moves += 1 << 6
		}
	}

	return moves
}

// Return the set of squares attacked by color cb.WToMove
func GetAttackedSquares(cb *board.Board) uint64 {
	// TODO: Is there a way to avoid reading 1 bits when accumulating moves?
	var pieces []int
	attackSquares := uint64(0)

	pieces = Read1BitsPawns(cb.Pawns[cb.WToMove])
	for _, square := range pieces {
		// Do not include pawn pushes.
		attackSquares |= cb.PAttacks[cb.WToMove][square]
	}

	pieces = read1Bits(cb.Knights[cb.WToMove])
	for _, square := range pieces {
		attackSquares |= cb.NAttacks[square]
	}
	pieces = read1Bits(cb.Bishops[cb.WToMove])
	for _, square := range pieces {
		attackSquares |= getBishopMoves(square, cb)
	}
	pieces = read1Bits(cb.Rooks[cb.WToMove])
	for _, square := range pieces {
		attackSquares |= getRookMoves(square, cb)
	}
	pieces = read1Bits(cb.Queens[cb.WToMove])
	for _, square := range pieces {
		attackSquares |= getQueenMoves(square, cb)
	}
	// Do not include castling.
	attackSquares |= cb.KAttacks[cb.KingSqs[cb.WToMove]]

	return attackSquares
}

type moveGenFunc func(int, *board.Board) uint64
type readBitsFunc func(uint64) []int

func GetAllMoves(cb *board.Board) []board.Move {
	// Return slice of all pseudo-legal moves for color cb.WToMove (king moves
	// are strictly legal)
	cb.Pieces[cb.WToMove] ^= uint64(1 << cb.KingSqs[cb.WToMove])
	cb.WToMove ^= 1
	attackedSquares := GetAttackedSquares(cb)
	cb.WToMove ^= 1
	cb.Pieces[cb.WToMove] ^= uint64(1 << cb.KingSqs[cb.WToMove])

	var capturesBlks uint64
	var attackerCount int
	if cb.Kings[cb.WToMove]&attackedSquares != 0 {
		capturesBlks, attackerCount = getCheckingSquares(cb)
	}

	kingSq := cb.KingSqs[cb.WToMove]
	moves := read1Bits(getKingMoves(kingSq, attackedSquares, cb) & ^cb.Pieces[cb.WToMove])

	allMoves := make([]board.Move, len(moves), 35)
	for i, toSq := range moves {
		allMoves[i] = board.Move{From: kingSq, To: toSq, Piece: "k", PromoteTo: ""}
	}
	// If attackerCount > 1 and king has no moves, it is checkmate
	if attackerCount > 1 {
		return allMoves
	}

	pieces := []uint64{cb.Pawns[cb.WToMove], cb.Knights[cb.WToMove],
		cb.Bishops[cb.WToMove], cb.Rooks[cb.WToMove], cb.Queens[cb.WToMove],
	}
	moveFuncs := []moveGenFunc{getPawnMoves, getKnightMoves, getBishopMoves,
		getRookMoves, getQueenMoves,
	}
	symbols := []string{"p", "n", "b", "r", "q"}

	// 29% perft() speed up and -40% malloc from having this loop in this function
	for i, piece := range pieces {
		for _, fromSq := range read1Bits(piece) {
			moves := read1Bits(moveFuncs[i](fromSq, cb) & ^cb.Pieces[cb.WToMove])
			for _, toSq := range moves {
				if i == 0 && (toSq < 8 || toSq > 55) &&
					(capturesBlks == 0 || uint64(1<<toSq)&capturesBlks != 0) {
					allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: "p", PromoteTo: "n"})
					allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: "p", PromoteTo: "b"})
					allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: "p", PromoteTo: "r"})
					allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: "p", PromoteTo: "q"})
				} else if capturesBlks == 0 || uint64(1<<toSq)&capturesBlks != 0 {
					allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: symbols[i], PromoteTo: ""})
				}
			}
		}
	}

	return allMoves
}

// Return the set of squares of pieces checking the king and interposition
// squares, and the number of checking pieces.
func getCheckingSquares(cb *board.Board) (uint64, int) {
	opponent := 1 ^ cb.WToMove
	attackerCount := 0

	kSquare := cb.KingSqs[cb.WToMove]
	pAttackers := cb.PAttacks[cb.WToMove][kSquare] & cb.Pawns[opponent]
	knightAttackers := cb.NAttacks[kSquare] & cb.Knights[opponent]
	bqAttackers := getBishopMoves(kSquare, cb) & (cb.Bishops[opponent] |
		cb.Queens[opponent])
	orthogAttackers := getRookMoves(cb.KingSqs[cb.WToMove], cb) &
		(cb.Rooks[opponent] | cb.Queens[opponent])

		// TODO: Remove king check?
	if cb.Kings[opponent]&cb.KAttacks[cb.KingSqs[cb.WToMove]] != 0 {
		fmt.Println(cb.KingSqs)
		cb.Print()
		panic("king is checking the other king")
	}
	if len(read1Bits(knightAttackers)) > 1 {
		cb.Print()
		panic(">1 knights are checking the king")
	}

	// There should be 0 or 1 attackers in each attack group.
	if pAttackers != 0 {
		attackerCount += 1
	}
	if knightAttackers != 0 {
		attackerCount += 1
	}

	panicMsgs := []string{">1 piece is checking king orthogonally",
		">1 piece is checking king diagonally"}
	attackers := []uint64{orthogAttackers, bqAttackers}

	// Add interposition squares if any exist.
	for i, attacker := range attackers {
		if attacker != 0 {
			attackerSquares := read1Bits(attacker)
			attackerCount += len(attackerSquares)
			// Possible optimization: check if attackerCount + len(attackers) > 1 before the loop
			if len(attackerSquares) > 1 {
				panic(panicMsgs[i])
			}
			dir := findDirection(cb.KingSqs[cb.WToMove], attackerSquares[0])
			attackers[i] = fillFromTo(cb.KingSqs[cb.WToMove], attackerSquares[0], dir)
		}
	}

	return pAttackers | knightAttackers | attackers[0] | attackers[1], attackerCount
}

func fillFromTo(from, to, direction int) uint64 {
	// Return a bitboard of squares between 'from and 'to', excluding 'from'
	// and including 'to'.
	bb := uint64(0)
	for sq := from + direction; sq != to; sq += direction {
		bb += 1 << sq
	}
	bb += 1 << to

	return bb
}

func findDirection(from, to int) int {
	// Return the direction from one square to another.
	// Assumes (from, to) is an orthogonal or diagonal move.
	var dir int
	diff := to - from
	// TODO: Change to lookup table for files, or use bitboards.
	files := board.GetFiles()

	switch {
	case diff%8 == 0:
		dir = 8
	case diff%9 == 0:
		dir = 9
	case -6 <= diff && diff <= 6:
		dir = 1
	case diff%7 == 0:
		fromInAFile := board.ContainsN(from, files[0])
		fromInHFile := board.ContainsN(from, files[3])
		toInAFile := board.ContainsN(to, files[0])
		toInHFile := board.ContainsN(to, files[3])
		if (fromInAFile && toInHFile) || (fromInHFile && toInAFile) {
			dir = 1
		} else {
			dir = 7
		}
	default:
		panic("invalid toSquare-fromSquare difference")
	}
	if diff < 0 {
		dir *= -1
	}

	return dir
}

func read1Bits(bb uint64) []int {
	// Using TrailingZeros64() seems as fast as bitshifting right while bb>0.
	squares := make([]int, 0, 4)
	for bb > 0 {
		squares = append(squares, bits.TrailingZeros64(bb))
		bb &= bb - 1
	}
	return squares
}

func Read1BitsPawns(bb uint64) []int {
	squares := make([]int, 0, 8)
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
