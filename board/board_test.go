package board

import (
	"testing"
)

// cb = chessboard, w = white, b = black, bb = bitboard

func TestNew(t *testing.T) {
	cb := New()

	if cb.WToMove != 1 {
		t.Errorf("initial wToMove: want=1, got=%d", cb.WToMove)
	}

	if cb.WPieces != uint64(1<<16)-1 {
		t.Errorf("initial wPieces: want=65535, got=%d", cb.WPieces)
	}
	if cb.WPawns != uint64(1<<16)-1-(1<<8-1) {
		t.Errorf("initial wPawns: want=%b, got=%b", 65279, cb.WPawns)
	}

	wPiecesUnion := cb.WPawns | cb.WRooks | cb.WKnights | cb.WBishops |
		cb.WQueens | cb.WKing
	if cb.WPieces != wPiecesUnion {
		t.Errorf("wPieces != union of all white pieces. want=65535,\ngot=%b",
			wPiecesUnion)
	}

	bPieces := uint64(1<<63) - 1 + 1<<63 - (1<<48 - 1)
	if cb.BPieces != bPieces {
		t.Errorf("initial bPieces: want=%b\n, got=%b", bPieces, cb.BPieces)
	}
	bPawns := uint64(1<<56) - 1 - (1<<48 - 1)
	if cb.BPawns != bPawns {
		t.Errorf("initial bPawns: want=%b\n, got=%b", bPawns, cb.BPawns)
	}

	bPiecesUnion := cb.BPawns | cb.BRooks | cb.BKnights | cb.BBishops |
		cb.BQueens | cb.BKing
	if cb.BPieces != bPiecesUnion {
		t.Errorf("bPieces != union of all black pieces. want=%b,\ngot=%b",
			uint64(1<<63)-1+(1<<63)-(1<<48-1), bPiecesUnion)
	}
}

type bbTestCase struct {
	square   int
	expected uint64
	actual   uint64
	name     string
}

func runMoveBBTests(t *testing.T, tests []bbTestCase) {
	for _, tt := range tests {
		if tt.actual != tt.expected {
			t.Errorf("incorrect bitboard for %v on square %d.\nexpected:\n%b, %T\ngot:\n%b, %T",
				tt.name, tt.square, tt.expected, tt.expected, tt.actual, tt.actual)
		}
	}
}

func TestPawnAttackBBs(t *testing.T) {
	bbs := makePawnBBs()
	tests := []bbTestCase{
		{
			square:   8,
			expected: uint64(1 << 17),
			actual:   bbs[1][8],
			name:     "wPawn",
		},
		{
			square:   8,
			expected: uint64(1 << 1),
			actual:   bbs[0][8],
			name:     "bPawn",
		},
		{
			square:   15,
			expected: uint64(1 << 22),
			actual:   bbs[1][15],
			name:     "wPawn",
		},
		{
			square:   15,
			expected: uint64(1 << 6),
			actual:   bbs[0][15],
			name:     "bPawn",
		},
		{
			square:   28,
			expected: uint64(1<<37 + 1<<35),
			actual:   bbs[1][28],
			name:     "wPawn",
		},
		{
			square:   28,
			expected: uint64(1<<21 + 1<<19),
			actual:   bbs[0][28],
			name:     "bPawn",
		},
		{
			square:   100,
			expected: uint64(0),
			actual:   bbs[1][5] + bbs[1][60] + bbs[0][0] + bbs[0][60],
			name:     "pawns on 1st and 8th ranks",
		},
	}

	runMoveBBTests(t, tests)
}

func TestKnightBBs(t *testing.T) {
	bbs := MakeKnightBBs()

	tests := []bbTestCase{
		{
			// file A
			square:   0,
			expected: uint64(1)<<17 + 1<<10,
			actual:   bbs[0],
			name:     "knight",
		},
		{
			// file B
			square:   57,
			expected: uint64(1)<<(57-17) + 1<<(57-15) + 1<<(57-6),
			actual:   bbs[57],
			name:     "knight",
		},
		{
			// file G
			square:   6,
			expected: uint64(1)<<(6+17) + 1<<(6+15) + 1<<(6+6),
			actual:   bbs[6],
			name:     "knight",
		},
		{
			// file H
			square:   55,
			expected: uint64(1)<<(55-17) + 1<<(55+6) + 1<<(55-10),
			actual:   bbs[55],
			name:     "knight",
		},
		{
			// center of board
			square: 20,
			expected: uint64(1)<<(20+17) + 1<<(20+15) + 1<<(20-17) + 1<<(20-15) +
				1<<(20-10) + 1<<(20+6) + 1<<(20+10) + 1<<(20-6),
			actual: bbs[20],
			name:   "knight",
		},
	}

	runMoveBBTests(t, tests)
}

