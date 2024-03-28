package board

import (
	"fmt"
	"math/bits"
	"math/rand/v2"
	"strings"
)

type Board struct {
	// TODO: move occupancies into one array? Possible memory speed boost
	// White to move. 1=true, 0=false. Use uint because bools cannot be xor'd
	WToMove uint

	Pieces  [2]uint64
	Pawns   [2]uint64
	Knights [2]uint64
	Bishops [2]uint64
	Rooks   [2]uint64
	Queens  [2]uint64
	Kings   [2]uint64

	KingSqs      [2]int8
	CastleRights [2][2]bool // [b, w][queenside, kingside]

	EpSquare  int8
	PrevMove  Move
	Zobrist   uint64
	HalfMoves uint8

	EvalMaterial   int
	PiecePhaseSum  int
	EvalMidGamePST int // piece-square tables (PST)
	EvalEndGamePST int
}

type Move struct {
	From, To         int8
	Piece, PromoteTo uint8
}

type Zobrist struct {
	ColorPieceSq [2][6][64]uint64
	BToMove      uint64
	Castle       [2][2]uint64
	EpFile       [8]uint64
}

var ZobristKeys Zobrist = buildZobristKeys()

func New() *Board {
	cb := &Board{
		WToMove: 1,

		Pieces:  [2]uint64{0xFFFF000000000000, 0xFFFF},
		Pawns:   [2]uint64{0xFF000000000000, 0xFF00},
		Knights: [2]uint64{1<<57 + 1<<62, 1<<1 + 1<<6},
		Bishops: [2]uint64{1<<58 + 1<<61, 1<<2 + 1<<5},
		Rooks:   [2]uint64{1<<56 + 1<<63, 1<<0 + 1<<7},
		Queens:  [2]uint64{1 << 59, 1 << 3},
		Kings:   [2]uint64{1 << 60, 1 << 4},

		KingSqs:      [2]int8{60, 4},
		CastleRights: [2][2]bool{{true, true}, {true, true}},

		EpSquare: 100,
		Zobrist:  0,

		PiecePhaseSum: 24,
	}
	cb.resetZobrist()
	cb.resetMidGameEndGamePST()

	return cb
}

func (cb *Board) resetZobrist() {
	zobrist := uint64(0)
	for color := range len(cb.Pawns) {
		pieceTypes := [6]uint64{cb.Pawns[color], cb.Knights[color], cb.Bishops[color],
			cb.Rooks[color], cb.Queens[color], cb.Kings[color],
		}
		for i, pieceBB := range pieceTypes {
			for pieceBB > 0 {
				zobrist ^= ZobristKeys.ColorPieceSq[color][i][bits.TrailingZeros64(pieceBB)]
				pieceBB &= pieceBB - 1
			}
		}
	}

	for i := range len(cb.CastleRights) {
		for j := range len(cb.CastleRights[0]) {
			if cb.CastleRights[i][j] {
				zobrist ^= ZobristKeys.Castle[i][j]
			}
		}
	}

	if cb.WToMove == 0 {
		zobrist ^= ZobristKeys.BToMove
	}

	// square 100 is an unused placeholder
	if cb.EpSquare != 100 {
		zobrist ^= ZobristKeys.EpFile[cb.EpSquare%8]
	}

	cb.Zobrist = zobrist
}

func buildZobristKeys() Zobrist {
	keys := Zobrist{}
	prng := rand.New(rand.NewPCG(17, 41))

	colorPieceSq := [2][6][64]uint64{}
	for color := 0; color < len(colorPieceSq); color++ {
		for pieceType := 0; pieceType < len(colorPieceSq[0]); pieceType++ {
			for square := 0; square < len(colorPieceSq[0][0]); square++ {
				colorPieceSq[color][pieceType][square] = prng.Uint64()
			}
		}
	}
	keys.ColorPieceSq = colorPieceSq
	keys.BToMove = prng.Uint64()
	for i := 0; i < len(keys.EpFile); i++ {
		keys.EpFile[i] = prng.Uint64()
	}
	for color := 0; color < len(keys.Castle); color++ {
		for qsideKside := 0; qsideKside < len(keys.Castle[0]); qsideKside++ {
			keys.Castle[color][qsideKside] = prng.Uint64()
		}
	}

	return keys
}

