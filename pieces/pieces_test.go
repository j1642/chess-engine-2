package pieces

import (
	"engine2/board"
	"testing"
)

type moveTestCase struct {
	expected uint64
	actual   uint64
}

func runMoveGenTests(t *testing.T, tests []moveTestCase) {
	for _, tt := range tests {
		if tt.expected != tt.actual {
			t.Errorf("want=\n%b,\ngot=\n%b", tt.expected, tt.actual)
		}
	}
}

func TestRookMoves(t *testing.T) {
	cb := board.New()
	tests := []moveTestCase{
		{
			expected: uint64(1<<1 + 1<<8),
			actual:   getRookMoves(0, cb),
		},
		{
			expected: uint64(1<<62 + 1<<55),
			actual:   getRookMoves(63, cb),
		},
		{
			expected: uint64(1<<24) - 1 - (1<<16 - 1) - 1<<20 +
				1<<12 + 1<<28 + 1<<36 + 1<<44 + 1<<52,
			actual: getRookMoves(20, cb),
		},
	}

	runMoveGenTests(t, tests)
}

func TestBishopMoves(t *testing.T) {
	cb := board.New()
	tests := []moveTestCase{
		{
			expected: uint64(1 << 9),
			actual:   getBishopMoves(0, cb),
		},
		{
			expected: uint64(1 + 1<<16 + 1<<2 + 1<<18 + 1<<27 + 1<<36 + 1<<45 + 1<<54),
			actual:   getBishopMoves(9, cb),
		},
		{
			expected: uint64(1<<10 + 1<<19 + 1<<37 + 1<<46 + 1<<55 +
				1<<14 + 1<<21 + 1<<35 + 1<<42 + 1<<49),
			actual: getBishopMoves(28, cb),
		},
	}

	runMoveGenTests(t, tests)
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
