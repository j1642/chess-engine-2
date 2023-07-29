package pieces

import (
	"math"
	"testing"
)

type bbTestCase struct {
	square   int
	expected uint64
	actual   uint64
	name     string
}

func TestKnightBBs(t *testing.T) {
	bbs := MakeKnightBBs()

	tests := []bbTestCase{
		{
			// file A
			square:   0,
			expected: uint64(math.Pow(2, 17) + math.Pow(2, 10)),
			actual:   bbs[0],
			name:     "knight",
		},
		{
			// file B
			square: 57,
			expected: uint64(math.Pow(2, 57-17) + math.Pow(2, 57-15) +
				1<<(57-6)),
			actual: bbs[57],
			name:   "knight",
		},
		{
			// file G
			square: 6,
			expected: uint64(math.Pow(2, 6+17) + math.Pow(2, 6+15) +
				1<<(6+6)),
			actual: bbs[6],
			name:   "knight",
		},
		{
			// file H
			square: 55,
			expected: uint64(math.Pow(2, 55-17) + math.Pow(2, 55+6) +
				1<<(55-10)),
			actual: bbs[55],
			name:   "knight",
		},
		{
			// center of board
			square: 20,
			expected: uint64(math.Pow(2, 20+17) + math.Pow(2, 20+15) +
				math.Pow(2, 20-17) + math.Pow(2, 20-15) + 1<<(20-10) +
				math.Pow(2, 20+6) + math.Pow(2, 20+10) + math.Pow(2, 20-6)),
			actual: bbs[20],
			name:   "knight",
		},
	}

	runBBTest(t, tests)
}

func runBBTest(t *testing.T, tests []bbTestCase) {
	for _, tt := range tests {
		if tt.actual != tt.expected {
			t.Errorf("incorrect bitboard moves for a %v on square %d.\nexpected:\n%b, %T\ngot:\n%b, %T",
				tt.name, tt.square, tt.expected, tt.expected, tt.actual, tt.actual)
		}
	}
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

	runBBTest(t, tests)
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

	runBBTest(t, tests)
}

func TestBinSearch(t *testing.T) {
	nums := [8]int{0, 0, 0, 1, 2, 3, 4, 5}
	if !binSearch(1, nums) {
		t.Error("binSearch should find 1 in nums")
	}
	if !binSearch(3, nums) {
		t.Error("binSearch should find 3 in nums")
	}
	if binSearch(6, nums) {
		t.Error("binSearch should not find 6 in nums")
	}
}