// Build a Board object from a Forsyth-Edwards notation (FEN) string
// example: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
func FromFen(fen string) (*Board, error) {
	// TODO: apply halfmove count, move count
	var color int
	cb := &Board{}
	square := int8(56)
	firstSpace := strings.IndexByte(fen, ' ')
	secondSpace := strings.IndexByte(fen[firstSpace+1:], ' ')
	pieceValues := [5]int{100, 300, 310, 500, 900} // material

	if firstSpace == -1 || secondSpace != 1 {
		return cb, fmt.Errorf("invalid FEN string")
	}
	slashCount := strings.Count(fen, "/")
	if slashCount != 7 {
		return cb, fmt.Errorf("invalid FEN slash count. want=7, got=%d", slashCount)
	}

	squaresInRank := 0
	cb.PiecePhaseSum = 0

	for _, char := range fen[:firstSpace] {
		if 'A' <= char && char <= 'Z' {
			color = 1
			squaresInRank += 1
		} else if 'a' <= char && char <= 'z' {
			color = 0
			squaresInRank += 1
		} else {
			color = 100 // placeholder value
		}

		switch {
		case '1' <= char && char <= '8':
			// Negate the "square += 1" at the end of the loop
			square += int8(char-'0') - 1
			squaresInRank += int(char - '0')
		case char == '/':
			// Negate the "square += 1" at the end of the loop
			square -= 17
			if squaresInRank != 8 {
				return cb, fmt.Errorf("invalid FEN: %d squares in rank", squaresInRank)
			}
			squaresInRank = 0
		case char == 'p' || char == 'P':
			cb.Pawns[color] += 1 << square
			if color == 1 {
				cb.EvalMaterial += pieceValues[0]
			} else {
				cb.EvalMaterial -= pieceValues[0]
			}
		case char == 'n' || char == 'N':
			cb.Knights[color] += 1 << square
			cb.PiecePhaseSum += 1
			if color == 1 {
				cb.EvalMaterial += pieceValues[1]
			} else {
				cb.EvalMaterial -= pieceValues[1]
			}
		case char == 'b' || char == 'B':
			cb.Bishops[color] += 1 << square
			cb.PiecePhaseSum += 1
			if color == 1 {
				cb.EvalMaterial += pieceValues[2]
			} else {
				cb.EvalMaterial -= pieceValues[2]
			}
		case char == 'r' || char == 'R':
			cb.Rooks[color] += 1 << square
			cb.PiecePhaseSum += 2
			if color == 1 {
				cb.EvalMaterial += pieceValues[3]
			} else {
				cb.EvalMaterial -= pieceValues[3]
			}
		case char == 'q' || char == 'Q':
			cb.Queens[color] += 1 << square
			cb.PiecePhaseSum += 4
			if color == 1 {
				cb.EvalMaterial += pieceValues[4]
			} else {
				cb.EvalMaterial -= pieceValues[4]
			}
		case char == 'k' || char == 'K':
			cb.Kings[color] += 1 << square
			cb.KingSqs[color] = square
		}

		square += 1
	}

	for i, char := range fen[firstSpace:] {
		switch {
		case char == 'b':
			cb.WToMove = 0
		case char == 'w':
			cb.WToMove = 1
		case char == 'K':
			cb.CastleRights[1][1] = true
		case char == 'k':
			cb.CastleRights[0][1] = true
		case char == 'Q':
			cb.CastleRights[1][0] = true
		case char == 'q':
			cb.CastleRights[0][0] = true
			// An empty castling field is also represented by '-'. Any EP will later
			// overwrite this.
		case char == '-':
			cb.EpSquare = 100
		case 'a' <= char && char <= 'h':
			// rank 1: square=0+column, rank 2: square=8+column, ...
			rank := 8 * (int8(fen[i+firstSpace+1]-'0') - 1)
			cb.EpSquare = (int8(char - 'a')) + rank
		}
	}

	cb.Pieces[0] = cb.Pawns[0] | cb.Knights[0] | cb.Bishops[0] |
		cb.Rooks[0] | cb.Queens[0] | cb.Kings[0]
	cb.Pieces[1] = cb.Pawns[1] | cb.Knights[1] | cb.Bishops[1] |
		cb.Rooks[1] | cb.Queens[1] | cb.Kings[1]

	// 24 is the max sum of piece phase values, as seen in the starting position
	if cb.PiecePhaseSum < 0 {
		panic("piece phase < 0")
	} else if 24 < cb.PiecePhaseSum {
		panic("piece phase > 24")
	}

	cb.resetZobrist()
	cb.resetMidGameEndGamePST()

	return cb, nil
}

func GetFiles() [4][8]int8 {
	fileA, fileB, fileG, fileH := [8]int8{}, [8]int8{}, [8]int8{}, [8]int8{}

	for i := int8(0); i < 8; i++ {
		fileA[i] = i * 8
		fileB[i] = i*8 + 1
		fileG[i] = i*8 + 6
		fileH[i] = i*8 + 7
	}

	return [4][8]int8{fileA, fileB, fileG, fileH}
}

type IntArray interface {
	[8]int8 | [4]int8 | [3]int8
}

// Linear search for small arrays
func ContainsN[T IntArray](n int8, nums T) bool {
	for i := 0; i < len(nums); i++ {
		if n == nums[i] {
			return true
		}
	}
	return false
}

type Position struct {
	WToMove uint

	Pieces  [2]uint64
	Pawns   [2]uint64
	Knights [2]uint64
	Bishops [2]uint64
	Rooks   [2]uint64
	Queens  [2]uint64
	Kings   [2]uint64

	KingSqs      [2]int8
	CastleRights [2][2]bool

	EpSquare  int8
	PrevMove  Move
	Zobrist   uint64
	HalfMoves uint8

	EvalMaterial   int
	PiecePhaseSum  int
	EvalMidGamePST int
	EvalEndGamePST int
}

