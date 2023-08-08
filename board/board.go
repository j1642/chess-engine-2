package board

import (
	"fmt"
	"math/bits"
	"strings"
)

type Board struct {
	// TODO: move occupancies into one array? Possible memory speed boost
	WToMove int // 1 or 0, true or false

	BwPieces  [2]uint64
	BwPawns   [2]uint64
	BwKnights [2]uint64
	BwBishops [2]uint64
	BwRooks   [2]uint64
	BwQueens  [2]uint64
	BwKing    [2]uint64

	PAttacks       [2][64]uint64
	NAttacks       [64]uint64
	KAttacks       [64]uint64
	SlidingAttacks [8][64]uint64

	KingSquare   [2]int
	CastleRights [2][2]bool // [b, w][queenside, kingside]

	EpSquare int
}

func New() *Board {
	return &Board{
		WToMove: 1,

		BwPieces:  [2]uint64{0xFFFF000000000000, 0xFFFF},
		BwPawns:   [2]uint64{0xFF000000000000, 0xFF00},
		BwKnights: [2]uint64{1<<57 + 1<<62, 1<<1 + 1<<6},
		BwBishops: [2]uint64{1<<58 + 1<<61, 1<<2 + 1<<5},
		BwRooks:   [2]uint64{1<<56 + 1<<63, 1<<0 + 1<<7},
		BwQueens:  [2]uint64{1 << 59, 1 << 3},
		BwKing:    [2]uint64{1 << 60, 1 << 4},

		PAttacks:       makePawnBBs(),
		NAttacks:       makeKnightBBs(),
		KAttacks:       makeKingBBs(),
		SlidingAttacks: makeSlidingAttackBBs(),

		KingSquare:   [2]int{60, 4},
		CastleRights: [2][2]bool{{true, true}, {true, true}},

		EpSquare: 100,
	}
}

func FromFen(fen string) (*Board, error) {
	// Build a Board from Forsyth-Edwards notation (FEN).
	/*
	   TODO: What is a good design for changing board.Board fields based on piece type?
	   Separate functions seem cluttered,
	   a map[rune]uint64 is nice for making uint64 but not for changing Board fields,
	   and switch statements seem too stuck in the details.
	*/
	var color int
	cb := &Board{}
	square := 56
	firstSpace := strings.IndexByte(fen, ' ')
	secondSpace := strings.IndexByte(fen[firstSpace+1:], ' ')

	if firstSpace == -1 || secondSpace != 1 {
		return cb, fmt.Errorf("invalid FEN string")
	}
	slashCount := strings.Count(fen, "/")
	if slashCount != 7 {
		return cb, fmt.Errorf("invalid FEN slash count. want=7, got=%d", slashCount)
	}

	for _, char := range fen[:firstSpace] {
		if 'A' <= char && char <= 'Z' {
			color = 1
		} else if 'a' <= char && char <= 'z' {
			color = 0
		} else {
			color = 100 // placeholder value
		}

		switch {
		case '1' <= char && char <= '8':
			// Negate the "square += 1" at the end of the loop
			square += int(char-'0') - 1
		case char == '/':
			// Negate the "square += 1" at the end of the loop
			square -= 17
		case char == 'p' || char == 'P':
			cb.BwPawns[color] += 1 << square
		case char == 'n' || char == 'N':
			cb.BwKnights[color] += 1 << square
		case char == 'b' || char == 'B':
			cb.BwBishops[color] += 1 << square
		case char == 'r' || char == 'R':
			cb.BwRooks[color] += 1 << square
		case char == 'q' || char == 'Q':
			cb.BwQueens[color] += 1 << square
		case char == 'k' || char == 'K':
			cb.BwKing[color] += 1 << square
			cb.KingSquare[color] = square
		}

		square += 1
	}

	// TODO: Include move count?
	for i, char := range fen[firstSpace:] {
		switch {
		case char == 'b':
			cb.WToMove = 0
		case char == 'w':
			cb.WToMove = 1
		case char == 'K':
			cb.CastleRights[1][1] = true
		case char == 'k':
			cb.CastleRights[0][1] = true
		case char == 'Q':
			cb.CastleRights[1][0] = true
		case char == 'q':
			cb.CastleRights[0][0] = true
			// An empty castling field is also represented by '-'. Any EP will later
			// overwrite this.
		case char == '-':
			cb.EpSquare = 100
		case 'a' <= char && char <= 'h':
			// rank 1: square=0+column, rank 2: square=8+column, ...
			rank := 8 * (int(fen[i+firstSpace+1]-'0') - 1)
			cb.EpSquare = (int(char - 'a')) + rank
		}
	}

	cb.BwPieces[0] = cb.BwPawns[0] | cb.BwKnights[0] | cb.BwBishops[0] |
		cb.BwRooks[0] | cb.BwQueens[0] | cb.BwKing[0]
	cb.BwPieces[1] = cb.BwPawns[1] | cb.BwKnights[1] | cb.BwBishops[1] |
		cb.BwRooks[1] | cb.BwQueens[1] | cb.BwKing[1]

	cb.PAttacks = makePawnBBs()
	cb.NAttacks = makeKnightBBs()
	cb.KAttacks = makeKingBBs()
	cb.SlidingAttacks = makeSlidingAttackBBs()

	return cb, nil
}

