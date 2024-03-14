package engine

import (
	"engine2/board"
	"engine2/pieces"
	"fmt"
	"log"
	"math/bits"
)

type TtEntry struct {
	Hash uint64
	Move board.Move
	Eval int
	Age  uint8
	Node rune // 'p': principal variation node, 'a': all node, 'c': cut node
}

const MATE = 1 << 20

var tTable = make(map[uint64]TtEntry, 500_000)
var Negamax = negamax
var emptyMove = board.Move{}

func negamax(alpha, beta, depth int, cb *board.Board, orig_depth int, orig_age uint8, pvMoves ...board.Move) (int, board.Move) {
	if depth == 0 {
		return evaluate(cb), cb.PrevMove
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

	// PV move given from iter. deepening and the side to move is the original side to move
	if len(pvMoves) > 0 && pvMoves[orig_depth-depth] != emptyMove { //&& orig_depth % 2 == depth % 2 {
		fmt.Println("orig, depth:", orig_depth, depth)
		fmt.Println("pvMoves:", pvMoves)
		// When orig - depth == 1, a new pvMove needs to be found organically
		if orig_depth > 1 && depth != 1 {
			if orig_depth-depth >= len(pvMoves) {
				log.Fatalf("out of bounds error: len=%d, idx=%d", len(pvMoves), orig_depth-depth)
			}
			pvMove := pvMoves[orig_depth-depth]
			// Linear search confirms the move exists, remove eventually?
			foundPvMode := false
			for i, move := range moves {
				//if move == pvMove[orig_depth-depth] {
				if move == pvMove {
					foundPvMode = true
					moves[0], moves[i] = moves[i], moves[0]
					break
				}
			}
			if !foundPvMode {
				cb.Print()
				fmt.Println("white to move:", cb.WToMove)
				// was panicking b/c tried to add white PV move on black's turn
				//panic("invalid move in this position?")
				fmt.Println(moves)
				log.Fatalf("invalid move in this position? move=%v", pvMove)
			}

			// clear pvMove after using because it is only used in one branch of the tree
			pvMoves[orig_depth-depth] = emptyMove
		}
	}

	for _, move := range moves {
		if move == emptyMove {
			panic("empty move")
		}
		pieces.MovePiece(move, cb)
		// Check legality of pseudo-legal moves. King moves are strictly legal already
		if move.Piece == 'k' || cb.Kings[1^cb.WToMove]&pieces.GetAttackedSquares(cb) == 0 {
			if stored, ok := tTable[cb.Zobrist]; ok && stored.Hash == cb.Zobrist {
				board.RestorePosition(pos, cb)
				switch stored.Node {
				case 'c':
					return stored.Eval, stored.Move
				case 'a':
				case 'p':
					if stored.Eval >= beta {
						return beta, move
					} else if stored.Eval > alpha {
						alpha = stored.Eval
					}
				default:
					panic("invalid node type")
				}
				continue
			} else {
				score, _ = negamax(-1*beta, -1*alpha, depth-1, cb, orig_depth, orig_age, pvMoves...)
				score *= -1
			}

			if score >= beta {
				//tTable[cb.Zobrist] = TtEntry{Hash: cb.Zobrist, Eval: beta, Age: orig_age, Move: move, Node: 'c'}
				board.RestorePosition(pos, cb)
				return beta, move
			} else if score > alpha {
				alpha = score
				bestMove = move
				//tTable[cb.Zobrist] = TtEntry{Hash: cb.Zobrist, Eval: score, Age: orig_age, Move: bestMove, Node: 'p'}
				if len(pvMoves) != 0 {
					if orig_depth-depth >= len(pvMoves) {
						fmt.Println("pvMoves:", pvMoves)
						log.Fatalf("index error: len=%d, idx=%d", len(pvMoves), orig_depth-depth)
					}
					if bestMove != emptyMove {
						pvMoves[orig_depth-depth] = bestMove
						fmt.Printf("pvMoves[%d] = %v ***************\n", orig_depth-depth, bestMove)
					}
				}
			} else {
				//tTable[cb.Zobrist] = TtEntry{Hash: cb.Zobrist, Eval: score, Age: orig_age, Move: bestMove, Node: 'a'}
			}
		}
		board.RestorePosition(pos, cb)
	}

	/*
	   if orig_depth-depth >= len(pvMoves) {
	       log.Fatalf("index error: len=%d, idx=%d", len(pvMoves), orig_depth-depth)
	   }
	   if bestMove != emptyMove {
	       pvMoves[orig_depth-depth] = bestMove
	       fmt.Printf("pvMoves[%d] = %v ***************\n", orig_depth-depth, bestMove)
	   }
	*/
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

// Successively call negamax() with increasing depth. It is generally faster than
// one search to a given depth
var IterativeDeepening = iterativeDeepening

func iterativeDeepening(cb *board.Board, depth int) (int, board.Move) {
	var eval int
	var move board.Move
	pvMoves := make([]board.Move, depth)

	for ply := 1; ply <= depth; ply++ {
		fmt.Println("starting ID ply", ply)
		/*
		   if move.From == 0 && move.To == 0 {
		       eval, move = negamax(-(1 << 30), 1<<30, ply, cb, ply, cb.HalfMoves)
		   } else {
		*/
		eval, move = negamax(-(1 << 30), 1<<30, ply, cb, ply, cb.HalfMoves, pvMoves...)
		//}
		pvMoves[0] = move
		fmt.Println("new [0] pv:", move, "*************")
	}

	return eval, move
}
