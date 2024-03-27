// Move generation
package pieces

import (
	"bufio"
	"fmt"
	"github.com/j1642/chess-engine-2/board"
	"github.com/j1642/chess-engine-2/moves"
	"log"
	"math/bits"
	"os"
	"strings"
)

/*
bb = bitboard, cb = chessboard
Magic numbers 0, ..., 63 and 1<<0, ..., 1<<63 are squares of the chessboard.
*/

const (
	PAWN     = uint8(0)
	KNIGHT   = uint8(1)
	BISHOP   = uint8(2)
	ROOK     = uint8(3)
	QUEEN    = uint8(4)
	KING     = uint8(5)
	NO_PIECE = uint8(9)
)

func MovePiece(move board.Move, cb *board.Board) {
	fromBB := uint64(1 << move.From)
	toBB := uint64(1 << move.To)
	if cb.EpSquare != 100 {
		cb.Zobrist ^= board.ZobristKeys.EpFile[cb.EpSquare%8]
	}

	if toBB&(cb.Pieces[1^cb.WToMove]^cb.Kings[1^cb.WToMove]) != 0 {
		capturePiece(toBB, move.To, cb)
	}

	cb.Pieces[cb.WToMove] ^= fromBB + toBB

	switch move.Piece {
	case PAWN:
		cb.Pawns[cb.WToMove] ^= fromBB + toBB
		if move.To-move.From == 16 || move.To-move.From == -16 {
			cb.EpSquare = (move.To + move.From) / 2
			cb.Zobrist ^= board.ZobristKeys.EpFile[cb.EpSquare%8]
			cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][0][move.To]
		} else if move.To < 8 || move.To > 55 {
			promotePawn(toBB, move.To, cb, move.PromoteTo)
			cb.EpSquare = 100
		} else if move.To == cb.EpSquare {
			captureSq := move.To + 8
			if cb.WToMove == 1 {
				captureSq = move.To - 8
			}
			cb.Pawns[1^cb.WToMove] ^= uint64(1 << captureSq)
			cb.Pieces[1^cb.WToMove] ^= uint64(1 << captureSq)
			cb.EpSquare = 100
			cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][0][move.To]
		} else {
			cb.EpSquare = 100
			cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][0][move.To]
		}
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][0][move.From]
		cb.HalfMoves = 1
	case KNIGHT:
		cb.Knights[cb.WToMove] ^= fromBB + toBB
		cb.EpSquare = 100
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][1][move.From]
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][1][move.To]
	case BISHOP:
		cb.Bishops[cb.WToMove] ^= fromBB + toBB
		cb.EpSquare = 100
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][2][move.From]
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][2][move.To]
	case ROOK:
		cb.Rooks[cb.WToMove] ^= fromBB + toBB
		if move.From == 0 || move.From == 56 {
			if cb.CastleRights[cb.WToMove][0] == true {
				cb.Zobrist ^= board.ZobristKeys.Castle[cb.WToMove][0]
			}
			cb.CastleRights[cb.WToMove][0] = false
		} else if move.From == 7 || move.From == 63 {
			if cb.CastleRights[cb.WToMove][1] == true {
				cb.Zobrist ^= board.ZobristKeys.Castle[cb.WToMove][1]
			}
			cb.CastleRights[cb.WToMove][1] = false
		}
		cb.EpSquare = 100
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][3][move.From]
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][3][move.To]
	case QUEEN:
		cb.Queens[cb.WToMove] ^= fromBB + toBB
		cb.EpSquare = 100
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][4][move.From]
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][4][move.To]
	case KING:
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
		if cb.CastleRights[cb.WToMove][0] == true {
			cb.Zobrist ^= board.ZobristKeys.Castle[cb.WToMove][0]
		}
		if cb.CastleRights[cb.WToMove][1] == true {
			cb.Zobrist ^= board.ZobristKeys.Castle[cb.WToMove][1]
		}
		cb.CastleRights[cb.WToMove][0] = false
		cb.CastleRights[cb.WToMove][1] = false
		cb.EpSquare = 100
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][5][move.From]
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][5][move.To]
	case NO_PIECE:
		break
	default:
		panic("empty or invalid piece type")
	}

	cb.PrevMove = move
	cb.WToMove ^= 1
	cb.Zobrist ^= board.ZobristKeys.BToMove
	cb.HalfMoves += 1
}

