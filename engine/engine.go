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
	return 2
}
