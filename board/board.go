package board

import (
	"fmt"
	"math/bits"
	"math/rand"
	"strings"
)

type Board struct {
	// TODO: move occupancies into one array? Possible memory speed boost
	WToMove int // 1 or 0, true or false

	Pieces  [2]uint64
	Pawns   [2]uint64
	Knights [2]uint64
	Bishops [2]uint64
	Rooks   [2]uint64
	Queens  [2]uint64
	Kings   [2]uint64

	PAttacks       [2][64]uint64
	NAttacks       [64]uint64
	KAttacks       [64]uint64
	SlidingAttacks [8][64]uint64

	KingSqs      [2]int
	CastleRights [2][2]bool // [b, w][queenside, kingside]

	EpSquare int
}

func New() *Board {
	return &Board{
		WToMove: 1,

		Pieces:  [2]uint64{0xFFFF000000000000, 0xFFFF},
		Pawns:   [2]uint64{0xFF000000000000, 0xFF00},
		Knights: [2]uint64{1<<57 + 1<<62, 1<<1 + 1<<6},
		Bishops: [2]uint64{1<<58 + 1<<61, 1<<2 + 1<<5},
		Rooks:   [2]uint64{1<<56 + 1<<63, 1<<0 + 1<<7},
		Queens:  [2]uint64{1 << 59, 1 << 3},
		Kings:   [2]uint64{1 << 60, 1 << 4},

		PAttacks:       makePawnBBs(),
		NAttacks:       makeKnightBBs(),
		KAttacks:       makeKingBBs(),
		SlidingAttacks: makeSlidingAttackBBs(),

		KingSqs:      [2]int{60, 4},
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
			cb.Pawns[color] += 1 << square
		case char == 'n' || char == 'N':
			cb.Knights[color] += 1 << square
		case char == 'b' || char == 'B':
			cb.Bishops[color] += 1 << square
		case char == 'r' || char == 'R':
			cb.Rooks[color] += 1 << square
		case char == 'q' || char == 'Q':
			cb.Queens[color] += 1 << square
		case char == 'k' || char == 'K':
			cb.Kings[color] += 1 << square
			cb.KingSqs[color] = square
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

	cb.Pieces[0] = cb.Pawns[0] | cb.Knights[0] | cb.Bishops[0] |
		cb.Rooks[0] | cb.Queens[0] | cb.Kings[0]
	cb.Pieces[1] = cb.Pawns[1] | cb.Knights[1] | cb.Bishops[1] |
		cb.Rooks[1] | cb.Queens[1] | cb.Kings[1]

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

	for sq := 0; sq < 64; sq++ {
		switch {
		// For determining pawn checks on black king on the eighth rank.
		case sq > 56:
			if sq%8 == 0 {
				bbs[0][sq] += 1 << (sq - 7)
			} else if sq%8 == 7 {
				bbs[0][sq] += 1 << (sq - 9)
			} else {
				bbs[0][sq] += 1<<(sq-7) + 1<<(sq-9)
			}
		// For determining pawn checks on white king on the first rank.
		case sq < 8:
			if sq%8 == 0 {
				bbs[1][sq] += 1 << (sq + 9)
			} else if sq%8 == 7 {
				bbs[1][sq] += 1 << (sq + 7)
			} else {
				bbs[1][sq] += 1<<(sq+7) + 1<<(sq+9)
			}
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

	Pieces  [2]uint64
	Pawns   [2]uint64
	Knights [2]uint64
	Bishops [2]uint64
	Rooks   [2]uint64
	Queens  [2]uint64
	Kings   [2]uint64

	KingSqs      [2]int
	CastleRights [2][2]bool

	EpSquare int
}

func StorePosition(cb *Board) *Position {
	return &Position{
		WToMove: cb.WToMove,
		Pieces:  cb.Pieces,
		Pawns:   cb.Pawns,
		Knights: cb.Knights,
		Bishops: cb.Bishops,
		Rooks:   cb.Rooks,
		Queens:  cb.Queens,
		Kings:   cb.Kings,

		KingSqs:      cb.KingSqs,
		CastleRights: cb.CastleRights,

		EpSquare: cb.EpSquare,
	}
}

func RestorePosition(pos *Position, cb *Board) {
	cb.WToMove = pos.WToMove
	cb.Pieces = pos.Pieces
	cb.Pawns = pos.Pawns
	cb.Knights = pos.Knights
	cb.Bishops = pos.Bishops
	cb.Rooks = pos.Rooks
	cb.Queens = pos.Queens
	cb.Kings = pos.Kings

	cb.KingSqs = pos.KingSqs
	cb.CastleRights = pos.CastleRights

	cb.EpSquare = pos.EpSquare
}

func (cb *Board) Print() {
	// Possibly destructive to original cb.
	squares := [64]string{}
	copied := StorePosition(cb)

	pieces := [6]uint64{
		copied.Pawns[0] + copied.Pawns[1],
		copied.Knights[0] + copied.Knights[1],
		copied.Bishops[0] + copied.Bishops[1],
		copied.Rooks[0] + copied.Rooks[1],
		copied.Queens[0] + copied.Queens[1],
		copied.Kings[0] + copied.Kings[1],
	}
	symbols := [6]string{"p", "n", "b", "r", "q", "k"}

	for i, piece := range pieces {
		for piece != 0 {
			squares[bits.TrailingZeros64(piece)] = symbols[i]
			piece &= piece - 1
		}
	}

	for i, symbol := range squares {
		if copied.Pieces[1]&uint64(1<<i) != 0 {
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
	fmt.Println(squares[7])
}

func Print1Bits(bb uint64) {
	squares := [64]int{}
	for bb > 0 {
		squares[bits.TrailingZeros64(bb)] = 1
		bb &= bb - 1
	}
	for i := 56; i != 7; i++ {
		if squares[i] == 0 {
			fmt.Print("- ")
		} else {
			fmt.Print("1 ")
		}
		if i%8 == 7 {
			i -= 16
			fmt.Println()
		}
	}
	if squares[7] == 1 {
		fmt.Println("1")
	} else {
		fmt.Println("-")
	}
}

func randFewBits() uint64 {
	return rand.Uint64() & rand.Uint64() & rand.Uint64()
}

func count1Bits(bb uint64) int {
	count := 0
	for bb > 0 {
		count += 1
		bb &= bb - 1
	}
	return count
}

func rookMask(square int) uint64 {
	var mask uint64
	origFile := square % 8
	origRank := square / 8

	for r := origRank + 1; r <= 6; r++ {
		mask |= uint64(1 << (origFile + r*8))
	}
	for r := origRank - 1; r >= 1; r-- {
		mask |= uint64(1 << (origFile + r*8))
	}
	for f := origFile + 1; f <= 6; f++ {
		mask |= uint64(1 << (f + origRank*8))
	}
	for f := origFile - 1; f >= 1; f-- {
		mask |= uint64(1 << (f + origRank*8))
	}

	return mask
}

func bishopMask(square int) uint64 {
	var mask uint64
	origFile := square % 8
	origRank := square / 8

	// Northeast
	f := origFile + 1
	for r := origRank + 1; r <= 6 && f <= 6; r++ {
		mask |= uint64(1 << (f + r*8))
		f += 1
	}
	// Northwest
	f = origFile - 1
	for r := origRank + 1; r <= 6 && f >= 1; r++ {
		mask |= uint64(1 << (f + r*8))
		f -= 1
	}
	// Southeast
	f = origFile + 1
	for r := origRank - 1; r >= 1 && f <= 6; r-- {
		mask |= uint64(1 << (f + r*8))
		f += 1
	}
	// Southwest
	f = origFile - 1
	for r := origRank - 1; r >= 1 && f >= 1; r-- {
		mask |= uint64(1 << (f + r*8))
		f -= 1
	}

	return mask
}

func rookAttacks(square int, blockers uint64) uint64 {
	var attacks uint64
	origRank := square / 8
	origFile := square % 8

	// North
	for r := origRank + 1; r <= 7; r++ {
		attacks |= uint64(1 << (origFile + r*8))
		if blockers&uint64(1<<(origFile+r*8)) != 0 {
			break
		}
	}
	// South
	for r := origRank - 1; r >= 0; r-- {
		attacks |= uint64(1 << (origFile + r*8))
		if blockers&uint64(1<<(origFile+r*8)) != 0 {
			break
		}
	}
	// East
	for f := origFile + 1; f <= 7; f++ {
		attacks |= uint64(1 << (f + origRank*8))
		if blockers&uint64(1<<(f+origRank*8)) != 0 {
			break
		}
	}
	// West
	for f := origFile - 1; f >= 0; f-- {
		attacks |= uint64(1 << (f + origRank*8))
		if blockers&uint64(1<<(f+origRank*8)) != 0 {
			break
		}
	}
	return attacks
}

func bishopAttacks(square int, blockers uint64) uint64 {
	var attacks uint64
	origFile := square % 8
	origRank := square / 8

	// Northeast
	f := origFile + 1
	for r := origRank + 1; r <= 7 && f <= 7; r++ {
		attacks |= uint64(1 << (f + r*8))
		if blockers&uint64(1<<(f+r*8)) != 0 {
			break
		}
		f += 1
	}
	// Northwest
	f = origFile - 1
	for r := origRank + 1; r <= 7 && f >= 0; r++ {
		attacks |= uint64(1 << (f + r*8))
		if blockers&uint64(1<<(f+r*8)) != 0 {
			break
		}
		f -= 1
	}
	// Southeast
	f = origFile + 1
	for r := origRank - 1; r >= 0 && f <= 7; r-- {
		attacks |= uint64(1 << (f + r*8))
		if blockers&uint64(1<<(f+r*8)) != 0 {
			break
		}
		f += 1
	}
	// Southwest
	f = origFile - 1
	for r := origRank - 1; r >= 0 && f >= 0; r-- {
		attacks |= uint64(1 << (f + r*8))
		if blockers&uint64(1<<(f+r*8)) != 0 {
			break
		}
		f -= 1
	}

	return attacks
}