func GetFiles() [4][8]int {
	fileA, fileB, fileG, fileH := [8]int{}, [8]int{}, [8]int{}, [8]int{}

	for i := 0; i < 8; i++ {
		fileA[i] = i * 8
		fileB[i] = i*8 + 1
		fileG[i] = i*8 + 6
		fileH[i] = i*8 + 7
	}

	return [4][8]int{fileA, fileB, fileG, fileH}
}

/*
func getFiles() [4]map[int]bool {
    When only making knight BBs, maps are slower. Keep in case maps become
    faster when used for more pieces.

    fileA, fileB, fileG, fileH := make(map[int]bool, 8), make(map[int]bool, 8),
        make(map[int]bool, 8), make(map[int]bool, 8)

    for i := 0; i < 8; i++ {
        fileA[i*8] = true
        fileB[i*8+1] = true
        fileG[i*8+6] = true
        fileH[i*8+7] = true
    }

    return [4]map[int]bool{fileA, fileB, fileG, fileH}
}
*/

type IntArray interface {
	[8]int | [4]int | [3]int
}

func ContainsN[T IntArray](n int, nums T) bool {
	for i := 0; i < len(nums); i++ {
		if n == nums[i] {
			return true
		}
	}
	return false
}

func makePawnBBs() [2][64]uint64 {
	// First index is isWhite: 1 for white pawns, 0 for black pawns.
	bbs := [2][64]uint64{}

	for sq := 8; sq < 56; sq++ {
		switch {
		case sq%8 == 0:
			bbs[1][sq] += 1 << (sq + 9)
			bbs[0][sq] += 1 << (sq - 7)
		case sq%8 == 7:
			bbs[1][sq] += 1 << (sq + 7)
			bbs[0][sq] += 1 << (sq - 9)
		default:
			bbs[1][sq] += 1<<(sq+7) + 1<<(sq+9)
			bbs[0][sq] += 1<<(sq-7) + 1<<(sq-9)
		}
	}

	return bbs
}

func makeKnightBBs() [64]uint64 {
	bbs := [64]uint64{}
	directions := []int{}
	files := GetFiles()

	for sq := 0; sq < 64; sq++ {
		switch {
		case ContainsN(sq, files[0]):
			directions = []int{17, 10, -6, -15}
		case ContainsN(sq, files[1]):
			directions = []int{17, 15, 10, -6, -17, -15}
		case ContainsN(sq, files[2]):
			directions = []int{17, 15, -17, -15, 6, -10}
		case ContainsN(sq, files[3]):
			directions = []int{15, -17, 6, -10}
		default:
			directions = []int{17, 15, 10, -6, -17, -15, 6, -10}
		}
		/*
		   190ms bin search vs 180ms linear search
		   case binSearch(sq, files[0]):
		       directions = []int{17, 10, -6, -15}
		   case binSearch(sq, files[1]):
		       directions = []int{17, 15, 10, -6, -17, -15}
		   case binSearch(sq, files[2]):
		       directions = []int{17, 15, -17, -15, 6, -10}
		   case binSearch(sq, files[3]):
		       directions = []int{15, -17, 6, -10}
		*/

		/*
		   When only making knight BBs, maps are slower. Keep to check if maps are
		   faster when used for more pieces.

		   if  _, ok := files[0][sq]; ok {
		       directions = []int{17, 10, -6, -15}
		   } else if _, ok := files[1][sq]; ok {
		       directions = []int{17, 15, 10, -6, -17, -15}
		   } else if _, ok := files[2][sq]; ok {
		       directions = []int{17, 15, -17, -15, 6, -10}
		   } else if _, ok := files[3][sq]; ok {
		       directions = []int{15, -17, 6, -10}
		   } else {
		       directions = []int{17, 15, 10, -6, -17, -15, 6, -10}
		   }
		*/

		for _, d := range directions {
			if sq+d < 0 || sq+d > 63 {
				continue
			}
			bbs[sq] += 1 << (sq + d)
		}
	}

	return bbs
}

