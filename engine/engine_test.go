package engine

import (
	"github.com/j1642/chess-engine-2/board"
	"github.com/j1642/chess-engine-2/pieces"
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
	whiteMatesBlack, err := board.FromFen("8/8/8/8/8/5K2/6Q1/7k b - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	blackMatesWhite, err := board.FromFen("8/8/8/8/8/5k2/6q1/7K w - - 0 1")
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
			expected: -59,
		},
		{
			cb:       cbLoneRook,
			expected: 635,
		},
		{
			cb:       whiteMatesBlack,
			expected: -MATE,
		},
		{
			cb:       blackMatesWhite,
			expected: -MATE,
		},
	}

	runEvalTests(t, tests)
}

func runEvalTests(t *testing.T, tests []evalTestCase) {
	for i, tt := range tests {
		actual := evaluate(tt.cb)
		if tt.expected != actual {
			t.Errorf("eval[%d]: want=%d, got=%d", i, tt.expected, actual)
		}
	}
}

func TestEvalPawns(t *testing.T) {
	cbDoubled, err := board.FromFen("8/1pp5/8/2p5/5P2/8/5PP1/8 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	cbIsolated, err := board.FromFen("8/p1p5/8/8/8/8/1P4PP/8 w - - 0 1")
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
			cb: cbDoubled,
			expected: -eg_tables[0][49] - eg_tables[0][50] - eg_tables[0][34] +
				eg_tables[0][13^56] + eg_tables[0][14^56] + eg_tables[0][29^56],
		},
		{
			cb: cbIsolated,
			expected: 150 - eg_tables[0][48] - eg_tables[0][50] + eg_tables[0][9^56] +
				eg_tables[0][14^56] + eg_tables[0][15^56],
		},
		{
			cb: cbBlocked,
			expected: 50 - eg_tables[0][48] - eg_tables[0][49] +
				eg_tables[0][14^56] + eg_tables[0][15^56],
		},
	}

	runPawnEvalTests(t, tests)
}

func runPawnEvalTests(t *testing.T, tests []evalTestCase) {
	for i, tt := range tests {
		actual := evalPawns(tt.cb, 0, MAX_PHASE)
		if tt.expected != actual {
			t.Errorf("pawnEval[%d]: want=%d, got=%d", i, tt.expected, actual)
		}
	}
}

type searchTestCase struct {
	cb         *board.Board
	expectEval int
	expectMove board.Move
	depth      int
}

