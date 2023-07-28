package pieces

import (
	"math"
	"testing"
)

type bbTestCase struct {
	square   int
	expected uint64
	actual   uint64
}

func TestKnightBBs(t *testing.T) {
	bbs := MakeKnightBBs()

	tests := []bbTestCase{
		{
			// file A
			square:   0,
			expected: uint64(math.Pow(2, 17) + math.Pow(2, 10)),
			actual:   bbs[0],
		},
		{
			// file B
			square: 57,
			expected: uint64(math.Pow(2, 57-17) + math.Pow(2, 57-15) +
				1<<(57-6)),
			actual: bbs[57],
		},
		{
			// file G
			square: 6,
			expected: uint64(math.Pow(2, 6+17) + math.Pow(2, 6+15) +
				1<<(6+6)),
			actual: bbs[6],
		},
		{
			// file H
			square: 55,
			expected: uint64(math.Pow(2, 55-17) + math.Pow(2, 55+6) +
				1<<(55-10)),
			actual: bbs[55],
		},
		{
			// center of board
			square: 20,
			expected: uint64(math.Pow(2, 20+17) + math.Pow(2, 20+15) +
				math.Pow(2, 20-17) + math.Pow(2, 20-15) + 1<<(20-10) +
				math.Pow(2, 20+6) + math.Pow(2, 20+10) + math.Pow(2, 20-6)),
			actual: bbs[20],
		},
	}

	runBBTest(t, tests, "knight")
}

func runBBTest(t *testing.T, tests []bbTestCase, piece string) {
	for _, tt := range tests {
		if tt.actual != tt.expected {
			t.Errorf("incorrect bitboard moves for a %v on square %d.\nexpected:\n%b, %T\ngot:\n%b, %T",
				piece, tt.square, tt.expected, tt.expected, tt.actual, tt.actual)
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
		},
		{
			square:   7,
			expected: uint64(1<<(7-1) + 1<<(7+8) + 1<<(7+7)),
			actual:   bbs[7],
		},
		{
			square: 1,
			expected: uint64(1<<(1-1) +
				1<<(1+1) +
				1<<(1+7) +
				1<<(1+8) +
				1<<(1+9)),
			actual: bbs[1],
		},
		{
			square: 62,
			expected: uint64(1)<<(62+1) +
				uint64(1)<<(62-1) +
				uint64(1)<<(62-9) +
				uint64(1)<<(62-8) +
				uint64(1)<<(62-7),
			actual: bbs[62],
		},
	}

	runBBTest(t, tests, "king")
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
