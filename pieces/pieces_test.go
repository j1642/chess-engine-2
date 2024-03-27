package pieces

import (
	"fmt"
	"github.com/j1642/chess-engine-2/board"
	"github.com/j1642/chess-engine-2/moves"
	"strings"
	"testing"
)

type moveTestCase struct {
	expected uint64
	actual   uint64
}

func runMoveGenTests(t *testing.T, tests []moveTestCase) {
	for _, tt := range tests {
		if tt.expected != tt.actual {
			t.Errorf("want=%v, got=%v", read1Bits(tt.expected), read1Bits(tt.actual))
		}
	}
}

func TestPawnMoves(t *testing.T) {
	cb, err := board.FromFen("8/ppppppp1/P7/6Pp/6pP/8/1PPPPP2/8 w - a3 0 1")
	if err != nil {
		t.Error(err)
	}

	wTests := []moveTestCase{
		{
			// W double push and e.p. square attack
			expected: uint64(1<<16 + 1<<17 + 1<<25),
			actual:   GetPawnMoves(9, cb),
		},
		{
			// W blocked
			expected: uint64(0),
			actual:   GetPawnMoves(31, cb),
		},
		{
			// W partially blocked
			expected: uint64(1 << 23),
			actual:   GetPawnMoves(15, cb),
		},
	}
	cb1, err := board.FromFen("8/3k4/8/8/8/N7/PPPPPPPP/R1BQKBNR w - - 0 1")
	if err != nil {
		t.Error(err)
	}

	wTests = append(wTests, moveTestCase{
		expected: uint64(0),
		actual:   GetPawnMoves(8, cb1),
	})

	cb.WToMove = 0
	bTests := []moveTestCase{
		{
			// B blocked
			expected: uint64(0),
			actual:   GetPawnMoves(39, cb),
		},
		{
			// B partially blocked
			expected: uint64(1 << 46),
			actual:   GetPawnMoves(54, cb),
		},
		{
			// B attack and double push
			expected: uint64(1<<40 + 1<<41 + 1<<33),
			actual:   GetPawnMoves(49, cb),
		},
	}

	runMoveGenTests(t, wTests)
	runMoveGenTests(t, bTests)
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

func TestIsValidMove(t *testing.T) {
	cb := board.New()
	tests := []validMoveTestCase{
		{
			from: -1, to: 0,
			cb:       cb,
			expected: false,
			actual:   IsValidMove(-1, 0, NO_PIECE, cb),
		},
		{
			from: 0, to: 100,
			cb:       cb,
			expected: false,
			actual:   IsValidMove(0, 100, ROOK, cb),
		},
		// No piece present
		{
			from: 20, to: 21,
			cb:       cb,
			expected: false,
			actual:   IsValidMove(20, 21, NO_PIECE, cb),
		},
		// Pawn
		{
			from: 8, to: 24,
			cb:       cb,
			expected: true,
			actual:   IsValidMove(8, 16, PAWN, cb),
		},
		{
			from: 8, to: 8,
			cb:       cb,
			expected: false,
			actual:   IsValidMove(8, 8, PAWN, cb),
		},
		// Knight
		{
			from: 1, to: 18,
			cb:       cb,
			expected: true,
			actual:   IsValidMove(1, 18, KNIGHT, cb),
		},
		{
			from: 6, to: 22,
			cb:       cb,
			expected: false,
			actual:   IsValidMove(6, 22, KNIGHT, cb),
		},
		// Bishop
		{
			from: 2, to: 47,
			cb:       cb,
			expected: false,
			actual:   IsValidMove(2, 47, BISHOP, cb),
		},
		{
			from: 2, to: 11,
			cb:       cb,
			expected: false,
			actual:   IsValidMove(2, 11, BISHOP, cb),
		},
		// Bishop passing edge
		{
			from: 5, to: 32,
			cb:       cb,
			expected: false,
			actual:   IsValidMove(5, 32, BISHOP, cb),
		},
		// Rook
		{
			from: 0, to: 7,
			cb:       cb,
			expected: false,
			actual:   IsValidMove(0, 7, ROOK, cb),
		},
		{
			from: 7, to: 15,
			cb:       cb,
			expected: false,
			actual:   IsValidMove(7, 15, ROOK, cb),
		},
		// Rook/queen cannot wrap around the board edge
		{
			from: 7, to: 8,
			cb:       cb,
			expected: false,
			actual:   IsValidMove(7, 8, ROOK, cb),
		},
		// King
		{
			from: 40, to: 39,
			cb:       cb,
			expected: false,
			actual:   IsValidMove(40, 39, NO_PIECE, cb),
		},
		{
			from: 4, to: 13,
			cb:       cb,
			expected: false,
			actual:   IsValidMove(4, 13, KING, cb),
		},
	}

	runValidMoveTests(t, tests)
}

func TestMovePiece(t *testing.T) {
	cb := board.New()
	MovePiece(board.Move{To: 8, From: 16, Piece: PAWN, PromoteTo: NO_PIECE}, cb)

	if cb.WToMove != 0 {
		t.Errorf("WToMove: want=0, got=%d", cb.WToMove)
	}
	if cb.Pawns[1] != uint64(1<<17)-1-1<<8-(1<<8-1) {
		t.Errorf("wPawns: want=\n%b,\ngot=\n%b\n",
			uint64(1<<17)-1-1<<8-(1<<8-1), cb.Pawns[1])
	}
	if cb.Pieces[1] != uint64(1<<17)-1-1<<8 {
		t.Errorf("wPieces: want=\n%b,\ngot=\n%b",
			uint64(1<<17)-1-1<<8, cb.Pieces[1])
	}
	MovePiece(board.Move{From: 57, To: 42, Piece: KNIGHT, PromoteTo: NO_PIECE}, cb)
	MovePiece(board.Move{From: 16, To: 24, Piece: PAWN, PromoteTo: NO_PIECE}, cb)
	// Waiting move by ng8
	MovePiece(board.Move{From: 62, To: 45, Piece: KNIGHT, PromoteTo: NO_PIECE}, cb)
	MovePiece(board.Move{From: 24, To: 32, Piece: PAWN, PromoteTo: NO_PIECE}, cb)

	MovePiece(board.Move{From: 42, To: 32, Piece: KNIGHT, PromoteTo: NO_PIECE}, cb)

	if cb.WToMove != 1 {
		t.Errorf("WToMove: want=1, got=%d", cb.WToMove)
	}
	if cb.Pawns[1] != uint64(1<<16)-1-1<<8-(1<<8-1) {
		t.Errorf("wPawns: want=\n%b,\ngot=\n%b\n",
			uint64(1<<16)-1-1<<8-(1<<8-1), cb.Pawns[1])
	}
	if cb.Pieces[1] != uint64(1<<16)-1-1<<8 {
		t.Errorf("wPieces: want=\n%b,\ngot=\n%b",
			uint64(1<<16)-1-1<<8, cb.Pieces[1])
	}

	if cb.Knights[0] != uint64(1<<45+1<<32) {
		t.Errorf("bKnights: want=\n%b,\ngot=\n%b",
			uint64(1<<45+1<<32), cb.Knights[0])
	}
	bPieces := ^uint64(0) - (1<<48 - 1) + 1<<32 + 1<<45 - 1<<57 - 1<<62
	if cb.Pieces[0] != bPieces {
		t.Errorf("bPieces: want=\n%b,\ngot=\n%b",
			bPieces, cb.Pieces[0])
	}
}

func TestPromotePawn(t *testing.T) {
	// TODO: mock user input to test other promotePawn() branch
	cb := &board.Board{
		WToMove: 1,
		Pawns:   [2]uint64{1 << 1, 1 << 63},
		Queens:  [2]uint64{0, 0},
	}
	promotePawn(uint64(1<<63), 63, cb, QUEEN)

	if cb.Pawns[1] != uint64(0) {
		t.Errorf("pawn did not promote: want=0, got=%b", cb.Pawns[1])
	}
	if cb.Queens[1] != uint64(1<<63) {
		t.Error("promoted queen not present")
	}
}

func TestGetAttackedSquare(t *testing.T) {
	cb := board.New()

	expected := uint64(0xFFFF00) + 1<<1 + 1<<2 + 1<<3 + 1<<4 + 1<<5 + 1<<6
	if GetAttackedSquares(cb) != expected {
		t.Errorf("attacked/defended squares: want=%v, got=%v",
			read1Bits(expected), read1Bits(GetAttackedSquares(cb)))
	}
}

func TestCastling(t *testing.T) {
	// Castling moves king and rook, and removes remaining castling rights.
	cb, err := board.FromFen("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	if err != nil {
		t.Error(err)
	}

	MovePiece(board.Move{From: 4, To: 2, Piece: KING, PromoteTo: NO_PIECE}, cb)
	if cb.Kings[1] != uint64(1<<2) {
		t.Errorf("w king did not castle queenside. want=2, got=%v", read1Bits(cb.Kings[1]))
	}
	if cb.Rooks[1] != uint64(1<<3+1<<7) {
		t.Errorf("rook did not move for castling. want=[3 7], got=%v", read1Bits(cb.Rooks[1]))
	}
	if cb.CastleRights[1] != [2]bool{false, false} {
		t.Errorf("w king castle rights: want=[false false], got=%v", cb.CastleRights[1])
	}

	MovePiece(board.Move{From: 60, To: 62, Piece: KING, PromoteTo: NO_PIECE}, cb)
	if cb.Kings[0] != uint64(1<<62) {
		t.Errorf("b king did not castle kingside. want=62, got=%v", read1Bits(cb.Kings[0]))
	}
	if cb.Rooks[0] != uint64(1<<56+1<<61) {
		t.Errorf("rook did not move for castling. want=[56 61], got=%v", read1Bits(cb.Rooks[0]))
	}
	if cb.CastleRights[0] != [2]bool{false, false} {
		t.Errorf("b king castle rights: want=[false false], got=%v", cb.CastleRights[0])
	}

}

func TestCastlingRightsLostByRookMoveAndCapture(t *testing.T) {
	cb, err := board.FromFen("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	if err != nil {
		t.Error(err)
	}

	MovePiece(board.Move{From: 0, To: 56, Piece: ROOK, PromoteTo: NO_PIECE}, cb)
	if cb.Rooks[1] != uint64(1<<7+1<<56) {
		t.Errorf("wrong rook squares: want=[7, 56], got=%v", read1Bits(cb.Rooks[1]))
	}
	if cb.CastleRights[1] != [2]bool{false, true} {
		t.Errorf("w king castle rights: want=[false true], got=%v", cb.CastleRights[1])
	}
	if cb.CastleRights[0] != [2]bool{false, true} {
		t.Errorf("b king castle rights: want=[false true], got=%v", cb.CastleRights[0])
	}

	MovePiece(board.Move{From: 63, To: 7, Piece: ROOK, PromoteTo: NO_PIECE}, cb)
	if cb.Rooks[0] != uint64(1<<7) {
		t.Errorf("wrong rook squares: want=7, got=%v", read1Bits(cb.Rooks[0]))
	}
	if cb.CastleRights[1] != [2]bool{false, false} {
		t.Errorf("w king castle rights: want=[false false], got=%v", cb.CastleRights[1])
	}
	if cb.CastleRights[0] != [2]bool{false, false} {
		t.Errorf("b king castle rights: want=[false false], got=%v", cb.CastleRights[0])
	}
}

func TestGetKingMoves(t *testing.T) {
	cb1, err := board.FromFen("3r4/8/8/8/8/8/8/R3KR2 w Q - 0 1")
	if err != nil {
		t.Error(err)
	}
	cb1.Pieces[cb1.WToMove] ^= uint64(1 << cb1.KingSqs[cb1.WToMove])
	cb1.WToMove ^= 1
	attkSquares := GetAttackedSquares(cb1)
	cb1.WToMove ^= 1
	cb1.Pieces[cb1.WToMove] ^= uint64(1 << cb1.KingSqs[cb1.WToMove])

	tests := []moveTestCase{
		{
			// Cannot move into check, castle through check, or move into friendly piece.
			expected: uint64(1<<12 + 1<<13),
			actual:   GetKingMoves(cb1.KingSqs[1], attkSquares, cb1),
		},
	}

	cb2, err := board.FromFen("r3k2r/4R3/3P4/8/8/8/8/8 b Q - 0 1")
	if err != nil {
		t.Error(err)
	}
	cb2.Pieces[cb2.WToMove] ^= uint64(1 << cb2.KingSqs[cb2.WToMove])
	cb2.WToMove ^= 1
	attkSquares = GetAttackedSquares(cb2)
	cb2.WToMove ^= 1
	cb2.Pieces[cb2.WToMove] ^= uint64(1 << cb2.KingSqs[cb2.WToMove])

	tests = append(tests, moveTestCase{
		// Cannot castle out of check or capture protected piece.
		expected: uint64(1<<61 + 1<<59),
		actual:   GetKingMoves(cb2.KingSqs[0], attkSquares, cb2),
	})

	runMoveGenTests(t, tests)
}

type allMovesTestCase struct {
	expected, actual []board.Move
}

func runGetAllMovesTests(t *testing.T, tests []allMovesTestCase) {
	for _, tt := range tests {
		if len(tt.expected) != len(tt.actual) {
			t.Logf("expected: %v\n", tt.expected)
			t.Logf("actual: %v\n", tt.actual)
			t.Errorf("wrong length: want=%d (%v), got=%d (%v)",
				len(tt.expected), tt.expected, len(tt.actual), tt.actual)
		}
		for i, move := range tt.expected {
			if move != tt.actual[i] {
				t.Errorf("allMoves[%d]: want=%v, got=%v", i, tt.expected[i], tt.actual[i])
			}
		}
	}
}

func TestFillFromTo(t *testing.T) {
	expected := uint64(1 << 8)
	actual := fillFromTo(0, 8, 8)
	if expected != actual {
		t.Errorf("fillFromTo: want=[8], got=%v", read1Bits(actual))
	}
	expected = uint64(1<<8 + 1<<16)
	actual = fillFromTo(0, 16, 8)
	if expected != actual {
		t.Errorf("fillFromTo: want=[8 16], got=%v", read1Bits(actual))
	}
	expected = uint64(1<<53 + 1<<46 + 1<<39)
	actual = fillFromTo(60, 39, -7)
	if expected != actual {
		t.Errorf("fillFromTo: want=[39 46 53], got=%v", read1Bits(actual))
	}
}

type intIntTestCase struct {
	expected, actual int8
}

func TestFindDirection(t *testing.T) {
	tests := []intIntTestCase{
		// Roughly clockwise
		{
			expected: 8,
			actual:   findDirection(9, 17),
		},
		{
			expected: 9,
			actual:   findDirection(9, 18),
		},
		{
			expected: 1,
			actual:   findDirection(9, 10),
		},
		{
			expected: 1,
			actual:   findDirection(0, 7),
		},
		{
			expected: -1,
			actual:   findDirection(7, 0),
		},
		{
			expected: -7,
			actual:   findDirection(9, 2),
		},
		{
			expected: -8,
			actual:   findDirection(9, 1),
		},
		{
			expected: -9,
			actual:   findDirection(9, 0),
		},
		{
			expected: -1,
			actual:   findDirection(9, 8),
		},
		{
			expected: 7,
			actual:   findDirection(9, 16),
		},
	}

	for _, tt := range tests {
		if tt.expected != tt.actual {
			t.Errorf("findDirection: want=%d, got=%d", tt.expected, tt.actual)
		}
	}
}

type checkingSquaresCase struct {
	cb                            *board.Board
	expCapsBlks, actualCapsBlks   uint64
	expAttkCount, actualAttkCount int
}

func TestGetCheckingSquares(t *testing.T) {
	cb, err := board.FromFen("R5rR/8/8/8/8/8/8/RNBQ2K1 w - - 0 1")
	if err != nil {
		t.Error(err)
	}
	capturesBlockers, attackerCount := GetCheckingSquares(cb)
	tests := []checkingSquaresCase{
		{
			cb:             cb,
			expCapsBlks:    uint64(1<<62 + 1<<54 + 1<<46 + 1<<38 + 1<<30 + 1<<22 + 1<<14),
			actualCapsBlks: capturesBlockers,
			expAttkCount:   1, actualAttkCount: attackerCount,
		},
	}

	cb, err = board.FromFen("4k3/pppppppp/8/8/6n1/5P2/PPPPPKPP/5BNR w - - 0 1")
	if err != nil {
		t.Error(err)
	}
	capturesBlockers, attackerCount = GetCheckingSquares(cb)
	tests = append(tests, checkingSquaresCase{
		cb:             cb,
		expCapsBlks:    uint64(1 << 30),
		actualCapsBlks: capturesBlockers,
		expAttkCount:   1, actualAttkCount: attackerCount,
	})

	// Black pawn on 13 does not check white king on 20.
	cb, err = board.FromFen("n1n5/PPPk4/8/8/8/4K2q/5p1p/7N w - - 2 3")
	if err != nil {
		t.Error(err)
	}
	capturesBlockers, attackerCount = GetCheckingSquares(cb)
	tests = append(tests, checkingSquaresCase{
		cb:             cb,
		expCapsBlks:    uint64(1<<21 + 1<<22 + 1<<23),
		actualCapsBlks: capturesBlockers,
		expAttkCount:   1, actualAttkCount: attackerCount,
	})

	runGetCheckingSquaresTests(t, tests)
}

func runGetCheckingSquaresTests(t *testing.T, tests []checkingSquaresCase) {
	for _, tt := range tests {
		if tt.expAttkCount != tt.actualAttkCount {
			t.Errorf("wrong attackerCount: want=%d, got=%d",
				tt.expAttkCount, tt.actualAttkCount)
		}
		if tt.expCapsBlks != tt.actualCapsBlks {
			t.Errorf("wrong capturesBlockers: want=%v, got=%v",
				read1Bits(tt.expCapsBlks), read1Bits(tt.actualCapsBlks))
		}
	}
}

func TestGetAllMoves(t *testing.T) {
	cb, err := board.FromFen("R5rR/8/8/8/8/8/8/RNBQ2K1 w - - 0 1")
	if err != nil {
		t.Error(err)
	}
	tests := []allMovesTestCase{
		{
			// One checking piece which can be captured or blocked.
			expected: []board.Move{{From: 6, To: 5, Piece: KING, PromoteTo: NO_PIECE},
				{From: 6, To: 7, Piece: KING, PromoteTo: NO_PIECE},
				{From: 6, To: 13, Piece: KING, PromoteTo: NO_PIECE},
				{From: 6, To: 15, Piece: KING, PromoteTo: NO_PIECE},
				{From: 2, To: 38, Piece: BISHOP, PromoteTo: NO_PIECE},
				{From: 56, To: 62, Piece: ROOK, PromoteTo: NO_PIECE},
				{From: 63, To: 62, Piece: ROOK, PromoteTo: NO_PIECE},
				{From: 3, To: 30, Piece: QUEEN, PromoteTo: NO_PIECE}},
			actual: GetAllMoves(cb),
		},
	}

	cb1, err := board.FromFen("R5rR/8/8/8/8/8/5b2/RNBQ2K1 w - - 0 1")
	if err != nil {
		t.Error(err)
	}
	tests = append(tests, allMovesTestCase{
		// Two checking pieces, so only the king can move.
		expected: []board.Move{{From: 6, To: 5, Piece: KING, PromoteTo: NO_PIECE},
			{From: 6, To: 7, Piece: KING, PromoteTo: NO_PIECE},
			{From: 6, To: 13, Piece: KING, PromoteTo: NO_PIECE},
			{From: 6, To: 15, Piece: KING, PromoteTo: NO_PIECE}},
		actual: GetAllMoves(cb1),
	})

	cb2, err := board.FromFen("rnbqkbnr/ppppp1pp/5p2/7Q/8/4P3/PPPP1PPP/RNB1KBNR b KQkq - 0 1")
	if err != nil {
		t.Error(err)
	}
	tests = append(tests, allMovesTestCase{
		// Only one move is possible: pawn blocks check.
		expected: []board.Move{{From: 54, To: 46, Piece: PAWN, PromoteTo: NO_PIECE}},
		actual:   GetAllMoves(cb2),
	})

	cb3, err := board.FromFen("4k3/pppppppp/8/8/6n1/5P2/PPPPPKPP/5BNR w - - 0 1")
	if err != nil {
		t.Error(err)
	}
	tests = append(tests, allMovesTestCase{
		expected: []board.Move{{From: 13, To: 4, Piece: KING, PromoteTo: NO_PIECE},
			{From: 13, To: 22, Piece: KING, PromoteTo: NO_PIECE},
			{From: 21, To: 30, Piece: PAWN, PromoteTo: NO_PIECE}},
		actual: GetAllMoves(cb3),
	})

	cb4, err := board.FromFen("n1n5/PPPk4/8/8/8/4K2q/5p1p/7N w - - 2 3")
	if err != nil {
		t.Error(err)
	}
	tests = append(tests, allMovesTestCase{
		expected: []board.Move{{From: 20, To: 11, Piece: KING, PromoteTo: NO_PIECE},
			{From: 20, To: 12, Piece: KING, PromoteTo: NO_PIECE},
			{From: 20, To: 13, Piece: KING, PromoteTo: NO_PIECE},
			{From: 20, To: 27, Piece: KING, PromoteTo: NO_PIECE},
			{From: 20, To: 28, Piece: KING, PromoteTo: NO_PIECE},
			{From: 20, To: 29, Piece: KING, PromoteTo: NO_PIECE},
			{From: 7, To: 22, Piece: KNIGHT, PromoteTo: NO_PIECE}},
		actual: GetAllMoves(cb4),
	})

	runGetAllMovesTests(t, tests)
}

func TestRead1Bits(t *testing.T) {
	nums := read1Bits(uint64(0b11001))
	expected := []int8{0, 3, 4}

	for i, num := range nums {
		if num != expected[i] {
			t.Errorf("failed to read 1 bits: want=%d, got=%d", expected[i], nums[i])
		}
	}
}

func perft(depth int, cb *board.Board) int {
	if depth == 0 {
		return 1
	}
	nodes := 0
	moves := GetAllMoves(cb)
	pos := board.StorePosition(cb)

	for _, toFrom := range moves {
		MovePiece(toFrom, cb)
		if toFrom.Piece == KING || cb.Kings[1^cb.WToMove]&GetAttackedSquares(cb) == 0 {
			nodes += perft(depth-1, cb)
		}
		board.RestorePosition(pos, cb)
	}

	return nodes
}

func divide(depth int, cb *board.Board) {
	ranks := []string{"1", "2", "3", "4", "5", "6", "7", "8"}
	files := []string{"a", "b", "c", "d", "e", "f", "g", "h"}

	totalNodes := 0
	moves := GetAllMoves(cb)
	var pos *board.Position

	pos = board.StorePosition(cb)
	for _, fromTo := range moves {
		nodes := 0
		MovePiece(fromTo, cb)
		attackedSquares := GetAttackedSquares(cb)
		if cb.Kings[1^cb.WToMove]&attackedSquares == 0 {
			nodes += perft(depth-1, cb)
		}
		board.RestorePosition(pos, cb)

		fromAlgNotation := strings.Join([]string{files[fromTo.From%8], ranks[fromTo.From/8]}, "")
		toAlgNotation := strings.Join([]string{files[fromTo.To%8], ranks[fromTo.To/8]}, "")

		fmt.Printf("%s%s %v: %d\n",
			fromAlgNotation, toAlgNotation, fromTo.PromoteTo, nodes)
		totalNodes += nodes

	}
	fmt.Println("Total nodes:", totalNodes)
}

type perftTestCase struct {
	name                    string
	expected, actual, depth int
}

func TestPerft(t *testing.T) {
	cb := board.New()
	promoteCb, err := board.FromFen("n1n5/PPPk4/8/8/8/8/4Kppp/5N1N b - - 0 1")
	if err != nil {
		t.Error(err)
	}
	kiwipeteCb, err := board.FromFen("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	if err != nil {
		t.Error(err)
	}
	tests := []perftTestCase{
		{
			name:     "perft",
			depth:    5,
			expected: 4_865_609,
			//expected: 119_060_324,
			actual: perft(5, cb),
		},
		{
			name:     "promotePerft",
			depth:    5,
			expected: 3_605_103,
			actual:   perft(5, promoteCb),
		},
		{
			name:     "kiwipete",
			depth:    4,
			expected: 4085603,
			actual:   perft(4, kiwipeteCb),
		},
	}

	runPerftTests(t, tests)
}

func runPerftTests(t *testing.T, tests []perftTestCase) {
	for _, tt := range tests {
		if tt.expected != tt.actual {
			t.Errorf("%s(%d): want=%d, got=%d",
				tt.name, tt.depth, tt.expected, tt.actual)
		}
	}
}

type slidingMovesLookupTestCase struct {
	calcFunc      func(uint64, *board.Board) uint64
	lookupFunc    func(uint64, *board.Board) uint64
	piece_squares []int
}

func TestSlidingPiecesMovesLookup(t *testing.T) {
	// Compare all piece's calculated moves with lookup, magic bitboard moves
	// lookupRookMoves() depends on *board.Board.WToMove accuracy
	kiwipete, err := board.FromFen("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	if err != nil {
		t.Error(err)
	}

	rooks := kiwipete.Rooks[0] | kiwipete.Rooks[1]
	rook_squares := read1Bits(rooks)
	for _, square := range rook_squares {
		kiwipete.WToMove = 1
		if (1<<square)&kiwipete.Pieces[0] != 0 {
			kiwipete.WToMove = 0
		}
		calcMoves := moves.CalculateRookMoves(int(square), kiwipete)
		lookupMoves := lookupRookMoves(square, kiwipete)
		if calcMoves != lookupMoves {
			t.Errorf("for a rook on sq %d, calculated moves != magic BB moves, calc=%v, lookup=%v",
				square, read1Bits(calcMoves), read1Bits(lookupMoves))
		}
	}

	bishops := kiwipete.Bishops[0] | kiwipete.Bishops[1]
	bishop_squares := read1Bits(bishops)
	for _, square := range bishop_squares {
		kiwipete.WToMove = 1
		if (1<<square)&kiwipete.Pieces[0] != 0 {
			kiwipete.WToMove = 0
		}
		calcMoves := moves.CalculateBishopMoves(int(square), kiwipete)
		lookupMoves := lookupBishopMoves(square, kiwipete)
		if calcMoves != lookupMoves {
			t.Errorf("for a bishop on sq %d, calculated moves != magic BB moves, calc=%v, lookup=%v",
				square, read1Bits(calcMoves), read1Bits(lookupMoves))
		}
	}
}

func TestZobristHashing(t *testing.T) {
	// Moves, captures, promotions, castling, the color to move, and any en
	// passant square affect the position hash
	cb, err := board.FromFen("r3k2r/1Pp5/8/8/8/8/8/R3K2R w KQkq - 0 1")
	if err != nil {
		t.Error(err)
	}
	curZobrist := cb.Zobrist

	MovePiece(board.Move{From: 0, To: 56, Piece: ROOK, PromoteTo: NO_PIECE}, cb)
	curZobrist ^= board.ZobristKeys.ColorPieceSq[1][3][0]
	curZobrist ^= board.ZobristKeys.ColorPieceSq[1][3][56]
	curZobrist ^= board.ZobristKeys.ColorPieceSq[0][3][56]
	curZobrist ^= board.ZobristKeys.BToMove
	curZobrist ^= board.ZobristKeys.Castle[0][0] ^ board.ZobristKeys.Castle[1][0]
	if curZobrist != cb.Zobrist {
		t.Errorf("zobrist: want=%d, got=%d", curZobrist, cb.Zobrist)
	}

	MovePiece(board.Move{From: 63, To: 7, Piece: ROOK, PromoteTo: NO_PIECE}, cb)
	curZobrist ^= board.ZobristKeys.ColorPieceSq[0][3][63]
	curZobrist ^= board.ZobristKeys.ColorPieceSq[0][3][7]
	curZobrist ^= board.ZobristKeys.ColorPieceSq[1][3][7]
	curZobrist ^= board.ZobristKeys.BToMove
	curZobrist ^= board.ZobristKeys.Castle[0][1] ^ board.ZobristKeys.Castle[1][1]
	if curZobrist != cb.Zobrist {
		t.Errorf("zobrist: want=%d, got=%d", curZobrist, cb.Zobrist)
	}

	MovePiece(board.Move{From: 49, To: 57, Piece: PAWN, PromoteTo: QUEEN}, cb)
	curZobrist ^= board.ZobristKeys.ColorPieceSq[1][0][49]
	curZobrist ^= board.ZobristKeys.ColorPieceSq[1][4][57]
	curZobrist ^= board.ZobristKeys.BToMove
	if curZobrist != cb.Zobrist {
		t.Errorf("zobrist: want=%d, got=%d", curZobrist, cb.Zobrist)
	}

	MovePiece(board.Move{From: 50, To: 34, Piece: PAWN, PromoteTo: NO_PIECE}, cb)
	curZobrist ^= board.ZobristKeys.ColorPieceSq[0][0][50]
	curZobrist ^= board.ZobristKeys.ColorPieceSq[0][0][34]
	curZobrist ^= board.ZobristKeys.BToMove
	curZobrist ^= board.ZobristKeys.EpFile[cb.EpSquare%8]
	if curZobrist != cb.Zobrist {
		t.Errorf("zobrist: want=%d, got=%d", curZobrist, cb.Zobrist)
	}
}
