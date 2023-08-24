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

func TestNegamax(t *testing.T) {
	cb, err := board.FromFen("8/8/8/8/8/8/8/Rr6 w - - 0 1")
	if err != nil {
		t.Error(err)
	}
	eval, bestMove := negamax(-(1 << 30), 1<<30, 1, cb)
	expected := board.Move{From: 0, To: 1, Piece: "r", PromoteTo: ""}
	if bestMove != expected {
		t.Errorf("negamax best move: want=%v, got=%v, eval=%d",
			expected, bestMove, eval)
	}
}

func runEvalTests(t *testing.T, tests []evalTestCase) {
	for _, tt := range tests {
		actual := evaluate(tt.cb)
		if tt.expected != actual {
			t.Errorf("pawnEval: want=%d, got=%d", tt.expected, actual)
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
