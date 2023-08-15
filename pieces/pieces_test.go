package pieces

import (
	"engine2/board"
	"fmt"
	"strings"
	"testing"
	_ "time"
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
			actual:   getPawnMoves(9, cb),
		},
		{
			// W blocked
			expected: uint64(0),
			actual:   getPawnMoves(31, cb),
		},
		{
			// W partially blocked
			expected: uint64(1 << 23),
			actual:   getPawnMoves(15, cb),
		},
	}
	cb1, err := board.FromFen("8/3k4/8/8/8/N7/PPPPPPPP/R1BQKBNR w - - 0 1")
	if err != nil {
		t.Error(err)
	}

	wTests = append(wTests, moveTestCase{
		expected: uint64(0),
		actual:   getPawnMoves(8, cb1),
	})

	cb.WToMove = 0
	bTests := []moveTestCase{
		{
			// B blocked
			expected: uint64(0),
			actual:   getPawnMoves(39, cb),
		},
		{
			// B partially blocked
			expected: uint64(1 << 46),
			actual:   getPawnMoves(54, cb),
		},
		{
			// B attack and double push
			expected: uint64(1<<40 + 1<<41 + 1<<33),
			actual:   getPawnMoves(49, cb),
		},
	}

	runMoveGenTests(t, wTests)
	runMoveGenTests(t, bTests)
}

func TestRookMoves(t *testing.T) {
	cb := board.New()
	tests := []moveTestCase{
		{
			expected: uint64(1<<1 + 1<<8),
			actual:   getRookMoves(0, cb),
		},
		{
			expected: uint64(1<<62 + 1<<55),
			actual:   getRookMoves(63, cb),
		},
		{
			expected: uint64(1<<24) - 1 - (1<<16 - 1) - 1<<20 +
				1<<12 + 1<<28 + 1<<36 + 1<<44 + 1<<52,
			actual: getRookMoves(20, cb),
		},
	}

	runMoveGenTests(t, tests)
}

func TestBishopMoves(t *testing.T) {
	cb := board.New()
	tests := []moveTestCase{
		{
			expected: uint64(1 << 9),
			actual:   getBishopMoves(0, cb),
		},
		{
			expected: uint64(1 + 1<<16 + 1<<2 + 1<<18 + 1<<27 + 1<<36 + 1<<45 + 1<<54),
			actual:   getBishopMoves(9, cb),
		},
		{
			expected: uint64(1<<10 + 1<<19 + 1<<37 + 1<<46 + 1<<55 +
				1<<14 + 1<<21 + 1<<35 + 1<<42 + 1<<49),
			actual: getBishopMoves(28, cb),
		},
	}

	runMoveGenTests(t, tests)
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
			actual:   isValidMove(-1, 0, "", cb),
		},
		{
			from: 0, to: 100,
			cb:       cb,
			expected: false,
			actual:   isValidMove(0, 100, "r", cb),
		},
		// No piece present
		{
			from: 20, to: 21,
			cb:       cb,
			expected: false,
			actual:   isValidMove(20, 21, "", cb),
		},
		// Pawn
		{
			from: 8, to: 24,
			cb:       cb,
			expected: true,
			actual:   isValidMove(8, 16, "p", cb),
		},
		{
			from: 8, to: 8,
			cb:       cb,
			expected: false,
			actual:   isValidMove(8, 8, "p", cb),
		},
		// Knight
		{
			from: 1, to: 18,
			cb:       cb,
			expected: true,
			actual:   isValidMove(1, 18, "n", cb),
		},
		{
			from: 6, to: 22,
			cb:       cb,
			expected: false,
			actual:   isValidMove(6, 22, "n", cb),
		},
		// Bishop
		{
			from: 2, to: 47,
			cb:       cb,
			expected: false,
			actual:   isValidMove(2, 47, "b", cb),
		},
		{
			from: 2, to: 11,
			cb:       cb,
			expected: false,
			actual:   isValidMove(2, 11, "b", cb),
		},
		// Bishop passing edge
		{
			from: 5, to: 32,
			cb:       cb,
			expected: false,
			actual:   isValidMove(5, 32, "b", cb),
		},
		// Rook
		{
			from: 0, to: 7,
			cb:       cb,
			expected: false,
			actual:   isValidMove(0, 7, "r", cb),
		},
		{
			from: 7, to: 15,
			cb:       cb,
			expected: false,
			actual:   isValidMove(7, 15, "r", cb),
		},
		// Rook/queen cannot wrap around the board edge
		{
			from: 7, to: 8,
			cb:       cb,
			expected: false,
			actual:   isValidMove(7, 8, "r", cb),
		},
		// King
		{
			from: 40, to: 39,
			cb:       cb,
			expected: false,
			actual:   isValidMove(40, 39, "", cb),
		},
		{
			from: 4, to: 13,
			cb:       cb,
			expected: false,
			actual:   isValidMove(4, 13, "k", cb),
		},
	}

	runValidMoveTests(t, tests)
}

