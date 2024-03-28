// Search and evaluation
package engine

import (
	"bytes"
	"fmt"
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
	MATE          = 1 << 20
	ORIG_HASH_CAP = 1 << 20
	MAX_PHASE     = 1024

	PV_NODE  = uint8(0)
	ALL_NODE = uint8(1)
	CUT_NODE = uint8(2)
)

var cacheHits int
var cacheCollisions int
var negamaxCalls int
var evalCalls int
var quiesceCalls int

var tTable = make(map[uint64]TtEntry, ORIG_HASH_CAP)
var Negamax = negamax
var emptyMove = board.Move{}

func negamax(alpha, beta, depth int, cb *board.Board, orig_depth int, orig_age uint8, parentPartialPV *[]board.Move, completePV *pvLine) (int, board.Move) {
	negamaxCalls++
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
			if stored, ok := tTable[cb.Zobrist]; ok {
				// If no pv nodes are stored, is it ok to always used cached
				// nodes regardless of relative depths?
				if stored.Hash == cb.Zobrist && stored.Depth >= uint8(depth) {
					cacheHits++
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
				} else {
					cacheCollisions++
					delete(tTable, cb.Zobrist)
				}
			}
			score, _ = negamax(-1*beta, -1*alpha, depth-1, cb, orig_depth, orig_age, &line, completePV)
			score *= -1

			if score >= beta {
				tTable[cb.Zobrist] = TtEntry{Hash: cb.Zobrist, Eval: beta, Age: orig_age, Move: move, NodeType: CUT_NODE, Depth: uint8(depth)}
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
				tTable[cb.Zobrist] = TtEntry{Hash: cb.Zobrist, Eval: score, Age: orig_age, Move: bestMove, NodeType: ALL_NODE, Depth: uint8(depth)}
			}
		}
		board.RestorePosition(pos, cb)
	}

	return alpha, bestMove
}

// Return position evaluation in centipawns (0.01 pawns)
func evaluate(cb *board.Board) int {
	evalCalls++
	// TODO: king safety, rooks on (semi-)open files, bishop pair (>= 2),
	//   endgame rooks/queens on 7th rank, connected rooks,

	// Tapered piece-square tables (PST) for everything except pawns
	eval, egPhase, pieceCounts := evalPieceSquareTables(cb)

	// Material
	eval += 300 * (pieceCounts[1][0] - pieceCounts[0][0])
	eval += 310 * (pieceCounts[1][1] - pieceCounts[0][1])
	eval += 500 * (pieceCounts[1][2] - pieceCounts[0][2])
	eval += 900 * (pieceCounts[1][3] - pieceCounts[0][3])

	// TODO: outpost squares? Tapering required
	// TODO: remove knight moves to squares attacked by enemy pawns

	mgPhase := MAX_PHASE - egPhase
	eval += evalPawns(cb, mgPhase, egPhase) // material, structure, and pawn PST
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
func evalPawns(cb *board.Board, mgPhase, egPhase int) int {
	eval := 0
	pawnsInFile := [2][8]int{} // first index is [black, white]
	wPawnCount := 0
	bPawnCount := 0

	// Blocked pawns and tapered pawn piece-square table
	occupied := cb.Pieces[0] | cb.Pieces[1]
	wPawns := cb.Pawns[1]
	square := 0
	for wPawns > 0 {
		square = bits.TrailingZeros64(wPawns)
		pawnsInFile[1][square%8] += 1
		if uint64(1<<(square+8))&occupied != 0 {
			eval -= 50
		}
		wPawnCount += 1
		eval += (mgPhase*mg_tables[0][square^56] + egPhase*eg_tables[0][square^56]) / MAX_PHASE
		wPawns &= wPawns - 1
	}
	bPawns := cb.Pawns[0]
	for bPawns > 0 {
		square = bits.TrailingZeros64(bPawns)
		pawnsInFile[0][square%8] += 1
		if uint64(1<<(square-8))&occupied != 0 {
			eval += 50
		}
		bPawnCount += 1
		eval -= (mgPhase*mg_tables[0][square] + egPhase*eg_tables[0][square]) / MAX_PHASE
		bPawns &= bPawns - 1
	}

	// Material
	eval += 100 * (wPawnCount - bPawnCount)

	// Doubled
	for i := 0; i < 8; i++ {
		if pawnsInFile[1][i] > 1 {
			eval -= 50 * pawnsInFile[1][i]
		}
		if pawnsInFile[0][i] > 1 {
			eval += 50 * pawnsInFile[0][i]
		}
	}

	// Isolated
	delta := [2]int{50, -50}
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
		mobilityEval += 10 * (moveCount - oppMoveCount)
	} else {
		mobilityEval += 10 * (oppMoveCount - moveCount)
	}
	return mobilityEval
}

