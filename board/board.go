package board

import (
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

		PAttacks:       MakePawnBBs(),
		NAttacks:       MakeKnightBBs(),
		KAttacks:       MakeKingBBs(),
		SlidingAttacks: MakeSlidingAttackBBs(),

		KingSquare:   [2]int{60, 4},
		CastleRights: [2][2]bool{{true, true}, {true, true}},

		EpSquare: 100,
	}
}

func FromFen(fen string) *Board {
	/*
	   TODO: What is a good design for changing board.Board fields based on piece type?
	   Separate functions seem cluttered,
	   a map[rune]uint64 is nice for making uint64 but not for changing Board fields,
	   and switch statements seem too stuck in the details.
	*/
	/*
	   if !strings.Contains(in, " ") {
	       return cb, fmt.Errorf("invalid FEN string: does not contain spaces")
	   }*/
	square := 56
	cb := &Board{}
	var color int
	spaceIndex := strings.IndexAny(fen, " ")

	for _, char := range fen[:spaceIndex] {
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
	for i, char := range fen[spaceIndex:] {
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
		case char == '-':
			cb.EpSquare = 100
		case 'a' <= char && char <= 'h':
			factor := 8 * (int(fen[i+spaceIndex]-'a') - 1)
			cb.EpSquare = int(char-'a') * factor
		}
	}

	cb.BwPieces[0] = cb.BwPawns[0] | cb.BwKnights[0] | cb.BwBishops[0] |
		cb.BwRooks[0] | cb.BwQueens[0] | cb.BwKing[0]
	cb.BwPieces[1] = cb.BwPawns[1] | cb.BwKnights[1] | cb.BwBishops[1] |
		cb.BwRooks[1] | cb.BwQueens[1] | cb.BwKing[1]

	cb.PAttacks = MakePawnBBs()
	cb.NAttacks = MakeKnightBBs()
	cb.KAttacks = MakeKingBBs()
	cb.SlidingAttacks = MakeSlidingAttackBBs()

	return cb
}

func getFiles() [4][8]int {
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
fuenc getFiles() [4]map[int]bool {
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

func ContainsN(n int, nums [8]int) bool {
	for _, num := range nums {
		if n == num {
			return true
		}
	}
	return false
}

func MakePawnBBs() [2][64]uint64 {
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

func MakeKnightBBs() [64]uint64 {
	bbs := [64]uint64{}
	directions := []int{}
	files := getFiles()

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

func MakeKingBBs() [64]uint64 {
	bbs := [64]uint64{}
	directions := []int{}
	files := getFiles()

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

func MakeSlidingAttackBBs() [8][64]uint64 {
	bbs := [8][64]uint64{}
	files := getFiles()
	// TODO: make ContainsN() generic to remove wasted zeroes.
	// Or use slices instead of arrays.
	fileAForbidden := [8]int{-9, -1, 7, 0, 0, 0, 0, 0}
	fileHForbidden := [8]int{9, 1, -7, 0, 0, 0, 0, 0}

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
