package board

type Board struct {
	// TODO: Avoid branching by using knights[wToMove][square]
	WToMove int // 1 or 0, true or false

	WPieces  uint64
	WPawns   uint64
	WKnights uint64
	WBishops uint64
	WRooks   uint64
	WQueens  uint64
	WKing    uint64

	BPieces  uint64
	BPawns   uint64
	BKnights uint64
	BBishops uint64
	BRooks   uint64
	BQueens  uint64
	BKing    uint64

	NAttacks       [64]uint64
	KAttacks       [64]uint64
	SlidingAttacks [8][64]uint64
}

func New() *Board {
	return &Board{
		WToMove: 1,

		WPieces:  uint64(1)<<16 - 1,
		WPawns:   uint64(1)<<16 - 1 - (uint64(1)<<8 - 1),
		WKnights: uint64(1<<1 + 1<<6),
		WBishops: uint64(1<<2 + 1<<5),
		WRooks:   uint64(1<<0 + 1<<7),
		WQueens:  uint64(1 << 3),
		WKing:    uint64(1 << 4),

		BPieces:  uint64(1)<<63 - 1 + uint64(1)<<63 - (uint64(1)<<48 - 1),
		BPawns:   uint64(1)<<56 - 1 - (uint64(1)<<48 - 1),
		BKnights: uint64(1<<57 + 1<<62),
		BBishops: uint64(1<<58 + 1<<61),
		BRooks:   uint64(1<<56 + 1<<63),
		BQueens:  uint64(1 << 59),
		BKing:    uint64(1 << 60),

		NAttacks:       MakeKnightBBs(),
		KAttacks:       makeKingBBs(),
		SlidingAttacks: makeSlidingAttackBBs(),
	}
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

func containsN(n int, nums [8]int) bool {
	for _, num := range nums {
		if n == num {
			return true
		}
	}
	return false
}

func MakeKnightBBs() [64]uint64 {
	bbs := [64]uint64{}
	directions := []int{}
	files := getFiles()

	for sq := 0; sq < 64; sq++ {
		switch {
		case containsN(sq, files[0]):
			directions = []int{17, 10, -6, -15}
		case containsN(sq, files[1]):
			directions = []int{17, 15, 10, -6, -17, -15}
		case containsN(sq, files[2]):
			directions = []int{17, 15, -17, -15, 6, -10}
		case containsN(sq, files[3]):
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
	files := getFiles()

	for sq := 0; sq < 64; sq++ {
		switch {
		// file A
		case containsN(sq, files[0]):
			directions = []int{8, 9, 1, -7, -8}
		// file H
		case containsN(sq, files[3]):
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
	files := getFiles()
	// TODO: make containsN() generic to remove wasted zeroes.
	// Or use slices instead of arrays.
	fileAForbidden := [8]int{-9, -1, 7, 0, 0, 0, 0, 0}
	fileHForbidden := [8]int{9, 1, -7, 0, 0, 0, 0, 0}

	// Movement directions are ordered clockwise.
	for i, dir := range [8]int{8, 9, 1, -7, -8, -9, -1, 7} {
		for sq := 0; sq < 64; sq++ {
			if containsN(sq, files[0]) && containsN(dir, fileAForbidden) {
				continue
			} else if containsN(sq, files[3]) && containsN(dir, fileHForbidden) {
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
					(containsN(newSq, files[0]) || containsN(newSq, files[3])) {
					break
				}
			}
		}
	}

	return bbs
}