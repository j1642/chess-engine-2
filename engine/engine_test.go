package engine

import (
	"engine2/board"
	_ "engine2/pieces"
	"testing"
)

func TestEvaluateNewBoard(t *testing.T) {
	eval := evaluate(board.New())
	if eval != 0 {
		t.Errorf("start position eval != 0, got=%d", eval)
	}
}

type pawnTestCase struct {
	cb       *board.Board
	expected int
}

func runPawnEvalTests(t *testing.T, tests []pawnTestCase) {
	for _, tt := range tests {
		actual := evalPawns(tt.cb)
		if tt.expected != actual {
			t.Errorf("pawnEval: want=%d, got=%d", tt.expected, actual)
		}
	}
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
	tests := []pawnTestCase{
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
}
