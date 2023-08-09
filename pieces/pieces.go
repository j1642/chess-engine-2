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
//      - branching effects in getRookMoves() not obvious with perft(5)

type move struct {
	from, to  int
	piece     string
	promoteTo string
}

func movePiece(move move, cb *board.Board) {
	// TODO: Refactor to remove switch. Maybe make a parent array board.Occupied.

	// Temporary check for perft
	if !isValidMove(move.from, move.to, move.piece, cb) {
		fmt.Println(move.from, move.to, move.piece)
		panic("illegal move")
	}

	cb.EpSquare = 100

	fromBB := uint64(1 << move.from)
	toBB := uint64(1 << move.to)

	if toBB&(cb.BwPieces[1^cb.WToMove]^cb.BwKing[1^cb.WToMove]) != 0 {
		capturePiece(toBB, cb)
	}

	cb.BwPieces[cb.WToMove] ^= fromBB + toBB
	switch move.piece {
	case "p":
		cb.BwPawns[cb.WToMove] ^= fromBB + toBB
		if move.to-move.from == 16 || move.to-move.from == -16 {
			cb.EpSquare = (move.to + move.from) / 2
		}
		if move.to < 8 || move.to > 55 {
			promotePawn(toBB, cb, move.promoteTo)
		}
	case "n":
		cb.BwKnights[cb.WToMove] ^= fromBB + toBB
	case "b":
		cb.BwBishops[cb.WToMove] ^= fromBB + toBB
	case "r":
		cb.BwRooks[cb.WToMove] ^= fromBB + toBB
		if move.from == 0 || fromBB == 63 {
			cb.CastleRights[cb.WToMove][0] = false
		} else if move.from == 7 || move.from == 63 {
			cb.CastleRights[cb.WToMove][1] = false
		}
	case "q":
		cb.BwQueens[cb.WToMove] ^= fromBB + toBB
	case "k":
		if move.to-move.from == 2 || move.to-move.from == -2 {
			if cb.CastleRights[cb.WToMove][0] && (move.to == 2 || move.to == 58) {
				cb.BwRooks[cb.WToMove] ^= uint64(1<<(move.to-2) + 1<<(move.to+1))
			} else if cb.CastleRights[cb.WToMove][1] && (move.to == 6 || move.to == 62) {
				cb.BwRooks[cb.WToMove] ^= uint64(1<<(move.to+1) + 1<<(move.to-1))
			} else {
				panic("king moving two squares, but is not castling")
			}
		}
		cb.BwKing[cb.WToMove] ^= fromBB + toBB
		cb.KingSquare[cb.WToMove] = move.to
		cb.CastleRights[cb.WToMove][0] = false
		cb.CastleRights[cb.WToMove][1] = false
	default:
		// This branch should never execute.
		panic("empty or invalid piece type")
	}

	cb.WToMove ^= 1
}

