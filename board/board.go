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
	}
	cb.resetZobrist()

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

	if firstSpace == -1 || secondSpace != 1 {
		return cb, fmt.Errorf("invalid FEN string")
	}
	slashCount := strings.Count(fen, "/")
	if slashCount != 7 {
		return cb, fmt.Errorf("invalid FEN slash count. want=7, got=%d", slashCount)
	}

	squaresInRank := 0

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
		case char == 'n' || char == 'N':
			cb.Knights[color] += 1 << square
		case char == 'b' || char == 'B':
			cb.Bishops[color] += 1 << square
		case char == 'r' || char == 'R':
			cb.Rooks[color] += 1 << square
		case char == 'q' || char == 'Q':
			cb.Queens[color] += 1 << square
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

	cb.resetZobrist()

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
