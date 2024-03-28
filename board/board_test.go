package board

import (
	"reflect"
	"testing"
)

// cb = chessboard, w = white, b = black, bb = bitboard

func TestNew(t *testing.T) {
	cb := New()

	if cb.WToMove != 1 {
		t.Errorf("initial wToMove: want=1, got=%d", cb.WToMove)
	}

	if cb.Pieces[1] != uint64(1<<16)-1 {
		t.Errorf("initial wPieces: want=65535, got=%b", cb.Pieces[1])
	}
	if cb.Pawns[1] != uint64(1<<16)-1-(1<<8-1) {
		t.Errorf("initial wPawns: want=%b, got=%b", 65279, cb.Pawns[1])
	}

	wPiecesUnion := cb.Pawns[1] | cb.Rooks[1] | cb.Knights[1] | cb.Bishops[1] |
		cb.Queens[1] | cb.Kings[1]
	if cb.Pieces[1] != wPiecesUnion {
		t.Errorf("wPieces != union of all white pieces. want=65535,\ngot=%b",
			wPiecesUnion)
	}

	bPieces := uint64(1<<63) - 1 + 1<<63 - (1<<48 - 1)
	if cb.Pieces[0] != bPieces {
		t.Errorf("initial bPieces: want=%b\n, got=%b", bPieces, cb.Pieces[0])
	}
	bPawns := uint64(1<<56) - 1 - (1<<48 - 1)
	if cb.Pawns[0] != bPawns {
		t.Errorf("initial bPawns: want=%b\n, got=%b", bPawns, cb.Pawns[0])
	}

	bPiecesUnion := cb.Pawns[0] | cb.Rooks[0] | cb.Knights[0] |
		cb.Bishops[0] | cb.Queens[0] | cb.Kings[0]
	if cb.Pieces[0] != bPiecesUnion {
		t.Errorf("bPieces != union of all black pieces. want=%b,\ngot=%b",
			uint64(1<<63)-1+(1<<63)-(1<<48-1), bPiecesUnion)
	}
}

type bbTestCase struct {
	square   int
	expected uint64
	actual   uint64
	name     string
}

func runMoveBBTests(t *testing.T, tests []bbTestCase) {
	for _, tt := range tests {
		if tt.actual != tt.expected {
			t.Errorf("incorrect bitboard for %v on square %d.\nexpected:\n%b, %T\ngot:\n%b, %T",
				tt.name, tt.square, tt.expected, tt.expected, tt.actual, tt.actual)
		}
	}
}

func TestFromFen(t *testing.T) {
	cbNew := New()
	valNew := reflect.ValueOf(cbNew).Elem()
	fieldTypeNew := valNew.Type()

	cbFromFen, err := FromFen("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Error(err)
	}
	valFen := reflect.ValueOf(cbFromFen).Elem()
	fieldTypeFen := valFen.Type()

	for i := 0; i < valNew.NumField(); i++ {
		if valNew.Field(i).Interface() != valFen.Field(i).Interface() {
			t.Errorf("fromFen failed:  want %s=%v, got %s=%v\n",
				fieldTypeNew.Field(i).Name, valNew.Field(i).Interface(),
				fieldTypeFen.Field(i).Name, valFen.Field(i).Interface())
		}
	}

	cb, err := FromFen("8/ppppppp1/8/6Pp/6pP/8/PPPPPP2/8 w - a3 0 1")
	if err != nil {
		t.Error(err)
	}
	if cb.EpSquare != 16 {
		t.Errorf("cb.EpSquare: want=16, got=%d\n", cb.EpSquare)
	}

	// Test cb.PiecePhaseSum
	expected := 0
	if cb.PiecePhaseSum != expected {
		t.Errorf("egPhase: want=%d, got=%d", expected, cb.PiecePhaseSum)
	}
	cb, err = FromFen("8/8/8/8/8/8/8/RNBQKBNR w KQkq - 0 1")
	if err != nil {
		t.Error(err)
	}
	expected = 12
	if cb.PiecePhaseSum != expected {
		t.Errorf("egPhase: want=%d, got=%d", expected, cb.PiecePhaseSum)
	}

	// Test errors
	_, err = FromFen("1p7/8/8/8/8/8/8/8 w - - 0 1")
	if err == nil {
		t.Error("the eighth rank is too long; should return error")
	}
	_, err = FromFen("p7/8/8 w - - 0 1")
	if err == nil {
		t.Error("only 2 slashes, not 7; should return error")
	}
}

func TestResetZobrist(t *testing.T) {
	cb, err := FromFen("r3k3/8/8/8/8/8/8/R3K2R w KQq - 0 1")
	if err != nil {
		t.Error(err)
	}
	expected := uint64(0)
	expected ^= ZobristKeys.ColorPieceSq[0][3][56]
	expected ^= ZobristKeys.ColorPieceSq[0][5][60]
	expected ^= ZobristKeys.ColorPieceSq[1][3][0]
	expected ^= ZobristKeys.ColorPieceSq[1][3][7]
	expected ^= ZobristKeys.ColorPieceSq[1][5][4]
	expected ^= ZobristKeys.Castle[0][0]
	expected ^= ZobristKeys.Castle[1][0] ^ ZobristKeys.Castle[1][1]
	if expected != cb.Zobrist {
		t.Errorf("resetZobrist: want=%d, got=%d", expected, cb.Zobrist)
	}
}

func TestResetMidGameEndGamePST(t *testing.T) {
	cb, err := FromFen("rnbqkp2/8/8/8/8/8/RNBQKP2/8 w Qq - 0 1")
	if err != nil {
		t.Error(err)
	}

	expectedEvalMidGamePST := MgTables[3][8^56] + MgTables[1][9^56] +
		MgTables[2][10^56] + MgTables[4][11^56] + MgTables[5][12^56] + MgTables[0][13^56]
	expectedEvalMidGamePST -= MgTables[3][56] + MgTables[1][57] + MgTables[2][58] +
		MgTables[4][59] + MgTables[5][60] + MgTables[0][61]

	expectedEvalEndGamePST := EgTables[3][8^56] + EgTables[1][9^56] +
		EgTables[2][10^56] + EgTables[4][11^56] + EgTables[5][12^56] + EgTables[0][13^56]
	expectedEvalEndGamePST -= EgTables[3][56] + EgTables[1][57] + EgTables[2][58] +
		EgTables[4][59] + EgTables[5][60] + EgTables[0][61]

	if cb.EvalMidGamePST != expectedEvalMidGamePST {
		t.Errorf("EvalMidGamePST: want=%d, got=%d", expectedEvalMidGamePST, cb.EvalMidGamePST)
	}
	if cb.EvalEndGamePST != expectedEvalEndGamePST {
		t.Errorf("EvalEndGamePST: want=%d, got=%d", expectedEvalEndGamePST, cb.EvalEndGamePST)
	}
}