func capturePiece(squareBB uint64, cb *board.Board) {
	opponent := 1 ^ cb.WToMove
	cb.BwPieces[opponent] ^= squareBB

	switch {
	case squareBB&cb.BwPawns[opponent] != 0:
		cb.BwPawns[opponent] ^= squareBB
	case squareBB&cb.BwKnights[opponent] != 0:
		cb.BwKnights[opponent] ^= squareBB
	case squareBB&cb.BwBishops[opponent] != 0:
		cb.BwBishops[opponent] ^= squareBB
	case squareBB&cb.BwRooks[opponent] != 0:
		// TODO: move castling checks to a less-frequented function
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
		cb.BwRooks[opponent] ^= squareBB
	case squareBB&cb.BwQueens[opponent] != 0:
		cb.BwQueens[opponent] ^= squareBB
	default:
		panic("no captured piece bitboard matches")
	}
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
	default:
		// pieceType is not valid
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
	occupied := cb.BwPieces[0] | cb.BwPieces[1]

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
	// Return legal king moves.
	cb.WToMove ^= 1
	opponentAttackedSquares := getAttackedSquares(cb)
	cb.WToMove ^= 1

	occupied := cb.BwPieces[0] | cb.BwPieces[1]
	moves := cb.KAttacks[square] & ^opponentAttackedSquares & ^cb.BwPieces[cb.WToMove]

	if cb.WToMove == 0 {
		if cb.CastleRights[0][0] && (1<<57+1<<58+1<<59)&occupied == 0 &&
			(1<<58+1<<59+1<<60)&opponentAttackedSquares == 0 {
			moves += 1 << 58
		}
		if cb.CastleRights[0][1] && (1<<61+1<<62)&occupied == 0 &&
			(1<<60+1<<61+1<<62)&opponentAttackedSquares == 0 {
			moves += 1 << 62
		}
	} else {
		if cb.CastleRights[1][0] && (1<<1+1<<2+1<<3)&occupied == 0 &&
			(1<<2+1<<3+1<<4)&opponentAttackedSquares == 0 {
			moves += 1 << 2
		}
		if cb.CastleRights[1][1] && (1<<5+1<<6)&occupied == 0 &&
			(1<<4+1<<5+1<<6)&opponentAttackedSquares == 0 {
			moves += 1 << 6
		}
	}

	return moves
}

func getAttackedSquares(cb *board.Board) uint64 {
	// TODO: Is there a way to avoid reading 1 bits when accumulating moves?
	pieces := []int{}
	attackSquares := uint64(0)

	// TODO: Try to refactor without using a switch statement.
	pieces = read1Bits(cb.BwPawns[cb.WToMove])
	for _, square := range pieces {
		// Do not include pawn pushes.
		attackSquares |= cb.PAttacks[cb.WToMove][square]
	}
	pieces = read1Bits(cb.BwKnights[cb.WToMove])
	for _, square := range pieces {
		attackSquares |= getKnightMoves(square, cb)
	}
	pieces = read1Bits(cb.BwBishops[cb.WToMove])
	for _, square := range pieces {
		attackSquares |= getBishopMoves(square, cb)
	}
	pieces = read1Bits(cb.BwRooks[cb.WToMove])
	for _, square := range pieces {
		attackSquares |= getRookMoves(square, cb)
	}
	pieces = read1Bits(cb.BwQueens[cb.WToMove])
	for _, square := range pieces {
		attackSquares |= getQueenMoves(square, cb)
	}
	// Do not include castling.
	attackSquares |= cb.KAttacks[cb.KingSquare[cb.WToMove]]

	return attackSquares
}

func getAllMoves(cb *board.Board) []move {
	var capturesBlocks uint64
	var attackerCount int
	allMoves := make([]move, 0, 35)

	cb.WToMove ^= 1
	attackedSquares := getAttackedSquares(cb)
	cb.WToMove ^= 1
	if cb.BwKing[cb.WToMove]&attackedSquares != 0 {
		capturesBlocks, attackerCount = getCheckingSquares(cb)
	}

	kingSquare := cb.KingSquare[cb.WToMove]
	moves := read1Bits(getKingMoves(kingSquare, cb) & ^cb.BwPieces[cb.WToMove])
	for _, toSquare := range moves {
		allMoves = append(allMoves, move{kingSquare, toSquare, "k", ""})
	}
	// If attackerCount > 1 and king has no moves, it is checkmate.
	if attackerCount > 1 {
		return allMoves
	}

	allMoves = append(allMoves,
		getPieceMoveList(cb.BwPawns[cb.WToMove], capturesBlocks, "p", getPawnMoves, cb)...,
	)
	allMoves = append(allMoves,
		getPieceMoveList(cb.BwKnights[cb.WToMove], capturesBlocks, "n", getKnightMoves, cb)...,
	)
	allMoves = append(allMoves,
		getPieceMoveList(cb.BwBishops[cb.WToMove], capturesBlocks, "b", getBishopMoves, cb)...,
	)
	allMoves = append(allMoves,
		getPieceMoveList(cb.BwRooks[cb.WToMove], capturesBlocks, "r", getRookMoves, cb)...,
	)
	allMoves = append(allMoves,
		getPieceMoveList(cb.BwQueens[cb.WToMove], capturesBlocks, "q", getQueenMoves, cb)...,
	)

	return allMoves
}