type pvLine struct {
	moves       []board.Move
	alreadyUsed []bool
}

// Successively call negamax() with increasing depth. It is generally faster than
// one search to a given depth
func IterativeDeepening(cb *board.Board, depth int, stop ...chan bool) (int, board.Move) {
	var eval int
	var move board.Move
	line := make([]board.Move, 0)
	completePVLine := pvLine{}
	completePVLine.alreadyUsed = make([]bool, depth)

PlyLoop:
	for ply := 1; ply <= depth; ply++ {
		eval, move = negamax(-(1 << 30), 1<<30, ply, cb, ply, cb.HalfMoves, &line, &completePVLine)
		completePVLine.moves = line
		for i := range completePVLine.alreadyUsed {
			completePVLine.alreadyUsed[i] = false
		}

		// UCI stdout. TODO: use ticker to reduce prints, if needed
		// TODO: finish. add the important spec fields
		fmt.Printf("info depth %d", ply)
		if eval != MATE && eval != -MATE {
			fmt.Printf(" score cp %d", eval)
		} else {
			fmt.Printf(" score mate %d", len(completePVLine.moves))
		}
		fmt.Printf(" hashfull %d", len(tTable)*1000/ORIG_HASH_CAP)
		fmt.Printf(" pv")
		algebraicMoves := convertMovesToLongAlgebraic(completePVLine.moves)
		for _, algebraicMove := range algebraicMoves {
			fmt.Printf(" %s", algebraicMove)
		}
		fmt.Println()

		// UCI stop. TODO: stop inside the iterDeep recursion tree while keeping the PV moves
		if len(stop) == 1 {
			select {
			case <-stop[0]:
				break PlyLoop
			default:
			}
		}
	}
	bestmove := convertMovesToLongAlgebraic([]board.Move{move})[0]
	fmt.Println("bestmove", bestmove)

	cleanTranspositionTable(cb.HalfMoves)

	return eval, move
}

