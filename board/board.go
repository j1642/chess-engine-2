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

	PAttacks       [2][64]uint64
	NAttacks       [64]uint64
	KAttacks       [64]uint64
	SlidingAttacks [8][64]uint64

	KingSqs      [2]int
	CastleRights [2][2]bool // [b, w][queenside, kingside]

	EpSquare int

	PrevMove    Move
	Zobrist     uint64
}

type Move struct {
	From, To  int
	Piece     rune
	PromoteTo rune
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

		PAttacks:       makePawnBBs(),
		NAttacks:       makeKnightBBs(),
		KAttacks:       makeKingBBs(),
		SlidingAttacks: makeSlidingAttackBBs(),

		KingSqs:      [2]int{60, 4},
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
			//pawnsBB := cb.Pawns[color]
			for pieceBB > 0 {
				zobrist ^= ZobristKeys.ColorPieceSq[color][i][bits.TrailingZeros64(pieceBB)]
				fmt.Println("color, piece, square:", color, i, bits.TrailingZeros64(pieceBB))
				pieceBB &= pieceBB - 1
			}
		}
		/*
		   knightsBB := cb.Knights[color]
		   for knightsBB > 0 {
		       zobrist ^= ZobristKeys.ColorPieceSq[color][1][bits.TrailingZeros64(knightsBB)]
		       knightsBB &= knightsBB - 1
		   }
		   bishopsBB := cb.Bishops[color]
		   for bishopsBB > 0 {
		       zobrist ^= ZobristKeys.ColorPieceSq[color][2][bits.TrailingZeros64(bishopsBB)]
		       bishopsBB &= bishopsBB - 1
		   }
		   rooksBB := cb.Rooks[color]
		   for rooksBB > 0 {
		       zobrist ^= ZobristKeys.ColorPieceSq[color][3][bits.TrailingZeros64(rooksBB)]
		       rooksBB &= rooksBB - 1
		   }
		   queensBB := cb.Queens[color]
		   for queensBB > 0 {
		       zobrist ^= ZobristKeys.ColorPieceSq[color][4][bits.TrailingZeros64(queensBB)]
		       queensBB &= queensBB - 1
		   }
		   kingsBB := cb.Kings[color]
		   for kingsBB > 0 {
		       zobrist ^= ZobristKeys.ColorPieceSq[color][5][bits.TrailingZeros64(kingsBB)]
		       kingsBB &= kingsBB - 1
		   }
		*/
	}

	for i := range len(cb.CastleRights) {
		for j := range len(cb.CastleRights[0]) {
			if cb.CastleRights[i][j] {
				zobrist ^= ZobristKeys.Castle[i][j]
				fmt.Println("castle color, qk:", i, j)
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
func FromFen(fen string) (*Board, error) {
	var color int
	cb := &Board{}
	square := 56
	firstSpace := strings.IndexByte(fen, ' ')
	secondSpace := strings.IndexByte(fen[firstSpace+1:], ' ')

	if firstSpace == -1 || secondSpace != 1 {
		return cb, fmt.Errorf("invalid FEN string")
	}
	slashCount := strings.Count(fen, "/")
	if slashCount != 7 {
		return cb, fmt.Errorf("invalid FEN slash count. want=7, got=%d", slashCount)
	}

	for _, char := range fen[:firstSpace] {
		if 'A' <= char && char <= 'Z' {
			color = 1
		} else if 'a' <= char && char <= 'z' {
			color = 0
		} else {
			color = 100 // placeholder value
		}

		switch {
		case '1' <= char && char <= '8':
			// Negate the "square += 1" at the end of the loop
			square += int(char-'0') - 1
		case char == '/':
			// Negate the "square += 1" at the end of the loop
			square -= 17
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

	// TODO: Include move count?
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
			rank := 8 * (int(fen[i+firstSpace+1]-'0') - 1)
			cb.EpSquare = (int(char - 'a')) + rank
		}
	}

	cb.Pieces[0] = cb.Pawns[0] | cb.Knights[0] | cb.Bishops[0] |
		cb.Rooks[0] | cb.Queens[0] | cb.Kings[0]
	cb.Pieces[1] = cb.Pawns[1] | cb.Knights[1] | cb.Bishops[1] |
		cb.Rooks[1] | cb.Queens[1] | cb.Kings[1]

	cb.PAttacks = makePawnBBs()
	cb.NAttacks = makeKnightBBs()
	cb.KAttacks = makeKingBBs()
	cb.SlidingAttacks = makeSlidingAttackBBs()
	cb.resetZobrist()

	return cb, nil
}

func GetFiles() [4][8]int {
	fileA, fileB, fileG, fileH := [8]int{}, [8]int{}, [8]int{}, [8]int{}

	for i := 0; i < 8; i++ {
		fileA[i] = i * 8
		fileB[i] = i*8 + 1
		fileG[i] = i*8 + 6
		fileH[i] = i*8 + 7
	}

	return [4][8]int{fileA, fileB, fileG, fileH}
}

type IntArray interface {
	[8]int | [4]int | [3]int
}

// Linear search for small arrays
func ContainsN[T IntArray](n int, nums T) bool {
	for i := 0; i < len(nums); i++ {
		if n == nums[i] {
			return true
		}
	}
	return false
}

// Return pawn attack bitboards, so attacks aren't repeatedly calculated on the fly
func makePawnBBs() [2][64]uint64 {
	// First index is cb.WToMove: 1 for white pawns, 0 for black pawns.
	bbs := [2][64]uint64{}
	for sq := 0; sq < 64; sq++ {
		switch {
		// Used in pieces.getCheckingSquares() for pawn checks on the black king on the eighth rank
		// NOT FOR PAWN ATTACKS. Pawns on the first and eighth ranks checked in pieces.getPawnMoves()
		case sq > 56:
			if sq%8 == 0 {
				bbs[0][sq] += 1 << (sq - 7)
			} else if sq%8 == 7 {
				bbs[0][sq] += 1 << (sq - 9)
			} else {
				bbs[0][sq] += 1<<(sq-7) + 1<<(sq-9)
			}
		// Used in pieces.getCheckingSquares() for pawn checks on the black king on the eighth rank
		// NOT FOR PAWN ATTACKS. Pawns on the first and eighth ranks checked in pieces.getPawnMoves()
		case sq < 8:
			if sq%8 == 0 {
				bbs[1][sq] += 1 << (sq + 9)
			} else if sq%8 == 7 {
				bbs[1][sq] += 1 << (sq + 7)
			} else {
				bbs[1][sq] += 1<<(sq+7) + 1<<(sq+9)
			}
		case sq%8 == 0:
			bbs[1][sq] += 1 << (sq + 9)
			bbs[0][sq] += 1 << (sq - 7)
		case sq%8 == 7:
			bbs[1][sq] += 1 << (sq + 7)
			bbs[0][sq] += 1 << (sq - 9)
		default:
			bbs[1][sq] += 1<<(sq+7) + 1<<(sq+9)
			bbs[0][sq] += 1<<(sq-7) + 1<<(sq-9)
		}
	}

	return bbs
}

func makeKnightBBs() [64]uint64 {
	bbs := [64]uint64{}
	var directions []int
	files := GetFiles()

	for sq := 0; sq < 64; sq++ {
		switch {
		case ContainsN(sq, files[0]):
			directions = []int{17, 10, -6, -15}
		case ContainsN(sq, files[1]):
			directions = []int{17, 15, 10, -6, -17, -15}
		case ContainsN(sq, files[2]):
			directions = []int{17, 15, -17, -15, 6, -10}
		case ContainsN(sq, files[3]):
			directions = []int{15, -17, 6, -10}
		default:
			directions = []int{17, 15, 10, -6, -17, -15, 6, -10}
		}

		for _, d := range directions {
			if sq+d < 0 || sq+d > 63 {
				continue
			}
			bbs[sq] += 1 << (sq + d)
		}
	}

	return bbs
}

func makeKingBBs() [64]uint64 {
	bbs := [64]uint64{}
	var directions []int
	files := GetFiles()

	for sq := 0; sq < 64; sq++ {
		switch {
		// file A
		case ContainsN(sq, files[0]):
			directions = []int{8, 9, 1, -7, -8}
		// file H
		case ContainsN(sq, files[3]):
			directions = []int{8, 7, -1, -9, -8}
		default:
			directions = []int{7, 8, 9, -1, 1, -9, -8, -7}
		}

		for _, d := range directions {
			if sq+d < 0 || sq+d > 63 {
				continue
			}
			bbs[sq] += 1 << (sq + d)
		}
	}

	return bbs
}

func makeSlidingAttackBBs() [8][64]uint64 {
	bbs := [8][64]uint64{}
	files := GetFiles()
	fileAForbidden := [3]int{-9, -1, 7}
	fileHForbidden := [3]int{9, 1, -7}

	// Movement directions are ordered clockwise, starting from north
	for i, dir := range [8]int{8, 9, 1, -7, -8, -9, -1, 7} {
		for sq := 0; sq < 64; sq++ {
			if ContainsN(sq, files[0]) && ContainsN(dir, fileAForbidden) {
				continue
			} else if ContainsN(sq, files[3]) && ContainsN(dir, fileHForbidden) {
				continue
			}

			for j := 1; j < 8; j++ {
				newSq := j*dir + sq
				if newSq < 0 || newSq > 63 {
					break
				}
				bbs[i][sq] += 1 << newSq
				// Found board edge
				if dir != 8 && dir != -8 &&
					(ContainsN(newSq, files[0]) || ContainsN(newSq, files[3])) {
					break
				}
			}
		}
	}

	return bbs
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

	KingSqs      [2]int
	CastleRights [2][2]bool

	EpSquare int

	PrevMove Move
	Zobrist  uint64
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

		EpSquare: cb.EpSquare,

		PrevMove: cb.PrevMove,
		Zobrist:  cb.Zobrist,
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
