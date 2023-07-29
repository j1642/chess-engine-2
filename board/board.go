package board

type Board struct {
	// TODO: Avoid branching by using knights[wToMove][square]
	wToMove int // 1 or 0, true or false

	wPieces  uint64
	wPawns   uint64
	wKnights uint64
	wBishops uint64
	wRooks   uint64
	wQueens  uint64
	wKing    uint64

	bPieces  uint64
	bPawns   uint64
	bKnights uint64
	bBishops uint64
	bRooks   uint64
	bQueens  uint64
	bKing    uint64
}

func New() *Board {
	return &Board{
		wToMove: 1,

		wPieces:  uint64(1)<<16 - 1,
		wPawns:   uint64(1)<<16 - 1 - (uint64(1)<<8 - 1),
		wKnights: uint64(1<<1 + 1<<6),
		wBishops: uint64(1<<2 + 1<<5),
		wRooks:   uint64(1<<0 + 1<<7),
		wQueens:  uint64(1 << 3),
		wKing:    uint64(1 << 4),

		bPieces:  uint64(1)<<63 - 1 + uint64(1)<<63 - (uint64(1)<<48 - 1),
		bPawns:   uint64(1)<<56 - 1 - (uint64(1)<<48 - 1),
		bKnights: uint64(1<<57 + 1<<62),
		bBishops: uint64(1<<58 + 1<<61),
		bRooks:   uint64(1<<56 + 1<<63),
		bQueens:  uint64(1 << 59),
		bKing:    uint64(1 << 60),
	}
}
