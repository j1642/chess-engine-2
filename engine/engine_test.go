package engine

import (
	"engine2/board"
	"testing"
)

func TestEvaluate(t *testing.T) {
	if evaluate(board.New()) != 2 {
		t.Error("starting position evaluation != 2")
	}
}

func TestNegamax(t *testing.T) {
	if negamax(-(1<<30), 1<<30, 2, board.New()) != 2 {
		t.Error("negamax outline broken")
	}
}
