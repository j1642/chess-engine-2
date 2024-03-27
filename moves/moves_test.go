package moves

import (
	"github.com/j1642/chess-engine-2/board"
	"math/bits"
	"testing"
)

type bbTestCase struct {
	square   int
	expected uint64
	actual   uint64
	name     string
}

func TestSlidingAttackBBs(t *testing.T) {
	bbs := makeSlidingAttackBBs()
	tests := []bbTestCase{
		{
			square: 0,
			expected: uint64(1)<<8 + uint64(1)<<16 + uint64(1)<<24 +
				uint64(1)<<32 + uint64(1)<<40 + uint64(1)<<48 +
				uint64(1)<<56,
			// Direction 0 is north, 1 is northeast, etc.
			actual: bbs[0][0],
			name:   "N ray",
		},
		{
			square: 0,
			expected: uint64(1)<<9 + uint64(1)<<18 + uint64(1)<<27 +
				uint64(1)<<36 + uint64(1)<<45 + uint64(1)<<54 +
				uint64(1)<<63,
			actual: bbs[1][0],
			name:   "NE ray",
		},
		{
			square: 0,
			expected: uint64(1)<<1 + uint64(1)<<2 + uint64(1)<<3 +
				uint64(1)<<4 + uint64(1)<<5 + uint64(1)<<6 +
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
			expected: uint64(1)<<17 + uint64(1)<<25 + uint64(1)<<33 +
				uint64(1)<<41 + uint64(1)<<49 + uint64(1)<<57,
			actual: bbs[0][9],
			name:   "N ray",
		},
		{
			square: 9,
			expected: uint64(1)<<18 + uint64(1)<<27 + uint64(1)<<36 +
				uint64(1)<<45 + uint64(1)<<54 + uint64(1)<<63,
			actual: bbs[1][9],
			name:   "NE ray",
		},
		{
			square: 9,
			expected: uint64(1)<<10 + uint64(1)<<11 + uint64(1)<<12 +
				uint64(1)<<13 + uint64(1)<<14 + uint64(1)<<15,
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
		// Moves cannot pass the board edges
		{
			square:   7,
			expected: uint64(0),
			actual:   bbs[2][7],
			name:     "E ray",
		},
		{
			square:   56,
			expected: uint64(0),
			actual:   bbs[0][56],
			name:     "N ray",
		},
		{
			square:   7,
			expected: uint64(0),
			actual:   bbs[1][7],
			name:     "NE ray",
		},
	}

	// Can't import board.RunMoveBBTests or BbTestCase from board_test.go after
	// making them public?
	runMoveBBTests(t, tests)
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
			expected: uint64(1<<12 + 1<<14 + 1<<51 + 1<<53),
			actual:   bbs[1][5] + bbs[1][60] + bbs[0][0] + bbs[0][60],
			name:     "pawns on 1st and 8th ranks",
		},
	}

	runMoveBBTests(t, tests)
}

func TestKnightBBs(t *testing.T) {
	bbs := makeKnightBBs()

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
			expected: uint64(1<<(1-1) + 1<<(1+1) + 1<<(1+7) +
				1<<(1+8) + 1<<(1+9)),
			actual: bbs[1],
			name:   "king",
		},
		{
			square: 62,
			expected: uint64(1)<<(62+1) + uint64(1)<<(62-1) + uint64(1)<<(62-9) +
				uint64(1)<<(62-8) + uint64(1)<<(62-7),
			actual: bbs[62],
			name:   "king",
		},
	}

	runMoveBBTests(t, tests)
}

func TestCalculateBishopMoves(t *testing.T) {
	cb := board.New()
	tests := []moveTestCase{
		{
			expected: uint64(1 << 9),
			actual:   CalculateBishopMoves(0, cb),
		},
		{
			expected: uint64(1 + 1<<16 + 1<<2 + 1<<18 + 1<<27 + 1<<36 + 1<<45 + 1<<54),
			actual:   CalculateBishopMoves(9, cb),
		},
		{
			expected: uint64(1<<10 + 1<<19 + 1<<37 + 1<<46 + 1<<55 +
				1<<14 + 1<<21 + 1<<35 + 1<<42 + 1<<49),
			actual: CalculateBishopMoves(28, cb),
		},
	}

	runMoveGenTests(t, tests)
}

func TestRookMoves(t *testing.T) {
	cb := board.New()
	tests := []moveTestCase{
		{
			expected: uint64(1<<1 + 1<<8),
			actual:   CalculateRookMoves(0, cb),
		},
		{
			expected: uint64(1<<62 + 1<<55),
			actual:   CalculateRookMoves(63, cb),
		},
		{
			expected: uint64(1<<24) - 1 - (1<<16 - 1) - 1<<20 +
				1<<12 + 1<<28 + 1<<36 + 1<<44 + 1<<52,
			actual: CalculateRookMoves(20, cb),
		},
	}

	runMoveGenTests(t, tests)
}

type moveTestCase struct {
	expected uint64
	actual   uint64
}

func runMoveGenTests(t *testing.T, tests []moveTestCase) {
	for _, tt := range tests {
		if tt.expected != tt.actual {
			t.Errorf("want=%v, got=%v", read1Bits(tt.expected), read1Bits(tt.actual))
		}
	}
}

func read1Bits(bb uint64) []int8 {
	// Using TrailingZeros64() seems as fast as bitshifting right while bb>0.
	squares := make([]int8, 0, 4)
	for bb > 0 {
		squares = append(squares, int8(bits.TrailingZeros64(bb)))
		bb &= bb - 1
	}
	return squares
}