func TestNegamax(t *testing.T) {
	wRookCapturesBRook, err := board.FromFen("8/8/8/8/8/8/8/Rr6 w - - 0 1")
	if err != nil {
		t.Error(err)
	}
	bRookCapturesWRook, err := board.FromFen("8/8/8/8/8/8/8/rR6 b - - 0 1")
	if err != nil {
		t.Error(err)
	}
	mateDepth0, err := board.FromFen("r2qk2r/pb4p1/1n2PbB1/2B5/p1p5/2P5/5PPP/RN2R1K1 b - - 1 0")
	if err != nil {
		t.Error(err)
	}
	mateDepth1, err := board.FromFen("r2qk2r/pb4p1/1n2PbB1/2B5/p1p5/2P5/5PPP/RN2R1K1 b - - 1 0")
	if err != nil {
		t.Error(err)
	}
	mateIn2Ply, err := board.FromFen("r2qk2r/pb4p1/1n2Pbp1/2B5/p1p5/2P5/2B2PPP/RN2R1K1 w - - 1 0")
	if err != nil {
		t.Error(err)
	}
	mateIn3Ply, err := board.FromFen("r2qk2r/pb4pp/1n2PbQ1/2B5/p1p5/2P5/2B2PPP/RN2R1K1 b - - 1 0")
	if err != nil {
		t.Error(err)
	}
	mateIn4Ply, err := board.FromFen("r2qk2r/pb4pp/1n2Pb2/2B2Q2/p1p5/2P5/2B2PPP/RN2R1K1 w - - 1 0")
	if err != nil {
		t.Error(err)
	}

	// TODO: Some of the expectEval might be wrong
	// mate is detected when the side to move cannot move, so the depth arg needs an extra ply
	tests := []searchTestCase{
		{cb: wRookCapturesBRook, expectEval: 644, expectMove: board.Move{From: 0, To: 1, Piece: pieces.ROOK, PromoteTo: pieces.NO_PIECE}, depth: 1},
		{cb: bRookCapturesWRook, expectEval: 608, expectMove: board.Move{From: 0, To: 1, Piece: pieces.ROOK, PromoteTo: pieces.NO_PIECE}, depth: 1},
		{cb: mateDepth0, expectEval: -MATE, expectMove: board.Move{From: 0, To: 0, Piece: 0, PromoteTo: 0}, depth: 0},
		{cb: mateDepth1, expectEval: -MATE, expectMove: board.Move{From: 0, To: 0, Piece: 0, PromoteTo: 0}, depth: 1},
		{cb: mateIn2Ply, expectEval: MATE, expectMove: board.Move{From: 10, To: 46, Piece: pieces.BISHOP, PromoteTo: pieces.NO_PIECE}, depth: 2},
		{cb: mateIn3Ply, expectEval: -MATE, expectMove: board.Move{From: 55, To: 46, Piece: pieces.PAWN, PromoteTo: pieces.NO_PIECE}, depth: 3},
		{cb: mateIn4Ply, expectEval: MATE, expectMove: board.Move{From: 37, To: 46, Piece: pieces.QUEEN, PromoteTo: pieces.NO_PIECE}, depth: 4},
	}

	for i, tt := range tests {
		// Unexplained bug: using math.MinInt, math.MaxInt as args breaks negamax
		//line := make([]board.Move, 0, tt.depth)
		line := make([]board.Move, tt.depth)
		completePVLine := pvLine{}
		completePVLine.alreadyUsed = make([]bool, tt.depth)

		eval, actualMove := negamax(-(1 << 30), 1<<30, tt.depth, tt.cb, tt.depth, tt.cb.HalfMoves, &line, &completePVLine)

		if actualMove != tt.expectMove {
			t.Errorf("negamax best move[%d]: want=%v, got=%v, eval=%d",
				i, tt.expectMove, actualMove, eval)
		}
		if eval != tt.expectEval {
			t.Errorf("negamax eval[%d]: want=%d, got=%d", i, tt.expectEval, eval)
		}
	}
}

func TestIterativeDeepening(t *testing.T) {
	kiwipete1, err := board.FromFen("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	if err != nil {
		t.Error(err)
	}
	kiwipete2, err := board.FromFen("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	if err != nil {
		t.Error(err)
	}

	// TODO: stockfish depth 30 takes at least 5 seconds, this takes less than 1
	// without any depth check alongside the zobrist check. If both checks exist,
	// depth 3 takes .8 sec, depth 4 takes >15 sec
	depth := 2
	eval1, move1 := IterativeDeepening(kiwipete1, depth)

	line := make([]board.Move, depth)
	completePVLine := pvLine{}
	completePVLine.alreadyUsed = make([]bool, depth)

	eval2, move2 := negamax(-(1 << 30), 1<<30, depth, kiwipete2, depth, kiwipete2.HalfMoves, &line, &completePVLine)

	emptyMove := board.Move{}
	if move1 == emptyMove {
		t.Errorf("iter deep returned an empty Move")
	}
	if move1 != move2 {
		t.Errorf("iter deep failed: %v != %v", move1, move2)
	}
	if eval1 != eval2 {
		t.Errorf("iter deep failed: %d != %d", eval1, eval2)
	}
}

func TestQuiesce(t *testing.T) {
	rooksKings, err := board.FromFen("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	if err != nil {
		t.Error(err)
	}
	eval := quiesce(-(1 << 30), 1<<30, rooksKings)
	expected := 727
	if eval != expected {
		t.Errorf("want=%d, got=%d", expected, eval)
	}
}

func TestConvertMovesToLongAlgebraic(t *testing.T) {
	cb, err := board.FromFen("N7/1P6/8/8/8/8/8/8 w - - 0 1")
	if err != nil {
		t.Error(err)
	}
	boardMoves := pieces.GetAllMoves(cb)
	actual := convertMovesToLongAlgebraic(boardMoves)
	expected := []string{"b7b8n", "b7b8b", "b7b8r", "b7b8q", "a8b6", "a8c7"}
	for i, actualAlgMove := range actual {
		if expected[i] != actualAlgMove {
			t.Errorf("algMoves[%d]: %s != %s", i, expected[i], actualAlgMove)
		}
	}
}

type phaseTestCase struct {
	cb                  *board.Board
	expectedPhase       int
	expectedEval        int
	expectedPieceCounts [2][4]int
}

