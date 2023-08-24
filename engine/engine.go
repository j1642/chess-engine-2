package engine

import (
	"engine2/board"
	"engine2/pieces"
)

func negamax(alpha, beta, depth int, cb *board.Board) int {
	if depth == 0 {
		return evaluate(cb)
	}
	var attackedSquares uint64
	pos := board.StorePosition(cb)

	for _, move := range pieces.GetAllMoves(cb) {
		pieces.MovePiece(move, cb)
		if move.Piece == "k" {
			attackedSquares = 0
		} else {
			attackedSquares = pieces.GetAttackedSquares(cb)
		}

		if cb.Kings[1^cb.WToMove]&attackedSquares == 0 {
			score := -1 * negamax(-beta, -alpha, depth-1, cb)
			if score >= beta {
				board.RestorePosition(pos, cb)
				return beta
			} else if score > alpha {
				alpha = score
			}
		}
		board.RestorePosition(pos, cb)
	}

	return alpha
}

func evaluate(cb *board.Board) int {
	eval := 10 * (count1Bits(cb.Pawns[1]) - count1Bits(cb.Pawns[0]))
	eval += 30 * (count1Bits(cb.Knights[1]) - count1Bits(cb.Knights[0]))
	eval += 31 * (count1Bits(cb.Bishops[1]) - count1Bits(cb.Bishops[0]))
	eval += 50 * (count1Bits(cb.Rooks[1]) - count1Bits(cb.Rooks[0]))
	eval += 90 * (count1Bits(cb.Queens[1]) - count1Bits(cb.Queens[0]))

	moveCount := len(pieces.GetAllMoves(cb))
	cb.WToMove ^= 1
	oppMoveCount := len(pieces.GetAllMoves(cb))
	cb.WToMove ^= 1
	if cb.WToMove == 1 {
		eval += moveCount - oppMoveCount
	} else {
		eval += oppMoveCount - moveCount
	}

	var wDoubledPawns, bDoubledPawns int
	file := uint64(0x101010101010101)

	for i := 0; i < 8; i++ {
		wDoubledPawns = count1Bits(file & cb.Pawns[1])
		if wDoubledPawns > 1 {
			eval -= 5 * wDoubledPawns
		}

		bDoubledPawns = count1Bits(file & cb.Pawns[0])
		if bDoubledPawns > 1 {
			eval += 5 * bDoubledPawns
		}
		file = file << 1
	}

	return eval
}

func count1Bits(bb uint64) int {
	count := 0
	for bb > 0 {
		count += 1
		bb &= bb - 1
	}
	return count
}
