package uci

import (
	"reflect"
	"strings"
	"testing"

	"github.com/j1642/chess-engine-2/board"
	"github.com/j1642/chess-engine-2/pieces"
)

type setPositionTestCase struct {
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

	tests := []setPositionTestCase{
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

type moveConversionTestCase struct {
	expectedTo, expectedFrom int8
	expectedPromoteTo        uint8
	input                    string
}

func TestConvertLongAlgebraicMoveToSquares(t *testing.T) {
	// TODO: add test cases where error != nil, test for != nil
	tests := []moveConversionTestCase{
		{
			// Normal move
			expectedFrom: 12, expectedTo: 28, expectedPromoteTo: pieces.NO_PIECE,
			input: "e2e4",
		},
		{
			// Promotion
			expectedFrom: 52, expectedTo: 60, expectedPromoteTo: pieces.QUEEN,
			input: "e7e8q",
		},
	}

	for _, tt := range tests {
		fromSq, toSq, promoteTo, err := convertLongAlgebraicMoveToSquares(tt.input)
		if fromSq != tt.expectedFrom || toSq != tt.expectedTo || promoteTo != tt.expectedPromoteTo {
			t.Errorf("from,to,promoteTo: want=%d,%d,%d, got=%d,%d,%d",
				tt.expectedFrom, tt.expectedTo, tt.expectedPromoteTo,
				fromSq, toSq, promoteTo,
			)
		}
		if err != nil {
			t.Error(err)
		}
	}
}

func TestBuildGoOptions(t *testing.T) {
	split := strings.Fields("go infinite depth 5 searchmoves d2d3 nodes 1000")
	actual := buildGoOptions(split)
	expected := goOptions{
		depth:       5,
		nodes:       1000,
		infinite:    true,
		searchmoves: []board.Move{{From: 11, To: 19, Piece: pieces.PAWN, PromoteTo: pieces.NO_PIECE}},
	}

	if len(actual.searchmoves) != len(expected.searchmoves) {
		t.Errorf("searchmoves lengths: %d != %d", len(actual.searchmoves), len(expected.searchmoves))
	}
	for i := range actual.searchmoves {
		if actual.searchmoves[i] != expected.searchmoves[i] {
			t.Errorf("moves[%d]: %v != %v", i, actual.searchmoves[i], expected.searchmoves[i])
		}
	}

	if actual.infinite != expected.infinite {
		t.Error("infinite")
	}
	if actual.depth != expected.depth {
		t.Error("depth")
	}
	if actual.nodes != expected.nodes {
		t.Error("nodes")
	}
}