func capturePiece(squareBB uint64, square int8, cb *board.Board) {
	opponent := 1 ^ cb.WToMove
	cb.Pieces[opponent] ^= squareBB
	cb.HalfMoves = 1

	switch {
	case squareBB&cb.Pawns[opponent] != 0:
		cb.Pawns[opponent] ^= squareBB
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[opponent][0][square]
	case squareBB&cb.Knights[opponent] != 0:
		cb.Knights[opponent] ^= squareBB
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[opponent][1][square]
	case squareBB&cb.Bishops[opponent] != 0:
		cb.Bishops[opponent] ^= squareBB
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[opponent][2][square]
	case squareBB&cb.Rooks[opponent] != 0:
		// int type mixing here seems ok based on investigation
		if opponent == 0 && squareBB == 1<<56 {
			cb.CastleRights[opponent][0] = false
			cb.Zobrist ^= board.ZobristKeys.Castle[opponent][0]
		} else if opponent == 0 && squareBB == 1<<63 {
			cb.CastleRights[opponent][1] = false
			cb.Zobrist ^= board.ZobristKeys.Castle[opponent][1]
		} else if opponent == 1 && squareBB == 0 {
			cb.CastleRights[opponent][0] = false
			cb.Zobrist ^= board.ZobristKeys.Castle[opponent][0]
		} else if opponent == 1 && squareBB == 1<<7 {
			cb.CastleRights[opponent][1] = false
			cb.Zobrist ^= board.ZobristKeys.Castle[opponent][1]
		}
		cb.Rooks[opponent] ^= squareBB
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[opponent][3][square]
	case squareBB&cb.Queens[opponent] != 0:
		cb.Queens[opponent] ^= squareBB
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[opponent][4][square]
	default:
		panic("no captured piece bitboard matches")
	}
}

func promotePawn(toBB uint64, square int8, cb *board.Board, promoteTo ...uint8) {
	if len(promoteTo) == 1 {
		switch {
		case promoteTo[0] == QUEEN:
			cb.Queens[cb.WToMove] ^= toBB
			cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][4][square]
		case promoteTo[0] == KNIGHT:
			cb.Knights[cb.WToMove] ^= toBB
			cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][1][square]
		case promoteTo[0] == BISHOP:
			cb.Bishops[cb.WToMove] ^= toBB
			cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][2][square]
		case promoteTo[0] == ROOK:
			cb.Rooks[cb.WToMove] ^= toBB
			cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][3][square]
		default:
			panic("invalid promoteTo")
		}
	} else {
		fmt.Print("promote pawn to N, B, R, or Q: ")
		userPromote := getUserInput()

		if userPromote == QUEEN || userPromote == KNIGHT || userPromote == BISHOP ||
			userPromote == ROOK {
			promotePawn(toBB, square, cb, userPromote)
		} else {
			fmt.Println("invalid promotion type, try again")
			promotePawn(toBB, square, cb)
		}
	}

	cb.Pawns[cb.WToMove] ^= toBB
}

func getUserInput() uint8 {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	err := scanner.Err()
	if err != nil {
		log.Println("failed to get input:", err)
		return getUserInput()
	}

	var piece uint8
	switch strings.ToLower(scanner.Text())[0] {
	case byte('n'):
		piece = KNIGHT
	case byte('b'):
		piece = BISHOP
	case byte('r'):
		piece = ROOK
	case byte('q'):
		piece = QUEEN
	default:
		return getUserInput()
	}
	return piece
}

// Use for user-submitted moves only?
// Checks for blocking pieces and disallows captures of friendly pieces.
// Does not consider check, pins, or legality of a pawn movement direction.
func IsValidMove(from, to int8, pieceType uint8, cb *board.Board) bool {
	if from < 0 || from > 63 || to < 0 || to > 63 || to == from {
		return false
	}
	if pieceType == NO_PIECE {
		log.Println("isValidMove: NO_PIECE has no valid moves")
	}

	toBB := uint64(1 << to)
	// Friendly piece collision
	if toBB&cb.Pieces[cb.WToMove] != 0 {
		return false
	}

	diff := to - from
	// to == from already excluded, no 0 move bugs from pawnDirections.
	pawnDirections := [2][4]int8{{-7, -8, -9, -16},
		{7, 8, 9, 16},
	}

	switch pieceType {
	case PAWN:
		if !board.ContainsN(diff, pawnDirections[cb.WToMove]) {
			return false
		}
	case KNIGHT:
		if toBB&moves.Knight[from] == 0 {
			return false
		}
	case BISHOP:
		if toBB&lookupBishopMoves(from, cb) == 0 {
			return false
		}
	case ROOK:
		if toBB&lookupRookMoves(from, cb) == 0 {
			return false
		}
	case QUEEN:
		if toBB&(lookupRookMoves(from, cb)|lookupBishopMoves(from, cb)) == 0 {
			return false
		}
	case KING:
		cb.Pieces[cb.WToMove] ^= uint64(1 << cb.KingSqs[cb.WToMove])
		cb.WToMove ^= 1
		attkSquares := GetAttackedSquares(cb)
		cb.WToMove ^= 1
		cb.Pieces[cb.WToMove] ^= uint64(1 << cb.KingSqs[cb.WToMove])
		if toBB&GetKingMoves(from, attkSquares, cb) == 0 {
			return false
		}
	default:
		// pieceType is not valid
		return false
	}

	return true
}