func makeKingBBs() [64]uint64 {
	bbs := [64]uint64{}
	directions := []int{}
	files := GetFiles()

	for sq := 0; sq < 64; sq++ {
		switch {
		// file A
		case ContainsN(sq, files[0]):
			directions = []int{8, 9, 1, -7, -8}
		// file H
		case ContainsN(sq, files[3]):
			directions = []int{8, 7, -1, -9, -8}
		default:
			directions = []int{7, 8, 9, -1, 1, -9, -8, -7}
		}

		for _, d := range directions {
			if sq+d < 0 || sq+d > 63 {
				continue
			}
			bbs[sq] += 1 << (sq + d)
		}
	}

	return bbs
}

func makeSlidingAttackBBs() [8][64]uint64 {
	bbs := [8][64]uint64{}
	files := GetFiles()
	fileAForbidden := [3]int{-9, -1, 7}
	fileHForbidden := [3]int{9, 1, -7}

	// Movement directions are ordered clockwise.
	for i, dir := range [8]int{8, 9, 1, -7, -8, -9, -1, 7} {
		for sq := 0; sq < 64; sq++ {
			if ContainsN(sq, files[0]) && ContainsN(dir, fileAForbidden) {
				continue
			} else if ContainsN(sq, files[3]) && ContainsN(dir, fileHForbidden) {
				continue
			}

			for j := 1; j < 8; j++ {
				newSq := j*dir + sq
				if newSq < 0 || newSq > 63 {
					break
				}
				bbs[i][sq] += 1 << newSq
				// Found board edge
				if dir != 8 && dir != -8 &&
					(ContainsN(newSq, files[0]) || ContainsN(newSq, files[3])) {
					break
				}
			}
		}
	}

	return bbs
}

type Position struct {
	WToMove int

	BwPieces  [2]uint64
	BwPawns   [2]uint64
	BwKnights [2]uint64
	BwBishops [2]uint64
	BwRooks   [2]uint64
	BwQueens  [2]uint64
	BwKing    [2]uint64

	KingSquare   [2]int
	CastleRights [2][2]bool

	EpSquare int
}

func StorePosition(cb *Board) *Position {
	return &Position{
		WToMove:   cb.WToMove,
		BwPieces:  cb.BwPieces,
		BwPawns:   cb.BwPawns,
		BwKnights: cb.BwKnights,
		BwBishops: cb.BwBishops,
		BwRooks:   cb.BwRooks,
		BwQueens:  cb.BwQueens,
		BwKing:    cb.BwKing,

		KingSquare:   cb.KingSquare,
		CastleRights: cb.CastleRights,

		EpSquare: cb.EpSquare,
	}
}

func RestorePosition(pos *Position, cb *Board) {
	cb.WToMove = pos.WToMove
	cb.BwPieces = pos.BwPieces
	cb.BwPawns = pos.BwPawns
	cb.BwKnights = pos.BwKnights
	cb.BwBishops = pos.BwBishops
	cb.BwRooks = pos.BwRooks
	cb.BwQueens = pos.BwQueens
	cb.BwKing = pos.BwKing

	cb.KingSquare = pos.KingSquare
	cb.CastleRights = pos.CastleRights

	cb.EpSquare = pos.EpSquare
}

func (cb *Board) Print() {
	// Possibly destructive to original cb.
	squares := [64]string{}
	copied := StorePosition(cb)

	pieces := [6]uint64{
		copied.BwPawns[0] + copied.BwPawns[1],
		copied.BwKnights[0] + copied.BwKnights[1],
		copied.BwBishops[0] + copied.BwBishops[1],
		copied.BwRooks[0] + copied.BwRooks[1],
		copied.BwQueens[0] + copied.BwQueens[1],
		copied.BwKing[0] + copied.BwKing[1],
	}
	symbols := [6]string{"p", "n", "b", "r", "q", "k"}

	for i, piece := range pieces {
		for piece != 0 {
			squares[bits.TrailingZeros64(piece)] = symbols[i]
			piece &= piece - 1
		}
	}

	for i, symbol := range squares {
		if copied.BwPieces[1]&uint64(1<<i) != 0 {
			squares[i] = strings.ToUpper(symbol)
		}
	}

	for i := 56; i != 7; i++ {
		if squares[i] == "" {
			fmt.Print("- ")
		} else {
			fmt.Printf("%s ", squares[i])
		}
		if i%8 == 7 {
			i -= 16
			fmt.Println()
		}
	}
	fmt.Print(squares[7], "\n")
}