// Find an ideal, stable position with no critical captures or exchanges
func quiesce(alpha, beta int, cb *board.Board) int {
	quiesceCalls++
	score := evaluate(cb)
	if score >= beta {
		return beta
	} else if score > alpha {
		alpha = score
	}

	marginOfError := 200 // centipawns
	pieceValues := [5]int{100, 300, 310, 500, 900}
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

// Remove cached nodes which were not just calculated
func cleanTranspositionTable(currentHalfMoveAge uint8) {
	if len(tTable) > ORIG_HASH_CAP/5*4 {
		for key, stored := range tTable {
			if stored.Age != currentHalfMoveAge {
				delete(tTable, key)
			}
		}
	}
}

func convertMovesToLongAlgebraic(moves []board.Move) []string {
	algMoves := make([]string, len(moves))
	chars := make([]byte, 0, 5)

	for i, move := range moves {
		chars = append(chars, (byte(move.From)%8)+'a')
		chars = append(chars, (byte(move.From)/8)+'1')
		chars = append(chars, (byte(move.To)%8)+'a')
		chars = append(chars, (byte(move.To)/8)+'1')

		if move.PromoteTo != pieces.NO_PIECE {
			switch move.PromoteTo {
			case pieces.KNIGHT:
				chars = append(chars, 'n')
			case pieces.BISHOP:
				chars = append(chars, 'b')
			case pieces.ROOK:
				chars = append(chars, 'r')
			case pieces.QUEEN:
				chars = append(chars, 'q')
			}
		}
		chars = bytes.TrimLeft(chars, "\x00")
		algMoves[i] = string(chars)
		clear(chars)
	}

	return algMoves
}

// Find the piece-square table evaluation of the position. Return:
//   - evaluation of non-pawn pieces according to the piece-square tables
//   - current phase of the game, from 0 (opening) to 100 (end game)
//   - count of each piece on the board [black, white][n, b, r, q]
func evalPieceSquareTables(cb *board.Board) (int, int, [2][4]int) {
	// Find piece locations, exlucing pawns
	pieceSquares := [2][4][]int{}
	allPieces := [2][4]uint64{
		{cb.Knights[0], cb.Bishops[0], cb.Rooks[0], cb.Queens[0]},
		{cb.Knights[1], cb.Bishops[1], cb.Rooks[1], cb.Queens[1]},
	}
	for color := range allPieces {
		for piece := range allPieces[color] {
			squares := make([]int, 0, 2)
			for allPieces[color][piece] > 0 {
				squares = append(squares, bits.TrailingZeros64(allPieces[color][piece]))
				allPieces[color][piece] &= allPieces[color][piece] - 1
			}
			pieceSquares[color][piece] = squares
		}
	}

	// 24 = 1*initial knights + 1*initial bishops + 2*initial rooks + 4*initial queens
	maxPiecePhase := 24
	egPhase := maxPiecePhase

	egPhase -= len(pieceSquares[0][0]) + len(pieceSquares[1][0])
	egPhase -= len(pieceSquares[0][1]) + len(pieceSquares[1][1])
	egPhase -= 2 * (len(pieceSquares[0][2]) + len(pieceSquares[1][2]))
	egPhase -= 4 * (len(pieceSquares[0][3]) + len(pieceSquares[1][3]))

	egPhase = egPhase * MAX_PHASE / maxPiecePhase
	mgPhase := MAX_PHASE - egPhase

	// PST eval
	eval := (mgPhase*mg_tables[5][cb.KingSqs[1]^56] + egPhase*eg_tables[5][cb.KingSqs[1]^56]) / MAX_PHASE
	eval -= (mgPhase*mg_tables[5][cb.KingSqs[0]] + egPhase*eg_tables[5][cb.KingSqs[0]]) / MAX_PHASE

	// eg_tables: [p, n, b, r, q, k]
	// pieceSquares: [n, b, r, q]
	for piece := range pieceSquares[0] {
		pstIdx := piece + 1
		for square := range pieceSquares[0][piece] {
			eval -= (mgPhase*mg_tables[pstIdx][pieceSquares[0][piece][square]] +
				egPhase*eg_tables[pstIdx][pieceSquares[0][piece][square]]) / MAX_PHASE
		}
		for square := range pieceSquares[1][piece] {
			eval += (mgPhase*mg_tables[pstIdx][pieceSquares[1][piece][square]^56] +
				egPhase*eg_tables[pstIdx][pieceSquares[1][piece][square]^56]) / MAX_PHASE
		}
	}

	return eval, egPhase, [2][4]int{
		{len(pieceSquares[0][0]), len(pieceSquares[0][1]), len(pieceSquares[0][2]), len(pieceSquares[0][3])},
		{len(pieceSquares[1][0]), len(pieceSquares[1][1]), len(pieceSquares[1][2]), len(pieceSquares[1][3])},
	}
}

var pawnMG = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	98, 134, 61, 95, 68, 126, 34, -11,
	-6, 7, 26, 31, 65, 56, 25, -20,
	-14, 13, 6, 21, 23, 12, 17, -23,
	-27, -2, -5, 12, 17, 6, 10, -25,
	-26, -4, -4, -10, 3, 3, 33, -12,
	-35, -1, -20, -23, -15, 24, 38, -22,
	0, 0, 0, 0, 0, 0, 0, 0,
}
var pawnEG = [64]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	178, 173, 158, 134, 147, 132, 165, 187,
	94, 100, 85, 67, 56, 53, 82, 84,
	32, 24, 13, 5, -2, 4, 17, 17,
	13, 9, -3, -7, -7, -8, 3, -1,
	4, 7, -6, 1, 0, -5, -1, -8,
	13, 8, 8, 10, 13, 0, 2, -7,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var knightMG = [64]int{
	-167, -89, -34, -49, 61, -97, -15, -107,
	-73, -41, 72, 36, 23, 62, 7, -17,
	-47, 60, 37, 65, 84, 129, 73, 44,
	-9, 17, 19, 53, 37, 69, 18, 22,
	-13, 4, 16, 13, 28, 19, 21, -8,
	-23, -9, 12, 10, 19, 17, 25, -16,
	-29, -53, -12, -3, -1, 18, -14, -19,
	-105, -21, -58, -33, -17, -28, -19, -23,
}
var knightEG = [64]int{
	-58, -38, -13, -28, -31, -27, -63, -99,
	-25, -8, -25, -2, -9, -25, -24, -52,
	-24, -20, 10, 9, -1, -9, -19, -41,
	-17, 3, 22, 22, 22, 11, 8, -18,
	-18, -6, 16, 25, 16, 17, 4, -18,
	-23, -3, -1, 15, 10, -3, -20, -22,
	-42, -20, -10, -5, -2, -20, -23, -44,
	-29, -51, -23, -15, -22, -18, -50, -64,
}