func lookupRookMoves(square int8, cb *board.Board) uint64 {
	occupied := cb.Pieces[0] | cb.Pieces[1]
	masked_blockers := moves.RookRelevantOccs[square] & occupied
	idx := (masked_blockers * moves.RookMagics[square]) >> (64 - moves.RookOneBitCounts[square])
	// Do not exclude piece protection (no `& ^cb.Pieces[cb.WToMove]`)
	return moves.RookMagicAttacks[square][idx]
}

func GetPawnMoves(square int8, cb *board.Board) uint64 {
	opponent := 1 ^ cb.WToMove

	if square < 8 || square > 55 {
		panic("pawns can't be on the first or last rank")
	}

	moves := moves.Pawn[cb.WToMove][square] & (cb.Pieces[opponent] | uint64(1<<cb.EpSquare))

	var dir, low, high int8
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

func getKnightMoves(square int8, cb *board.Board) uint64 {
	return moves.Knight[square]
}

func lookupBishopMoves(square int8, cb *board.Board) uint64 {
	occupied := cb.Pieces[0] | cb.Pieces[1]
	masked_blockers := moves.BishopRelevantOccs[square] & occupied
	idx := (masked_blockers * moves.BishopMagics[square]) >> (64 - moves.BishopOneBitCounts[square])
	// Do not exclude piece protection (no `& ^cb.Pieces[cb.WToMove]`)
	return moves.BishopMagicAttacks[square][idx]
}

func getQueenMoves(square int8, cb *board.Board) uint64 {
	return lookupRookMoves(square, cb) | lookupBishopMoves(square, cb)
}

// Return legal king moves.
func GetKingMoves(square int8, oppAttackedSquares uint64, cb *board.Board) uint64 {
	occupied := cb.Pieces[0] | cb.Pieces[1]
	moves := moves.King[square] & ^oppAttackedSquares & ^cb.Pieces[cb.WToMove]

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
	attackSquares := uint64(0)

	bb := cb.Pawns[cb.WToMove]
	for bb > 0 {
		attackSquares |= moves.Pawn[cb.WToMove][bits.TrailingZeros64(bb)]
		bb &= bb - 1
	}
	bb = cb.Knights[cb.WToMove]
	for bb > 0 {
		attackSquares |= moves.Knight[bits.TrailingZeros64(bb)]
		bb &= bb - 1
	}
	bb = cb.Bishops[cb.WToMove]
	for bb > 0 {
		attackSquares |= lookupBishopMoves(int8(bits.TrailingZeros64(bb)), cb)
		bb &= bb - 1
	}
	bb = cb.Rooks[cb.WToMove]
	for bb > 0 {
		attackSquares |= lookupRookMoves(int8(bits.TrailingZeros64(bb)), cb)
		bb &= bb - 1
	}
	bb = cb.Queens[cb.WToMove]
	for bb > 0 {
		attackSquares |= getQueenMoves(int8(bits.TrailingZeros64(bb)), cb)
		bb &= bb - 1
	}
	// Do not include castling.
	attackSquares |= moves.King[cb.KingSqs[cb.WToMove]]

	return attackSquares
}

type moveGenFunc func(int8, *board.Board) uint64

// Return slice of all pseudo-legal moves for color cb.WToMove, where any king
// moves are strictly legal. However, if the king is in check, only legal moves
// are returned
func GetAllMoves(cb *board.Board) []board.Move {
	cb.Pieces[cb.WToMove] ^= 1 << cb.KingSqs[cb.WToMove]
	cb.WToMove ^= 1
	attackedSquares := GetAttackedSquares(cb)
	cb.WToMove ^= 1
	cb.Pieces[cb.WToMove] ^= 1 << cb.KingSqs[cb.WToMove]

	var capturesBlks uint64
	var attackerCount int
	if cb.Kings[cb.WToMove]&attackedSquares != 0 {
		capturesBlks, attackerCount = GetCheckingSquares(cb)
	}

	// TODO: Trying to use a global allMoves did not work well
	allMoves := make([]board.Move, 0, 35)
	kingSq := cb.KingSqs[cb.WToMove]
	kingMovesBB := GetKingMoves(kingSq, attackedSquares, cb) & ^cb.Pieces[cb.WToMove]

	var toSq int8
	for kingMovesBB > 0 {
		toSq = int8(bits.TrailingZeros64(kingMovesBB))
		allMoves = append(allMoves, board.Move{From: kingSq, To: toSq, Piece: KING, PromoteTo: NO_PIECE})
		kingMovesBB &= kingMovesBB - 1
	}

	// If attackerCount > 1 and king has no moves, it is checkmate
	if attackerCount > 1 {
		return allMoves
	}

	pieces := [5]uint64{cb.Pawns[cb.WToMove], cb.Knights[cb.WToMove],
		cb.Bishops[cb.WToMove], cb.Rooks[cb.WToMove], cb.Queens[cb.WToMove],
	}
	moveFuncs := [5]moveGenFunc{GetPawnMoves, getKnightMoves, lookupBishopMoves,
		lookupRookMoves, getQueenMoves,
	}
	symbols := [5]uint8{PAWN, KNIGHT, BISHOP, ROOK, QUEEN}

	// 29% perft() speed up and -40% malloc from having this loop in this function
	var fromSq int8
	for i, pieceBB := range pieces {
		for pieceBB > 0 {
			fromSq = int8(bits.TrailingZeros64(pieceBB))
			pieceBB &= pieceBB - 1

			movesBB := moveFuncs[i](fromSq, cb) & ^cb.Pieces[cb.WToMove]
			for movesBB > 0 {
				toSq = int8(bits.TrailingZeros64(movesBB))
				movesBB &= movesBB - 1

				if capturesBlks == 0 || uint64(1<<toSq)&capturesBlks != 0 {
					if i != 0 || (7 < toSq && toSq < 56) {
						allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: symbols[i], PromoteTo: NO_PIECE})
					} else {
						allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: symbols[i], PromoteTo: KNIGHT})
						allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: symbols[i], PromoteTo: BISHOP})
						allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: symbols[i], PromoteTo: ROOK})
						allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: symbols[i], PromoteTo: QUEEN})
					}
				}
			}
		}
	}

	return allMoves
}

