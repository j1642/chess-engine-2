package engine

import (
	"engine2/board"
	_ "engine2/pieces"
	"testing"
)

type evalTestCase struct {
	cb       *board.Board
	expected int
}

func TestEvaluate(t *testing.T) {
	cbRooks, err := board.FromFen("8/8/8/8/8/8/8/Rr6 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	cbLoneRook, err := board.FromFen("8/8/8/8/8/8/8/R7 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	cornerCheckmate, err := board.FromFen("8/8/8/8/8/5K2/6Q1/7k b - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	tests := []evalTestCase{
		{
			cb:       board.New(),
			expected: 0,
		},
		{
			cb:       cbRooks,
			expected: -6,
		},
		{
			cb:       cbLoneRook,
			expected: 64,
		},
		{
			cb:       cornerCheckmate,
			expected: 1048552,
		},
	}

	runEvalTests(t, tests)
}

func TestEvaluatePawns(t *testing.T) {
	cbDoubled, err := board.FromFen("8/1pp5/8/2p5/5PP1/8/5PP1/8 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	cbIsolated, err := board.FromFen("8/pp6/8/8/8/8/1P4PP/8 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	cbBlocked, err := board.FromFen("8/pp6/nn6/8/8/7N/5NPP/8 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	tests := []evalTestCase{
		{
			cb:       board.New(),
			expected: 0,
		},
		{
			cb:       cbDoubled,
			expected: -10,
		},
		{
			cb:       cbIsolated,
			expected: -5,
		},
		{
			cb:       cbBlocked,
			expected: 5,
		},
	}

	runPawnEvalTests(t, tests)
}

type searchTestCase struct {
	cb       *board.Board
	expected board.Move
	depth    int
}

func TestNegamax(t *testing.T) {
	rookVRook, err := board.FromFen("8/8/8/8/8/8/8/Rr6 w - - 0 1")
	if err != nil {
		t.Error(err)
	}
	/*
		mateIn5Ply, err := board.FromFen("r2qk2r/pb4pp/1n2Pb2/2B2Q2/p1p5/2P5/2B2PPP/RN2R1K1 w - - 1 0")
		if err != nil {
			t.Error(err)
		}
	*/
	mateIn3Ply, err := board.FromFen("r2qk2r/pb4pp/1n2PbQ1/2B5/p1p5/2P5/2B2PPP/RN2R1K1 b - - 1 0")
	if err != nil {
		t.Error(err)
	}
	mateIn2Ply, err := board.FromFen("r2qk2r/pb4p1/1n2Pbp1/2B5/p1p5/2P5/2B2PPP/RN2R1K1 w - - 1 0")
	if err != nil {
		t.Error(err)
	}

	// mate is detected when the side to move cannot move, so needs x+1 plies compared to "mate in X"
	tests := []searchTestCase{
		// Issue: returning empty moves when depth > 1
		{cb: rookVRook, expected: board.Move{From: 0, To: 1, Piece: 'r', PromoteTo: ' '}, depth: 1},
		{cb: mateIn3Ply, expected: board.Move{From: 55, To: 46, Piece: 'p', PromoteTo: ' '}, depth: 3},
		{cb: mateIn2Ply, expected: board.Move{From: 10, To: 46, Piece: 'b', PromoteTo: ' '}, depth: 2},
		//{cb: mateIn4Ply, expected: board.Move{From: 37, To: 46, Piece: 'q', PromoteTo: ' '}, depth: 4},
	}

	for _, tt := range tests {
		// Unexplained bug: using math.MinInt, math.MaxInt as args breaks negamax
		eval, actual, err := negamax(-(1 << 30), 1<<30, tt.depth, tt.cb)
		if err != nil {
			t.Fatal(err)
		}
		if actual != tt.expected {
			t.Errorf("negamax best move: want=%v, got=%v, eval=%d",
				tt.expected, actual, eval)
		}
		tt.cb.Print()
	}
}

func runEvalTests(t *testing.T, tests []evalTestCase) {
	for _, tt := range tests {
		actual := evaluate(tt.cb)
		if tt.expected != actual {
			t.Errorf("eval: want=%d, got=%d", tt.expected, actual)
		}
	}
}

func runPawnEvalTests(t *testing.T, tests []evalTestCase) {
	for _, tt := range tests {
		actual := evalPawns(tt.cb)
		if tt.expected != actual {
			t.Errorf("pawnEval: want=%d, got=%d", tt.expected, actual)
		}
	}
}