func TestMovePiece(t *testing.T) {
	cb := board.New()
	movePiece(move{8, 16, "p", ""}, cb)

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
	movePiece(move{57, 42, "n", ""}, cb)
	movePiece(move{16, 24, "p", ""}, cb)
	// Waiting move by ng8
	movePiece(move{62, 45, "n", ""}, cb)
	movePiece(move{24, 32, "p", ""}, cb)

	movePiece(move{42, 32, "n", ""}, cb)

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
	promotePawn(uint64(1<<63), cb, "q")

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
	if getAttackedSquares(cb) != expected {
		t.Errorf("attacked/defender squares: want=%v, got=%v",
			read1Bits(expected), read1Bits(getAttackedSquares(cb)))
	}
}

func TestCastling(t *testing.T) {
	// Castling moves king and rook, and removes remaining castling rights.
	cb, err := board.FromFen("r3k2r/8/8/8/8/8/8/R3K2R w KQkq - 0 1")
	if err != nil {
		t.Error(err)
	}

	movePiece(move{4, 2, "k", ""}, cb)
	if cb.Kings[1] != uint64(1<<2) {
		t.Errorf("w king did not castle queenside. want=2, got=%v", read1Bits(cb.Kings[1]))
	}
	if cb.Rooks[1] != uint64(1<<3+1<<7) {
		t.Errorf("rook did not move for castling. want=[3 7], got=%v", read1Bits(cb.Rooks[1]))
	}
	if cb.CastleRights[1] != [2]bool{false, false} {
		t.Errorf("w king castle rights: want=[false false], got=%v", cb.CastleRights[1])
	}

	movePiece(move{60, 62, "k", ""}, cb)
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

	movePiece(move{0, 56, "r", ""}, cb)
	if cb.Rooks[1] != uint64(1<<7+1<<56) {
		t.Errorf("wrong rook squares: want=[7, 56], got=%v", read1Bits(cb.Rooks[1]))
	}
	if cb.CastleRights[1] != [2]bool{false, true} {
		t.Errorf("w king castle rights: want=[false true], got=%v", cb.CastleRights[1])
	}
	if cb.CastleRights[0] != [2]bool{false, true} {
		t.Errorf("b king castle rights: want=[false true], got=%v", cb.CastleRights[0])
	}

	movePiece(move{63, 7, "r", ""}, cb)
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
	attkSquares := getAttackedSquares(cb1)
	cb1.WToMove ^= 1
	cb1.Pieces[cb1.WToMove] ^= uint64(1 << cb1.KingSqs[cb1.WToMove])

	tests := []moveTestCase{
		{
			// Cannot move into check, castle through check, or move into friendly piece.
			expected: uint64(1<<12 + 1<<13),
			actual:   getKingMoves(cb1.KingSqs[1], attkSquares, cb1),
		},
	}

	cb2, err := board.FromFen("r3k2r/4R3/3P4/8/8/8/8/8 b Q - 0 1")
	if err != nil {
		t.Error(err)
	}
	cb2.Pieces[cb2.WToMove] ^= uint64(1 << cb2.KingSqs[cb2.WToMove])
	cb2.WToMove ^= 1
	attkSquares = getAttackedSquares(cb2)
	cb2.WToMove ^= 1
	cb2.Pieces[cb2.WToMove] ^= uint64(1 << cb2.KingSqs[cb2.WToMove])

	tests = append(tests, moveTestCase{
		// Cannot castle out of check or capture protected piece.
		expected: uint64(1<<61 + 1<<59),
		actual:   getKingMoves(cb2.KingSqs[0], attkSquares, cb2),
	})

	runMoveGenTests(t, tests)
}

type allMovesTestCase struct {
	expected, actual []move
}