// Return the set of squares of pieces checking the king and interposition
// squares, and the number of checking pieces.
func GetCheckingSquares(cb *board.Board) (uint64, int) {
	opponent := 1 ^ cb.WToMove
	attackerCount := 0

	kSquare := cb.KingSqs[cb.WToMove]
	pAttackers := moves.Pawn[cb.WToMove][kSquare] & cb.Pawns[opponent]
	knightAttackers := moves.Knight[kSquare] & cb.Knights[opponent]
	bqAttackers := lookupBishopMoves(kSquare, cb) & (cb.Bishops[opponent] |
		cb.Queens[opponent])
	orthogAttackers := lookupRookMoves(cb.KingSqs[cb.WToMove], cb) &
		(cb.Rooks[opponent] | cb.Queens[opponent])

	if cb.Kings[opponent]&moves.King[cb.KingSqs[cb.WToMove]] != 0 {
		fmt.Println(cb.KingSqs)
		cb.Print()
		panic("king is checking the other king")
	}
	if bits.OnesCount64(knightAttackers) > 1 {
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

	panicMsgs := [2]string{">1 piece is checking king orthogonally",
		">1 piece is checking king diagonally"}
	attackers := [2]uint64{orthogAttackers, bqAttackers}

	// Add interposition squares if any exist.
	for i, attacker := range attackers {
		if attacker != 0 {
			attackerSquares := read1Bits(attacker)
			attackerCount += len(attackerSquares)
			if len(attackerSquares) > 1 {
				if i == 0 && (cb.PrevMove.PromoteTo == ROOK || cb.PrevMove.PromoteTo == QUEEN) {
					// Two pieces can orthogonally check a king if one was just promoted
					// from a pawn, with the other piece previously protecting the pawn
				} else {
					panic(panicMsgs[i])
				}
			}
			dir := findDirection(cb.KingSqs[cb.WToMove], attackerSquares[0])
			attackers[i] = fillFromTo(cb.KingSqs[cb.WToMove], attackerSquares[0], dir)
		}
	}

	return pAttackers | knightAttackers | attackers[0] | attackers[1], attackerCount
}

// Return a bitboard of squares between `from and `to`, excluding `from`
// and including `to`
func fillFromTo(from, to, direction int8) uint64 {
	bb := uint64(0)
	for sq := from + direction; sq != to; sq += direction {
		bb += 1 << sq
	}
	bb += 1 << to

	return bb
}

// Return the direction from one square to another. Assumes (from, to) is an
// orthogonal or diagonal move
func findDirection(from, to int8) int8 {
	var dir int8
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

func read1Bits(bb uint64) []int8 {
	// Using TrailingZeros64() seems as fast as bitshifting right while bb>0.
	squares := make([]int8, 0, 4)
	for bb > 0 {
		squares = append(squares, int8(bits.TrailingZeros64(bb)))
		bb &= bb - 1
	}
	return squares
}
