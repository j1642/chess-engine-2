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

type validMoveTestCase struct {
	from, to         int
	cb               *board.Board
	expected, actual bool
}

func runValidMoveTests(t *testing.T, tests []validMoveTestCase) {
	for _, tt := range tests {
		if tt.expected != tt.actual {
			t.Errorf("move %d to %d: want=%v, got=%v",
				tt.from, tt.to, tt.expected, tt.actual)
		}
	}
}

func firstVal(b bool, s string) bool {
	return b
}

func TestIsValidMove(t *testing.T) {
	cb := board.New()
	tests := []validMoveTestCase{
		{
			from: -1, to: 0,
			cb:       cb,
			expected: false,
			actual:   firstVal(isValidMove(-1, 0, cb)),
		},
		{
			from: 0, to: 100,
			cb:       cb,
			expected: false,
			actual:   firstVal(isValidMove(0, 100, cb)),
		},
		// No piece present
		{
			from: 20, to: 21,
			cb:       cb,
			expected: false,
			actual:   firstVal(isValidMove(20, 21, cb)),
		},
		// Pawn
		{
			from: 8, to: 16,
			cb:       cb,
			expected: true,
			actual:   firstVal(isValidMove(8, 16, cb)),
		},
		{
			from: 8, to: 8,
			cb:       cb,
			expected: false,
			actual:   firstVal(isValidMove(8, 8, cb)),
		},
		// Knight
		{
			from: 1, to: 18,
			cb:       cb,
			expected: true,
			actual:   firstVal(isValidMove(1, 18, cb)),
		},
		{
			from: 6, to: 22,
			cb:       cb,
			expected: false,
			actual:   firstVal(isValidMove(6, 22, cb)),
		},
		// Bishop
		{
			from: 2, to: 47,
			cb:       cb,
			expected: true,
			actual:   firstVal(isValidMove(2, 47, cb)),
		},
		// Bishop passing edge
		{
			from: 5, to: 32,
			cb:       cb,
			expected: false,
			actual:   firstVal(isValidMove(5, 32, cb)),
		},
		// Rook
		{
			from: 0, to: 7,
			cb:       cb,
			expected: true,
			actual:   firstVal(isValidMove(0, 7, cb)),
		},
		{
			from: 7, to: 63,
			cb:       cb,
			expected: true,
			actual:   firstVal(isValidMove(7, 63, cb)),
		},
		// Rook passing edge
		{
			from: 7, to: 8,
			cb:       cb,
			expected: false,
			actual:   firstVal(isValidMove(7, 8, cb)),
		},
		// Queen
		{
			from: 3, to: 7,
			cb:       cb,
			expected: true,
			actual:   firstVal(isValidMove(3, 7, cb)),
		},
		{
			from: 7, to: 8,
			cb:       cb,
			expected: false,
			actual:   firstVal(isValidMove(7, 8, cb)),
		},
		{
			from: 7, to: 16,
			cb:       cb,
			expected: false,
			actual:   firstVal(isValidMove(7, 16, cb)),
		},
		// King
		{
			from: 40, to: 39,
			cb:       cb,
			expected: false,
			actual:   firstVal(isValidMove(40, 39, cb)),
		},
		{
			from: 4, to: 13,
			cb:       cb,
			expected: true,
			actual:   firstVal(isValidMove(4, 13, cb)),
		},
	}

	runValidMoveTests(t, tests)
}

func TestMovePiece(t *testing.T) {
	cb := board.New()
	movePiece(8, 16, cb)

	if cb.WToMove != 0 {
		t.Errorf("WToMove: want=0, got=%d", cb.WToMove)
	}
	if cb.BwPawns[1] != uint64(1<<17)-1-1<<8-(1<<8-1) {
		t.Errorf("wPawns: want=\n%b,\ngot=\n%b\n",
			uint64(1<<17)-1-1<<8-(1<<8-1), cb.BwPawns[1])
	}
	if cb.BwPieces[1] != uint64(1<<17)-1-1<<8 {
		t.Errorf("wPieces: want=\n%b,\ngot=\n%b",
			uint64(1<<17)-1-1<<8, cb.BwPieces[1])
	}
	movePiece(57, 42, cb)
	movePiece(16, 24, cb)
	// Waiting move by ng8
	movePiece(62, 45, cb)
	movePiece(24, 32, cb)

	movePiece(42, 32, cb)

	if cb.WToMove != 1 {
		t.Errorf("WToMove: want=1, got=%d", cb.WToMove)
	}
	if cb.BwPawns[1] != uint64(1<<16)-1-1<<8-(1<<8-1) {
		t.Errorf("wPawns: want=\n%b,\ngot=\n%b\n",
			uint64(1<<16)-1-1<<8-(1<<8-1), cb.BwPawns[1])
	}
	if cb.BwPieces[1] != uint64(1<<16)-1-1<<8 {
		t.Errorf("wPieces: want=\n%b,\ngot=\n%b",
			uint64(1<<16)-1-1<<8, cb.BwPieces[1])
	}

	if cb.BwKnights[0] != uint64(1<<45+1<<32) {
		t.Errorf("bKnights: want=\n%b,\ngot=\n%b",
			uint64(1<<45+1<<32), cb.BwKnights[0])
	}
	bPieces := ^uint64(0) - (1<<48 - 1) + 1<<32 + 1<<45 - 1<<57 - 1<<62
	if cb.BwPieces[0] != bPieces {
		t.Errorf("bPieces: want=\n%b,\ngot=\n%b",
			bPieces, cb.BwPieces[0])
	}
}

func TestPromotePawn(t *testing.T) {
	// TODO: mock user input to test other promotePawn() branch
	cb := &board.Board{
		WToMove:  1,
		BwPawns:  [2]uint64{1 << 1, 1 << 63},
		BwQueens: [2]uint64{0, 0},
	}
	promotePawn(uint64(1<<63), cb, "q")

	if cb.BwPawns[1] != uint64(0) {
		t.Errorf("pawn did not promote: want=0, got=%b", cb.BwPawns[1])
	}
	if cb.BwQueens[1] != uint64(1<<63) {
		t.Error("promoted queen not present")
	}
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