var bishopMG = [64]int{
	-29, 4, -82, -37, -25, -42, 7, -8,
	-26, 16, -18, -13, 30, 59, 18, -47,
	-16, 37, 43, 40, 35, 50, 37, -2,
	-4, 5, 19, 50, 37, 37, 7, -2,
	-6, 13, 13, 26, 34, 12, 10, 4,
	0, 15, 15, 15, 14, 27, 18, 10,
	4, 15, 16, 0, 7, 21, 33, 1,
	-33, -3, -14, -21, -13, -12, -39, -21,
}
var bishopEG = [64]int{
	-14, -21, -11, -8, -7, -9, -17, -24,
	-8, -4, 7, -12, -3, -13, -4, -14,
	2, -8, 0, -1, -2, 6, 0, 4,
	-3, 9, 12, 9, 14, 10, 3, 2,
	-6, 3, 13, 19, 7, 10, -3, -9,
	-12, -3, 8, 10, 13, 3, -7, -15,
	-14, -18, -7, -1, 4, -9, -15, -27,
	-23, -9, -23, -5, -9, -16, -5, -17,
}

var rookMG = [64]int{
	32, 42, 32, 51, 63, 9, 31, 43,
	27, 32, 58, 62, 80, 67, 26, 44,
	-5, 19, 26, 36, 17, 45, 61, 16,
	-24, -11, 7, 26, 24, 35, -8, -20,
	-36, -26, -12, -1, 9, -7, 6, -23,
	-45, -25, -16, -17, 3, 0, -5, -33,
	-44, -16, -20, -9, -1, 11, -6, -71,
	-19, -13, 1, 17, 16, 7, -37, -26,
}
var rookEG = [64]int{
	13, 10, 18, 15, 12, 12, 8, 5,
	11, 13, 13, 11, -3, 3, 8, 3,
	7, 7, 7, 5, 4, -3, -5, -3,
	4, 3, 13, 1, 2, 1, -1, 2,
	3, 5, 8, 4, -5, -6, -8, -11,
	-4, 0, -5, -1, -7, -12, -8, -16,
	-6, -6, 0, 2, -9, -9, -11, -3,
	-9, 2, 3, -1, -5, -13, 4, -20,
}

var queenMG = [64]int{
	-28, 0, 29, 12, 59, 44, 43, 45,
	-24, -39, -5, 1, -16, 57, 28, 54,
	-13, -17, 7, 8, 29, 56, 47, 57,
	-27, -27, -16, -16, -1, 17, -2, 1,
	-9, -26, -9, -10, -2, -4, 3, -3,
	-14, 2, -11, -2, -5, 2, 14, 5,
	-35, -8, 11, 2, 8, 15, -3, 1,
	-1, -18, -9, 10, -15, -25, -31, -50,
}
var queenEG = [64]int{
	-9, 22, 22, 27, 27, 19, 10, 20,
	-17, 20, 32, 41, 58, 25, 30, 0,
	-20, 6, 9, 49, 47, 35, 19, 9,
	3, 22, 24, 45, 57, 40, 57, 36,
	-18, 28, 19, 47, 31, 34, 39, 23,
	-16, -27, 15, 6, 9, 17, 10, 5,
	-22, -23, -30, -16, -16, -23, -36, -32,
	-33, -28, -22, -43, -5, -32, -20, -41,
}

var kingMG = [64]int{
	-65, 23, 16, -15, -56, -34, 2, 13,
	29, -1, -20, -7, -8, -4, -38, -29,
	-9, 24, 2, -16, -20, 6, 22, -22,
	-17, -20, -12, -27, -30, -25, -14, -36,
	-49, -1, -27, -39, -46, -44, -33, -51,
	-14, -14, -22, -46, -44, -30, -15, -27,
	1, 7, -8, -64, -43, -16, 9, 8,
	-15, 36, 12, -54, 8, -28, 24, 14,
}
var kingEG = [64]int{
	-74, -35, -18, -18, -11, 15, 4, -17,
	-12, 17, 14, 17, 17, 38, 23, 11,
	10, 17, 23, 15, 20, 45, 44, 13,
	-8, 22, 24, 27, 26, 33, 26, 3,
	-18, -4, 21, 24, 27, 23, 9, -11,
	-19, -3, 11, 21, 23, 16, 7, -9,
	-27, -11, 4, 13, 14, 4, -5, -17,
	-53, -34, -21, -11, -28, -14, -24, -43,
}

var mg_tables = [6][64]int{pawnMG, knightMG, bishopMG, rookMG, queenMG, kingMG}
var eg_tables = [6][64]int{pawnEG, knightEG, bishopEG, rookEG, queenEG, kingEG}