func runGetAllMovesTests(t *testing.T, tests []allMovesTestCase) {
	for _, tt := range tests {
		//fmt.Printf("expected: %v\n", tt.expected)
		//fmt.Printf("actual: %v\n", tt.actual)
		if len(tt.expected) != len(tt.actual) {
			t.Errorf("wrong length: want=%d, got=%d",
				len(tt.expected), len(tt.actual))
		}
		for i, move := range tt.expected {
			if move != tt.actual[i] {
				t.Errorf("allMoves: want=%v, got=%v", tt.expected[i], tt.actual[i])
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
	expected, actual int
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
	capturesBlockers, attackerCount := getCheckingSquares(cb)
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
	capturesBlockers, attackerCount = getCheckingSquares(cb)
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
	capturesBlockers, attackerCount = getCheckingSquares(cb)
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
			expected: []move{{6, 5, "k", ""}, {6, 7, "k", ""}, {6, 13, "k", ""},
				{6, 15, "k", ""}, {2, 38, "b", ""}, {56, 62, "r", ""},
				{63, 62, "r", ""}, {3, 30, "q", ""}},
			actual: getAllMoves(cb),
		},
	}

	cb1, err := board.FromFen("R5rR/8/8/8/8/8/5b2/RNBQ2K1 w - - 0 1")
	if err != nil {
		t.Error(err)
	}
	tests = append(tests, allMovesTestCase{
		// Two checking pieces, so only the king can move.
		expected: []move{{6, 5, "k", ""}, {6, 7, "k", ""}, {6, 13, "k", ""},
			{6, 15, "k", ""}},
		actual: getAllMoves(cb1),
	})

	cb2, err := board.FromFen("rnbqkbnr/ppppp1pp/5p2/7Q/8/4P3/PPPP1PPP/RNB1KBNR b KQkq - 0 1")
	if err != nil {
		t.Error(err)
	}
	tests = append(tests, allMovesTestCase{
		// Only one move is possible: pawn blocks check.
		expected: []move{{54, 46, "p", ""}},
		actual:   getAllMoves(cb2),
	})

	cb3, err := board.FromFen("4k3/pppppppp/8/8/6n1/5P2/PPPPPKPP/5BNR w - - 0 1")
	if err != nil {
		t.Error(err)
	}
	tests = append(tests, allMovesTestCase{
		expected: []move{{13, 4, "k", ""}, {13, 22, "k", ""}, {21, 30, "p", ""}},
		actual:   getAllMoves(cb3),
	})

	cb4, err := board.FromFen("n1n5/PPPk4/8/8/8/4K2q/5p1p/7N w - - 2 3")
	if err != nil {
		t.Error(err)
	}
	tests = append(tests, allMovesTestCase{
		expected: []move{{20, 11, "k", ""}, {20, 12, "k", ""},
			{20, 13, "k", ""}, {20, 27, "k", ""}, {20, 28, "k", ""},
			{20, 29, "k", ""}, {7, 22, "n", ""}},
		actual: getAllMoves(cb4),
	})

	runGetAllMovesTests(t, tests)
}

func TestRead1Bits(t *testing.T) {
	nums := read1Bits(uint64(0b11001))
	expected := []int{0, 3, 4}

	for i, num := range nums {
		if num != expected[i] {
			t.Errorf("failed to read 1 bits: want=%d, got=%d", expected[i], nums[i])
		}
	}
}

func TestBinSearch(t *testing.T) {
	nums := [8]int{0, 0, 0, 1, 2, 3, 4, 5}
	if !binSearch(1, nums) {
		t.Error("binSearch should find 1 in nums")
	}
	if !binSearch(3, nums) {
		t.Error("binSearch should find 3 in nums")
	}
	if binSearch(6, nums) {
		t.Error("binSearch should not find 6 in nums")
	}
}

func perft(depth int, cb *board.Board) int {
	if depth == 0 {
		return 1
	}
	nodes := 0
	moves := getAllMoves(cb)
	pos := board.StorePosition(cb)
	var attackedSquares uint64

	for _, toFrom := range moves {
		movePiece(toFrom, cb)
		if toFrom.piece == "k" {
			attackedSquares = uint64(0)
		} else {
			attackedSquares = getAttackedSquares(cb)
		}
		if cb.Kings[1^cb.WToMove]&attackedSquares == 0 {
			nodes += perft(depth-1, cb)
		}
		board.RestorePosition(pos, cb)
	}

	return nodes
}

func divide(depth int, cb *board.Board, ranksFiles ...[]string) {
	ranks := []string{"1", "2", "3", "4", "5", "6", "7", "8"}
	files := []string{"a", "b", "c", "d", "e", "f", "g", "h"}

	totalNodes := 0
	moves := getAllMoves(cb)
	var pos *board.Position

	pos = board.StorePosition(cb)
	for _, fromTo := range moves {
		nodes := 0
		movePiece(fromTo, cb)
		attackedSquares := getAttackedSquares(cb)
		if cb.Kings[1^cb.WToMove]&attackedSquares == 0 {
			nodes += perft(depth-1, cb)
		}
		board.RestorePosition(pos, cb)

		fromAlgNotation := strings.Join([]string{files[fromTo.from%8], ranks[fromTo.from/8]}, "")
		toAlgNotation := strings.Join([]string{files[fromTo.to%8], ranks[fromTo.to/8]}, "")

		fmt.Printf("%s%s %s: %d\n",
			fromAlgNotation, toAlgNotation, fromTo.promoteTo, nodes)
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
			actual:   perft(5, cb),
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