func TestKingBBs(t *testing.T) {
	bbs := makeKingBBs()
	tests := []bbTestCase{
		{
			square:   0,
			expected: uint64(1<<8 + 1<<9 + 1<<1),
			actual:   bbs[0],
			name:     "king",
		},
		{
			square:   7,
			expected: uint64(1<<(7-1) + 1<<(7+8) + 1<<(7+7)),
			actual:   bbs[7],
			name:     "king",
		},
		{
			square: 1,
			expected: uint64(1<<(1-1) +
				1<<(1+1) +
				1<<(1+7) +
				1<<(1+8) +
				1<<(1+9)),
			actual: bbs[1],
			name:   "king",
		},
		{
			square: 62,
			expected: uint64(1)<<(62+1) +
				uint64(1)<<(62-1) +
				uint64(1)<<(62-9) +
				uint64(1)<<(62-8) +
				uint64(1)<<(62-7),
			actual: bbs[62],
			name:   "king",
		},
	}

	runMoveBBTests(t, tests)
}

func TestSlidingAttackBBs(t *testing.T) {
	bbs := makeSlidingAttackBBs()
	tests := []bbTestCase{
		{
			square: 0,
			expected: uint64(1)<<8 +
				uint64(1)<<16 +
				uint64(1)<<24 +
				uint64(1)<<32 +
				uint64(1)<<40 +
				uint64(1)<<48 +
				uint64(1)<<56,
			// Direction 0 is north, 1 is northeast, etc.
			actual: bbs[0][0],
			name:   "N ray",
		},
		{
			square: 0,
			expected: uint64(1)<<9 +
				uint64(1)<<18 +
				uint64(1)<<27 +
				uint64(1)<<36 +
				uint64(1)<<45 +
				uint64(1)<<54 +
				uint64(1)<<63,
			actual: bbs[1][0],
			name:   "NE ray",
		},
		{
			square: 0,
			expected: uint64(1)<<1 +
				uint64(1)<<2 +
				uint64(1)<<3 +
				uint64(1)<<4 +
				uint64(1)<<5 +
				uint64(1)<<6 +
				uint64(1)<<7,
			actual: bbs[2][0],
			name:   "E ray",
		},
		{
			square:   0,
			expected: uint64(0),
			actual:   bbs[3][0],
			name:     "SE ray",
		},
		{
			square: 9,
			expected: uint64(1)<<17 +
				uint64(1)<<25 +
				uint64(1)<<33 +
				uint64(1)<<41 +
				uint64(1)<<49 +
				uint64(1)<<57,
			actual: bbs[0][9],
			name:   "N ray",
		},
		{
			square: 9,
			expected: uint64(1)<<18 +
				uint64(1)<<27 +
				uint64(1)<<36 +
				uint64(1)<<45 +
				uint64(1)<<54 +
				uint64(1)<<63,
			actual: bbs[1][9],
			name:   "NE ray",
		},
		{
			square: 9,
			expected: uint64(1)<<10 +
				uint64(1)<<11 +
				uint64(1)<<12 +
				uint64(1)<<13 +
				uint64(1)<<14 +
				uint64(1)<<15,
			actual: bbs[2][9],
			name:   "E ray",
		},
		{
			square:   9,
			expected: uint64(1) << 2,
			actual:   bbs[3][9],
			name:     "SE ray",
		},
		{
			square:   9,
			expected: uint64(1) << 1,
			actual:   bbs[4][9],
			name:     "S ray",
		},
		{
			square:   9,
			expected: uint64(1),
			actual:   bbs[5][9],
			name:     "SW ray",
		},
		{
			square:   9,
			expected: uint64(1) << 8,
			actual:   bbs[6][9],
			name:     "W ray",
		},
		{
			square:   9,
			expected: uint64(1) << 16,
			actual:   bbs[7][9],
			name:     "NW ray",
		},
	}

	runMoveBBTests(t, tests)
}