type moveGenFunc func(int, *board.Board) uint64

func getPieceMoveList(piecesBB, capturesBlocks uint64, piece string,
	moveGen moveGenFunc, cb *board.Board) []move {
	// Return slice of all moves, given piece locations and their move gen
	// function.
	allMoves := make([]move, 0, 4)

	for _, fromSquare := range read1Bits(piecesBB) {
		moves := read1Bits(moveGen(fromSquare, cb) & ^cb.BwPieces[cb.WToMove])
		for _, toSquare := range moves {
			if capturesBlocks == 0 {
				allMoves = append(allMoves, move{fromSquare, toSquare, piece, ""})
			} else {
				if 1<<toSquare&capturesBlocks != 0 {
					allMoves = append(allMoves, move{fromSquare, toSquare, piece, ""})
				}
			}
		}
	}

	return allMoves
}

func getCheckingSquares(cb *board.Board) (uint64, int) {
	// Return squares pieces other than the king can move to escape check and
	// the number of pieces checking the king.
	opponent := 1 ^ cb.WToMove
	attackerCount := 0

	// TODO: remove king from possible attackers
	knightAttackers := getKnightMoves(cb.KingSquare[cb.WToMove], cb) & cb.BwKnights[opponent]
	diagAttackers := getBishopMoves(cb.KingSquare[cb.WToMove], cb) &
		(cb.BwPawns[opponent] | cb.BwBishops[opponent] | cb.BwQueens[opponent] |
			cb.BwKing[opponent])
	orthogAttackers := getRookMoves(cb.KingSquare[cb.WToMove], cb) &
		(cb.BwRooks[opponent] | cb.BwQueens[opponent] | cb.BwKing[opponent])

		// TODO: Remove temporary sanity checks.
	if cb.BwKing[opponent]&(diagAttackers|orthogAttackers) != 0 {
		cb.Print()
		panic("king is checking the other king")
	}
	if len(read1Bits(knightAttackers)) > 1 {
		cb.Print()
		panic(">1 knights are checking the king")
	}

	// There should be 0 or 1 attackers in each attack group.
	if knightAttackers != 0 {
		attackerCount += 1
	}

	panicMsgs := []string{">1 piece is checking king orthogonally",
		">1 piece is checking king diagonally"}
	attackers := []uint64{orthogAttackers, diagAttackers}

	// Add interposition squares if any exist.
	for i, attacker := range attackers {
		if attacker != 0 {
			attackerCount += 1
			attackerSquares := read1Bits(attacker)
			if len(attackerSquares) > 1 {
				if i == 1 {
					// TODO: Refactor excluding pawns not in range of the king
					for _, attackerSquare := range attackerSquares {
						diff := attackerSquare - cb.KingSquare[cb.WToMove]
						if cb.BwPawns[opponent]&1<<attackerSquare != 0 &&
							diff*diff != 81 && diff*diff != 49 {
							// "checking" piece is a far away pawn
							attackers[i] -= 1 << attackerSquare
						}
					}
				} else {
					panic(panicMsgs[i])
				}
			}
			dir := findDirection(cb.KingSquare[cb.WToMove], attackerSquares[0])
			attackers[i] = fillFromTo(cb.KingSquare[cb.WToMove], attackerSquares[0], dir)
		}
	}

	return knightAttackers | attackers[0] | attackers[1], attackerCount
}

func fillFromTo(from, to, direction int) uint64 {
	// Return a bitboard excluding the 'from' square and including the 'to' square.
	bb := uint64(0)
	for sq := from + direction; sq != to; sq += direction {
		bb += 1 << sq
	}
	bb += 1 << to

	return bb
}

func findDirection(from, to int) int {
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
