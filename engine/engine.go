package engine

import (
	"engine2/board"
	"engine2/pieces"
	"fmt"
	"math/bits"
)

func negamax(alpha, beta, depth int, cb *board.Board) (int, board.Move) {
	if depth == 0 {
		return evaluate(cb), cb.PrevMove
	}
	var lastMove, bestMove board.Move
	var score int
	pos := board.StorePosition(cb)

	for _, move := range pieces.GetAllMoves(cb) {
		pieces.MovePiece(move, cb)
		// Check legality of pseudo-legal moves. King moves are strictly legal already
		if move.Piece == 'k' || cb.Kings[1^cb.WToMove]&pieces.GetAttackedSquares(cb) == 0 {
			score, lastMove = negamax(-1*beta, -1*alpha, depth-1, cb)
			score *= -1
			if score >= beta {
				board.RestorePosition(pos, cb)
				//fmt.Println("beta cut:", beta, lastMove)
				return beta, lastMove
			} else if score > alpha {
				alpha = score
				bestMove = lastMove
				//fmt.Println("alpha =", alpha, "bestMoveSoFar = ", bestMove)
			}
		}
		board.RestorePosition(pos, cb)
	}

	return alpha, bestMove
}

// Return position evaluation in decipawns (0.1 pawns)
func evaluate(cb *board.Board) int {
	// TODO: piece square tables, evaluation tapering (middle to endgame),
	//   king safety, rooks on (semi-)open files, bishop pair (>= 2),
	//   endgame rooks/queens on 7th rank, connected rooks,

	eval := 10 * (bits.OnesCount64(cb.Pawns[1]) - bits.OnesCount64(cb.Pawns[0]))
	eval += 30 * (bits.OnesCount64(cb.Knights[1]) - bits.OnesCount64(cb.Knights[0]))
	eval += 31 * (bits.OnesCount64(cb.Bishops[1]) - bits.OnesCount64(cb.Bishops[0]))
	eval += 50 * (bits.OnesCount64(cb.Rooks[1]) - bits.OnesCount64(cb.Rooks[0]))
	eval += 90 * (bits.OnesCount64(cb.Queens[1]) - bits.OnesCount64(cb.Queens[0]))

	// TODO: outpost squares? Tapering required
	// TODO: remove knight moves to squares attacked by enemy pawns
	moveCount := len(pieces.GetAllMoves(cb))
	if moveCount == 0 {
		fmt.Println("stalemate or checkmate")
	}
	cb.WToMove ^= 1
	oppMoveCount := len(pieces.GetAllMoves(cb))
	cb.WToMove ^= 1
	if cb.WToMove == 1 {
		eval += moveCount - oppMoveCount
	} else {
		eval += oppMoveCount - moveCount
	}

	eval += evalPawns(cb)

	// Negamax requires eval respective to the color-to-move
	if cb.WToMove == 0 {
		eval *= -1
	}

	return eval
}

// Return evaluation of doubled, blocked, and isolated pawns.
func evalPawns(cb *board.Board) int {
	eval := 0
	var wPawnsInFile, bPawnsInFile [8]int

	// TODO: replace counting doubled/isolated pawns with an intersection of a
	// pre-calculated mask (masks[2][64]) and cb.Pawns[WToMove]
	// Doubled
	file := uint64(0x101010101010101)
	for i := 0; i < 8; i++ {
		wPawnsInFile[i] = bits.OnesCount64(file & cb.Pawns[1])
		if wPawnsInFile[i] > 1 {
			eval -= 5 * wPawnsInFile[i]
		}
		bPawnsInFile[i] = bits.OnesCount64(file & cb.Pawns[0])
		if bPawnsInFile[i] > 1 {
			eval += 5 * bPawnsInFile[i]
		}
		file = file << 1
	}

	// Isolated
	delta := [2]int{-5, 5}
	for i, colorPawns := range [2][8]int{wPawnsInFile, bPawnsInFile} {
		for j, file := range colorPawns {
			switch j {
			case 0:
				// If pawn(s) are in the A file and no friendly pawns are in the B file
				if file > 0 && colorPawns[1] == 0 {
					eval += delta[i]
				}
			case 7:
				if file > 0 && colorPawns[6] == 0 {
					eval += delta[i]
				}
			default:
				if file > 0 && colorPawns[j-1] == 0 && colorPawns[j+1] == 0 {
					eval += delta[i]
				}
			}
		}
	}

	// Blocked
	occupied := cb.Pieces[0] | cb.Pieces[1]
	for _, sq := range pieces.Read1BitsPawns(cb.Pawns[1]) {
		if uint64(1<<(sq+8))&occupied != 0 {
			eval -= 5
		}
	}
	for _, sq := range pieces.Read1BitsPawns(cb.Pawns[0]) {
		if uint64(1<<(sq-8))&occupied != 0 {
			eval += 5
		}
	}

	return eval
}