func StorePosition(cb *Board) *Position {
	return &Position{
		WToMove: cb.WToMove,
		Pieces:  cb.Pieces,
		Pawns:   cb.Pawns,
		Knights: cb.Knights,
		Bishops: cb.Bishops,
		Rooks:   cb.Rooks,
		Queens:  cb.Queens,
		Kings:   cb.Kings,

		KingSqs:      cb.KingSqs,
		CastleRights: cb.CastleRights,

		EpSquare:  cb.EpSquare,
		PrevMove:  cb.PrevMove,
		Zobrist:   cb.Zobrist,
		HalfMoves: cb.HalfMoves,

		EvalMaterial:   cb.EvalMaterial,
		PiecePhaseSum:  cb.PiecePhaseSum,
		EvalMidGamePST: cb.EvalMidGamePST,
		EvalEndGamePST: cb.EvalEndGamePST,
	}
}

func RestorePosition(pos *Position, cb *Board) {
	cb.WToMove = pos.WToMove
	cb.Pieces = pos.Pieces
	cb.Pawns = pos.Pawns
	cb.Knights = pos.Knights
	cb.Bishops = pos.Bishops
	cb.Rooks = pos.Rooks
	cb.Queens = pos.Queens
	cb.Kings = pos.Kings

	cb.KingSqs = pos.KingSqs
	cb.CastleRights = pos.CastleRights

	cb.EpSquare = pos.EpSquare
	cb.PrevMove = pos.PrevMove
	cb.Zobrist = pos.Zobrist
	cb.HalfMoves = pos.HalfMoves

	cb.EvalMaterial = pos.EvalMaterial
	cb.PiecePhaseSum = pos.PiecePhaseSum
	cb.EvalMidGamePST = pos.EvalMidGamePST
	cb.EvalEndGamePST = pos.EvalEndGamePST
}

func (cb *Board) Print() {
	// Possibly destructive to original cb, so print a copy
	squares := [64]string{}
	copied := StorePosition(cb)

	pieces := [6]uint64{
		copied.Pawns[0] + copied.Pawns[1],
		copied.Knights[0] + copied.Knights[1],
		copied.Bishops[0] + copied.Bishops[1],
		copied.Rooks[0] + copied.Rooks[1],
		copied.Queens[0] + copied.Queens[1],
		copied.Kings[0] + copied.Kings[1],
	}
	symbols := [6]string{"p", "n", "b", "r", "q", "k"}

	for i, piece := range pieces {
		for piece != 0 {
			squares[bits.TrailingZeros64(piece)] = symbols[i]
			piece &= piece - 1
		}
	}

	for i, symbol := range squares {
		if copied.Pieces[1]&uint64(1<<i) != 0 {
			squares[i] = strings.ToUpper(symbol)
		}
	}

	for i := 56; i != 7; i++ {
		if squares[i] == "" {
			fmt.Print("- ")
		} else {
			fmt.Printf("%s ", squares[i])
		}
		if i%8 == 7 {
			i -= 16
			fmt.Println()
		}
	}
	fmt.Println(squares[7])
}

// Set cb.EvalMidGamePST and cb.EvalEndGamePST
func (cb *Board) resetMidGameEndGamePST() {
	allPieces := [2][5]uint64{
		{cb.Pawns[0], cb.Knights[0], cb.Bishops[0], cb.Rooks[0], cb.Queens[0]},
		{cb.Pawns[1], cb.Knights[1], cb.Bishops[1], cb.Rooks[1], cb.Queens[1]},
	}
	cb.EvalMidGamePST = 0
	cb.EvalEndGamePST = 0

	// Find piece locations, exluding pawns
	for color := range allPieces {
		for piece := range allPieces[color] {
			for allPieces[color][piece] > 0 {
				square := bits.TrailingZeros64(allPieces[color][piece])
				if color == 0 {
					cb.EvalMidGamePST -= MgTables[piece][square]
					cb.EvalEndGamePST -= EgTables[piece][square]
				} else {
					cb.EvalMidGamePST += MgTables[piece][square^56]
					cb.EvalEndGamePST += EgTables[piece][square^56]
				}
				allPieces[color][piece] &= allPieces[color][piece] - 1
			}
		}
	}
	cb.EvalMidGamePST += MgTables[5][cb.KingSqs[1]^56] - MgTables[5][cb.KingSqs[0]]
	cb.EvalEndGamePST += EgTables[5][cb.KingSqs[1]^56] - EgTables[5][cb.KingSqs[0]]
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

var MgTables = [6][64]int{pawnMG, knightMG, bishopMG, rookMG, queenMG, kingMG}
var EgTables = [6][64]int{pawnEG, knightEG, bishopEG, rookEG, queenEG, kingEG}