func TestEvalPieceSquareTables(t *testing.T) {
	oneRook, err := board.FromFen("r7/8/8/8/8/8/8/8 b - - 0 1")
	if err != nil {
		t.Error(err)
	}

	oneOfEachPiece, err := board.FromFen("8/rnbqk3/8/8/8/RNBQK3/8/8 w - - 0 1")
	if err != nil {
		t.Error(err)
	}
	egPhase := 8 * MAX_PHASE / 24
	mgPhase := MAX_PHASE - egPhase
	expEval := -(mgPhase*mg_tables[3][48] + egPhase*eg_tables[3][48]) / MAX_PHASE
	expEval -= (mgPhase*mg_tables[1][49] + egPhase*eg_tables[1][49]) / MAX_PHASE
	expEval -= (mgPhase*mg_tables[2][50] + egPhase*eg_tables[2][50]) / MAX_PHASE
	expEval -= (mgPhase*mg_tables[4][51] + egPhase*eg_tables[4][51]) / MAX_PHASE
	expEval -= (mgPhase*mg_tables[5][52] + egPhase*eg_tables[5][52]) / MAX_PHASE

	expEval += (mgPhase*mg_tables[3][16^56] + egPhase*eg_tables[3][16^56]) / MAX_PHASE
	expEval += (mgPhase*mg_tables[1][17^56] + egPhase*eg_tables[1][17^56]) / MAX_PHASE
	expEval += (mgPhase*mg_tables[2][18^56] + egPhase*eg_tables[2][18^56]) / MAX_PHASE
	expEval += (mgPhase*mg_tables[4][19^56] + egPhase*eg_tables[4][19^56]) / MAX_PHASE
	expEval += (mgPhase*mg_tables[5][20^56] + egPhase*eg_tables[5][20^56]) / MAX_PHASE

	tests := []phaseTestCase{
		{
			// start position knights, bishops, rook, queen count
			cb:                  board.New(),
			expectedPhase:       0,
			expectedPieceCounts: [2][4]int{{2, 2, 2, 1}, {2, 2, 2, 1}},
			expectedEval:        0,
		},
		{
			cb:                  oneRook,
			expectedPhase:       22 * MAX_PHASE / 24,
			expectedPieceCounts: [2][4]int{{0, 0, 1, 0}, {0, 0, 0, 0}},
			expectedEval:        eg_tables[3][0] - 1 + eg_tables[5][0^56] - eg_tables[5][0],
		},
		{
			// no pieces left, endgame phase is at its maximum
			cb:                  &board.Board{},
			expectedPhase:       MAX_PHASE,
			expectedPieceCounts: [2][4]int{{0, 0, 0, 0}, {0, 0, 0, 0}},
			expectedEval:        eg_tables[5][0^56] - eg_tables[5][0],
		},
		{
			cb:                  oneOfEachPiece,
			expectedPhase:       egPhase,
			expectedPieceCounts: [2][4]int{{1, 1, 1, 1}, {1, 1, 1, 1}},
			expectedEval:        expEval,
		},
	}

	for i, tt := range tests {
		eval, actualEgPhase, pieceCounts := evalPieceSquareTables(tt.cb)
		if tt.expectedPhase != actualEgPhase {
			t.Errorf("phase[%d]: want=%d, got=%d", i, tt.expectedPhase, actualEgPhase)
		}
		if tt.expectedPieceCounts != pieceCounts {
			t.Errorf("pieceCounts[%d]: want=%d, got=%d", i, tt.expectedPieceCounts, pieceCounts)
		}
		if tt.expectedEval != eval {
			t.Errorf("pstEval[%d]: want=%d, got=%d", i, tt.expectedEval, eval)
		}
	}
}

/*
func TestDisastrousBestMove(t *testing.T) {
	cb, err := board.FromFen("rnb1k1nr/pppp1ppp/8/4P3/1q6/2N5/PP2PPPP/R1BQKBNR b KQkq - 1 5")
	if err != nil {
		t.Error(err)
	}
	_, move := IterativeDeepening(cb, 4)
	if move.From == 25 && move.To == 9 && move.Piece == pieces.QUEEN {
		t.Error("hanging queen after pawn capture")
        cb.Print()
	} else {
		t.Errorf("did not hang queen: from,to,piece = %d,%d,%d", move.From, move.To, move.Piece)
	}
}
*/
