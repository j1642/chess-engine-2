package engine

import (
	"github.com/j1642/chess-engine-2/board"
	"github.com/j1642/chess-engine-2/pieces"
	"math/bits"
)

type TtEntry struct {
	Hash                 uint64
	Eval                 int
	Move                 board.Move
	NodeType, Age, Depth uint8
}

const (
	MATE = 1 << 20

	PV_NODE  = uint8(0)
	ALL_NODE = uint8(1)
	CUT_NODE = uint8(2)
)

var tTable = make(map[uint64]TtEntry, 1_000_000)
var Negamax = negamax
var emptyMove = board.Move{}

func negamax(alpha, beta, depth int, cb *board.Board, orig_depth int, orig_age uint8, parentPartialPV *[]board.Move, completePV *pvLine) (int, board.Move) {
	if depth == 0 {
		return quiesce(alpha, beta, cb), cb.PrevMove
	}
	var bestMove board.Move
	var score int
	pos := board.StorePosition(cb)

	moves := pieces.GetAllMoves(cb)
	if len(moves) == 0 {
		// End of branch when depth > 0, checkmate or stalemate
		score = -MATE
		// Negamax evaluations are relative to the side to move. Regardless of
		// the side to move, being in checkmate is bad, and is a negative score
		return score, cb.PrevMove
	}

	// if a PV move exists for this depth and it has not been used yet
	if len(completePV.moves) > 0 && orig_depth > 1 && depth > 1 && len(completePV.moves) >= orig_depth-depth && !(completePV.alreadyUsed)[orig_depth-depth] {
		pvMove := completePV.moves[orig_depth-depth]
		// Linear search confirms the move exists, remove eventually?
		foundPvMode := false
		for i, move := range moves {
			if move == pvMove {
				foundPvMode = true
				moves[0], moves[i] = moves[i], moves[0]
				completePV.alreadyUsed[orig_depth-depth] = true
				break
			}
		}
		if !foundPvMode {
			panic("invalid move in this position")
		}
	}

	line := make([]board.Move, 0)

	for _, move := range moves {
		if move == emptyMove {
			panic("cannot do an empty move")
		}
		pieces.MovePiece(move, cb)
		// Check legality of pseudo-legal moves. King moves are strictly legal already
		if move.Piece == pieces.KING || cb.Kings[1^cb.WToMove]&pieces.GetAttackedSquares(cb) == 0 {
			// TODO: add depth check, otherwise calculate new negamax at the new, deeper depth
			if stored, ok := tTable[cb.Zobrist]; ok && stored.Hash == cb.Zobrist && stored.Depth >= uint8(depth) {
				board.RestorePosition(pos, cb)
				switch stored.NodeType {
				case CUT_NODE:
					return stored.Eval, stored.Move
				case ALL_NODE:
				case PV_NODE:
					if stored.Eval >= beta {
						return beta, move
					} else if stored.Eval > alpha {
						alpha = stored.Eval
					}

					// PV block
					if len(*parentPartialPV) == 0 {
						*parentPartialPV = append(*parentPartialPV, move)
					} else {
						(*parentPartialPV)[0] = move
					}
					for i := range line {
						if len(*parentPartialPV) <= i+1 {
							*parentPartialPV = append(*parentPartialPV, line[i])
						} else {
							(*parentPartialPV)[1+i] = line[i]
						}
					}
				default:
					panic("invalid node type")
				}
				continue
			}
			score, _ = negamax(-1*beta, -1*alpha, depth-1, cb, orig_depth, orig_age, &line, completePV)
			score *= -1

			if score >= beta {
				//tTable[cb.Zobrist] = TtEntry{Hash: cb.Zobrist, Eval: beta, Age: orig_age, Move: move, NodeType: CUT_NODE, Depth: uint8(depth)}
				board.RestorePosition(pos, cb)
				return beta, move
			} else if score > alpha {
				alpha = score
				bestMove = move
				//tTable[cb.Zobrist] = TtEntry{Hash: cb.Zobrist, Eval: score, Age: orig_age, Move: bestMove, NodeType: PV_NODE, Depth: uint8(depth)}

				// PV block
				if len(*parentPartialPV) == 0 {
					*parentPartialPV = append(*parentPartialPV, move)
				} else {
					(*parentPartialPV)[0] = move
				}
				for i := range line {
					if len(*parentPartialPV) <= i+1 {
						*parentPartialPV = append(*parentPartialPV, line[i])
					} else {
						(*parentPartialPV)[1+i] = line[i]
					}
				}
			} else {
				//tTable[cb.Zobrist] = TtEntry{Hash: cb.Zobrist, Eval: score, Age: orig_age, Move: bestMove, NodeType: ALL_NODE, Depth: uint8(depth)}
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

	eval += evalPawns(cb)
	mobilityEval := evaluateMobility(cb)
	if mobilityEval == -MATE {
		// checkmate or stalemate
		return mobilityEval
	}
	eval += mobilityEval

	// Negamax requires eval respective to the color-to-move
	if cb.WToMove == 0 {
		eval *= -1
	}

	return eval
}

// Return evaluation of doubled, blocked, and isolated pawns.
func evalPawns(cb *board.Board) int {
	eval := 0
	pawnsInFile := [2][8]int{} // first index is [black, white]

	// Blocked
	occupied := cb.Pieces[0] | cb.Pieces[1]
	wPawns := cb.Pawns[1]
	square := 0
	for wPawns > 0 {
		square = bits.TrailingZeros64(wPawns)
		pawnsInFile[1][square%8] += 1
		if uint64(1<<(square+8))&occupied != 0 {
			eval -= 5
		}
		wPawns &= wPawns - 1
	}
	bPawns := cb.Pawns[0]
	for bPawns > 0 {
		square = bits.TrailingZeros64(bPawns)
		pawnsInFile[0][square%8] += 1
		if uint64(1<<(square-8))&occupied != 0 {
			eval += 5
		}
		bPawns &= bPawns - 1
	}

	// Doubled
	for i := 0; i < 8; i++ {
		if pawnsInFile[1][i] > 1 {
			eval -= 5 * pawnsInFile[1][i]
		}
		if pawnsInFile[0][i] > 1 {
			eval += 5 * pawnsInFile[0][i]
		}
	}

	// Isolated
	delta := [2]int{5, -5}
	for i := range pawnsInFile {
		for j := range pawnsInFile[i] {
			if pawnsInFile[i][j] > 0 {
				switch j {
				case 0:
					// If pawn(s) are in the A file and no friendly pawns are in the B file
					if pawnsInFile[i][1] == 0 {
						eval += delta[i]
					}
				case 7:
					if pawnsInFile[i][6] == 0 {
						eval += delta[i]
					}
				default:
					if pawnsInFile[i][j-1] == 0 && pawnsInFile[i][j+1] == 0 {
						eval += delta[i]
					}
				}
			}
		}
	}

	return eval
}

func evaluateMobility(cb *board.Board) int {
	cb.WToMove ^= 1
	oppMovesBB := pieces.GetAttackedSquares(cb)
	cb.WToMove ^= 1

	movesBB := pieces.GetAttackedSquares(cb)
	origMovesBB := movesBB
	// Include legal king moves and castling
	movesBB |= pieces.GetKingMoves(cb.KingSqs[cb.WToMove], oppMovesBB, cb)
	// Include pawn forward moves
	pawnsBB := cb.Pawns[cb.WToMove]
	for pawnsBB > 0 {
		movesBB |= pieces.GetPawnMoves(int8(bits.TrailingZeros64(pawnsBB)), cb)
		pawnsBB &= pawnsBB - 1
	}
	movesBB &= ^cb.Pieces[cb.WToMove]
	moveCount := bits.OnesCount64(movesBB)

	cb.WToMove ^= 1
	// Include legal king moves and castling
	oppMovesBB |= pieces.GetKingMoves(cb.KingSqs[cb.WToMove], origMovesBB, cb)
	// Include pawn forward moves
	oppPawnsBB := cb.Pawns[cb.WToMove]
	for oppPawnsBB > 0 {
		oppMovesBB |= pieces.GetPawnMoves(int8(bits.TrailingZeros64(oppPawnsBB)), cb)
		oppPawnsBB &= oppPawnsBB - 1
	}
	oppMovesBB &= ^cb.Pieces[cb.WToMove]
	oppMoveCount := bits.OnesCount64(oppMovesBB)
	cb.WToMove ^= 1

	if cb.Kings[cb.WToMove]&oppMovesBB != 0 {
		// Only use slow getAllMoves() when king is in check, returns legal moves
		moveCount = len(pieces.GetAllMoves(cb))
	}
	// Checkmate and stalemate checks for the side to move
	// BUG for stalemate: when king is in check, GetAllMoves() returns legal moves only.
	//   Otherwise, illegal pseudo-legal moves may be included, which need to be
	//   removed to detect stalemate
	if moveCount == 0 && bits.OnesCount64(cb.Pieces[cb.WToMove]) > 0 {
		if _, countChecks := pieces.GetCheckingSquares(cb); countChecks > 0 {
			// Mate is always bad for the side-to-move, so it is a negative eval
			return -MATE
		}
		// else stalemate
	}
	mobilityEval := 0
	if cb.WToMove == 1 {
		mobilityEval += moveCount - oppMoveCount
	} else {
		mobilityEval += oppMoveCount - moveCount
	}
	return mobilityEval
}

type pvLine struct {
	moves       []board.Move
	alreadyUsed []bool
}

// Successively call negamax() with increasing depth. It is generally faster than
// one search to a given depth
func iterativeDeepening(cb *board.Board, depth int) (int, board.Move) {
	var eval int
	var move board.Move
	line := make([]board.Move, 0)
	completePVLine := pvLine{}
	completePVLine.alreadyUsed = make([]bool, depth)

	for ply := 1; ply <= depth; ply++ {
		eval, move = negamax(-(1 << 30), 1<<30, ply, cb, ply, cb.HalfMoves, &line, &completePVLine)
		completePVLine.moves = line
		for i := range completePVLine.alreadyUsed {
			completePVLine.alreadyUsed[i] = false
		}
	}

	return eval, move
}

// Find an ideal, stable position with no critical captures or exchanges
func quiesce(alpha, beta int, cb *board.Board) int {
	score := evaluate(cb)
	if score >= beta {
		return beta
	} else if score > alpha {
		alpha = score
	}

	marginOfError := 20 // decipawns
	pieceValues := [5]int{10, 30, 31, 50, 90}
	oppPieces := [5]uint64{cb.Pawns[cb.WToMove^1], cb.Knights[cb.WToMove^1],
		cb.Bishops[cb.WToMove^1], cb.Rooks[cb.WToMove^1], cb.Queens[cb.WToMove^1],
	}
	var capturedPieceValue int

	// Prune if gaining a queen doesn't raise alpha
	if alpha > score+pieceValues[4] {
		return alpha
	}

	moves := pieces.GetAllMoves(cb)
	// TODO: include other forcing moves like check and promotion?
	// Discard non-capture moves
	for i := 0; i < len(moves); i++ {
		moveToBB := uint64(1 << moves[i].To)
		// if the move is not a capture
		if (moveToBB & cb.Pieces[cb.WToMove^1]) == 0 {
			moves[i], moves[len(moves)-1] = moves[len(moves)-1], moves[i]
			moves = moves[:len(moves)-1]
			// Re-examine this index because it holds a different move now
			i--
			continue
		}
		// Delta pruning of captures that are unlikely to improve alpha
		for i := range oppPieces {
			if moveToBB&oppPieces[i] != 0 {
				capturedPieceValue = pieceValues[i]
				break
			}
		}
		if alpha > score+capturedPieceValue+marginOfError {
			// Prune move. Same as !isCapture block
			moves[i], moves[len(moves)-1] = moves[len(moves)-1], moves[i]
			moves = moves[:len(moves)-1]
			// Re-examine this index because it holds a different move now
			i--
		}
	}

	position := board.StorePosition(cb)
	for _, capture := range moves {
		pieces.MovePiece(capture, cb)

		if capture.Piece == pieces.KING || cb.Kings[1^cb.WToMove]&pieces.GetAttackedSquares(cb) == 0 {
			score = -quiesce(-beta, -alpha, cb)
		}
		board.RestorePosition(position, cb)

		if score >= beta {
			return beta
		} else if score > alpha {
			alpha = score
		}
	}

	return alpha
}
