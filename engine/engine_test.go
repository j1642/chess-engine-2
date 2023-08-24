package engine

import (
	"engine2/board"
	"testing"
)

func TestEvaluate(t *testing.T) {
	eval := evaluate(board.New())
	if eval != 0 {
		t.Errorf("starting position evaluation != 0, got=%d", eval)
	}
}

func TestNegamax(t *testing.T) {
	eval := negamax(-(1 << 30), 1<<30, 2, board.New())
	if eval != 0 {
		t.Errorf("negamax start position eval != 0, got=%d", eval)
	}
}
