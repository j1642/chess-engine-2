package uci

import (
	"reflect"
	"strings"
	"testing"

	"github.com/j1642/chess-engine-2/board"
	"github.com/j1642/chess-engine-2/pieces"
)

type testCase struct {
	expected, actual *board.Board
}

func TestSetPosition(t *testing.T) {
	// "position startpos moves e2e4 e7e5"
	// "position fen ... moves e2e4"
	newBoard := board.New()
	kingsPawn, err := board.FromFen("rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 1 3")
	if err != nil {
		t.Error(err)
	}
	kingsPawn.PrevMove = board.Move{From: 52, To: 36, Piece: pieces.PAWN, PromoteTo: pieces.NO_PIECE}
	kingsPawn.HalfMoves = 2

	startFromFen := "position fen rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	actual1 := buildPosition(strings.Fields(startFromFen))
	actual2 := buildPosition(strings.Fields("position startpos"))

	actual3 := buildPosition(strings.Fields(startFromFen + " moves"))
	actual4 := buildPosition(strings.Fields("position startpos moves"))

	actual5 := buildPosition(strings.Fields(startFromFen + " moves e2e4 e7e5"))
	actual6 := buildPosition(strings.Fields("position startpos moves e2e4 e7e5"))

	tests := []testCase{
		{
			expected: newBoard,
			actual:   actual1,
		},
		{
			expected: newBoard,
			actual:   actual2,
		},
		{
			expected: newBoard,
			actual:   actual3,
		},
		{
			expected: newBoard,
			actual:   actual4,
		},
		{
			expected: kingsPawn,
			actual:   actual5,
		},
		{
			expected: kingsPawn,
			actual:   actual6,
		},
	}

	for i, tt := range tests {
		valExpected := reflect.ValueOf(tt.expected).Elem()
		fieldTypeExpected := valExpected.Type()

		valActual := reflect.ValueOf(tt.actual).Elem()
		fieldTypeActual := valActual.Type()

		for i := 0; i < valExpected.NumField(); i++ {
			if valExpected.Field(i).Interface() != valActual.Field(i).Interface() {
				t.Errorf("fromActual failed:  want %s=%v, got %s=%v\n",
					fieldTypeExpected.Field(i).Name, valExpected.Field(i).Interface(),
					fieldTypeActual.Field(i).Name, valActual.Field(i).Interface())
			}
		}

		if *tt.actual != *tt.expected {
			t.Errorf("i=%d, actual != expected", i)
		}
	}
}
