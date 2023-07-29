package board

import (
	"testing"
)

// cb = chessboard, w = white, b = black

func TestNew(t *testing.T) {
	cb := New()

	if cb.wToMove != 1 {
		t.Errorf("initial wToMove: want=1, got=%d", cb.wToMove)
	}

	if cb.wPieces != uint64(1)<<16-1 {
		t.Errorf("initial wPieces: want=65535, got=%d", cb.wPieces)
	}
	if cb.wPawns != uint64(1)<<16-1-(uint64(1)<<8-1) {
		t.Errorf("initial wPawns: want=%b, got=%b", 65279, cb.wPawns)
	}

	wPiecesUnion := cb.wPawns | cb.wRooks | cb.wKnights | cb.wBishops |
		cb.wQueens | cb.wKing
	if cb.wPieces != wPiecesUnion {
		t.Errorf("wPieces != union of all white pieces. want=65535,\ngot=%b",
			wPiecesUnion)
	}

	bPieces := uint64(1)<<63 - 1 + uint64(1)<<63 - (uint64(1)<<48 - 1)
	if cb.bPieces != bPieces {
		t.Errorf("initial bPieces: want=%b\n, got=%b", bPieces, cb.bPieces)
	}
	bPawns := uint64(1)<<56 - 1 - (uint64(1)<<48 - 1)
	if cb.bPawns != bPawns {
		t.Errorf("initial bPawns: want=%b\n, got=%b", bPawns, cb.bPawns)
	}

	bPiecesUnion := cb.bPawns | cb.bRooks | cb.bKnights | cb.bBishops |
		cb.bQueens | cb.bKing
	if cb.bPieces != bPiecesUnion {
		t.Errorf("bPieces != union of all black pieces. want=%b,\ngot=%b",
			uint64(1)<<63-1+uint64(1)<<63-(uint64(1)<<48-1), bPiecesUnion)
	}
}
