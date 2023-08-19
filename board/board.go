package board

import (
	"fmt"
	"math/bits"
	"math/rand"
	"strings"
)

var BishopMagics [64]map[int]uint64
var RookMagics [64]map[int]uint64

func init() {
	makeMagicAttacks()
}

func makeMagicAttacks() {
	var bBlockers, rBlockers []uint64
	var bMask, rMask uint64
	var key int

	for sq := 0; sq < 64; sq++ {
		BishopMagics[sq] = make(map[int]uint64, 1<<BishopBits[sq])
		RookMagics[sq] = make(map[int]uint64, 1<<RookBits[sq])

		rMask = RookMask(sq)
		rBlockers = occupancyCombos(read1BitsBoard(rMask))
		bMask = BishopMask(sq)
		bBlockers = occupancyCombos(read1BitsBoard(bMask))
		for _, blockers := range bBlockers {
			key = int((blockers * BMagics[sq]) >> (63 - BishopBits[sq]))
			BishopMagics[sq][key] = bishopAttacks(sq, blockers)
		}
		for _, blockers := range rBlockers {
			key = int((blockers * RMagics[sq]) >> (63 - RookBits[sq]))
			RookMagics[sq][key] = rookAttacks(sq, blockers)
		}
	}
}

func read1BitsBoard(bb uint64) []int {
	ones := make([]int, 0, 12)
	for bb > 0 {
		ones = append(ones, bits.LeadingZeros64(bb))
		bb &= bb - 1
	}
	return ones
}

type Board struct {
	// TODO: move occupancies into one array? Possible memory cache improvement
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

func getRandNum() uint64 {
	n := rand.Uint64() & 0xFFFF
	n |= (rand.Uint64() & 0xFFFF) << 16
	n |= (rand.Uint64() & 0xFFFF) << 32
	n |= (rand.Uint64() & 0xFFFF) << 48
	return n
}

func randFewBits() uint64 {
	return rand.Uint64() & rand.Uint64() & rand.Uint64()
	//return getRandNum() & getRandNum() & getRandNum()
}

func count1Bits(bb uint64) int {
	count := 0
	for bb > 0 {
		count += 1
		bb &= bb - 1
	}
	return count
}

func RookMask(square int) uint64 {
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

func BishopMask(square int) uint64 {
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

// Amount of possible attacks without blockers, excluding attacks to
// board edges
var RookBits = [64]int{
	12, 11, 11, 11, 11, 11, 11, 12,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	11, 10, 10, 10, 10, 10, 10, 11,
	12, 11, 11, 11, 11, 11, 11, 12,
}
var BishopBits = [64]int{
	6, 5, 5, 5, 5, 5, 5, 6,
	5, 5, 5, 5, 5, 5, 5, 5,
	5, 5, 7, 7, 7, 7, 5, 5,
	5, 5, 7, 9, 9, 7, 5, 5,
	5, 5, 7, 9, 9, 7, 5, 5,
	5, 5, 7, 7, 7, 7, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 5,
	6, 5, 5, 5, 5, 5, 5, 6,
}

func occupancyCombos(nums []int) []uint64 {
	// Return all possible occupancy combination bitboards for the given squares.
	combos := make([]uint64, 0, 1<<len(nums))
	result := make([]int, len(nums))

	// TODO: check memory use, probably not optimal
	// A closure seems like the easiest way to build 'combos'
	var combos2 func([]int, int, int, []int)
	combos2 = func(arr []int, l, start int, result []int) {
		if l == 0 {
			sum := uint64(0)
			for _, num := range result {
				// Zero and other corner squares shoulnd't need to be checked
				// Loops over zeroes when finding the sum. Not optimal
				if num != 0 {
					sum += 1 << num
				}
			}
			combos = append(combos, sum)
			return
		}
		for i := start; i <= len(arr)-l; i++ {
			result[len(result)-l] = arr[i]
			combos2(arr, l-1, i+1, result)
		}
	}

	combos = append(combos, 1)
	for k := 1; k < len(nums)+1; k++ {
		combos2(nums, k, 0, result)
	}

	return combos
}

func popLS1B(bb *uint64) int {
	ls1b := bits.TrailingZeros64(*bb)
	*bb &= *bb - 1
	return ls1b
}

func index_to_uint64(index, bits int, m *uint64) uint64 {
	var result uint64
	var j int

	for i := 0; i < bits; i++ {
		j = popLS1B(m)
		if index&1<<i != 0 {
			result |= 1 << j
		}
	}

	return result
}

func transform(b, magic uint64, bits int) int {
	return int((b * magic) >> (64 - bits))
}

func FindMagic(sq, m int, rb string) uint64 {
	var a, b, used [4096]uint64
	var mask, magic uint64

	if rb == "b" {
		mask = BishopMask(sq)
	} else {
		mask = RookMask(sq)
	}

	n := count1Bits(mask)

	var fail bool
	var i, j, k int
	for i = 0; i < (1 << n); i++ {
		b[i] = index_to_uint64(i, n, &mask)
		if rb == "b" {
			a[i] = bishopAttacks(sq, b[i])
		} else {
			a[i] = rookAttacks(sq, b[i])
		}
	}
	for k = 0; k < 100_000_000; k++ {
		magic = randFewBits()
		if count1Bits((mask*magic)&0xFF00000000000000) < 6 {
			continue
		}
		for i = 0; i < 4096; i++ {
			used[i] = 0
		}
		fail = false
		for i = 0; !fail && i < (1<<n); i++ {
			j = transform(b[i], magic, m)
			if used[j] == 0 {
				used[j] = a[i]
			} else if used[j] != a[i] {
				fail = true
			}
		}
		if !fail {
			return magic
		}
	}
	fmt.Println("failed")
	return 0
}

var RMagics = [64]uint64{
	0x2080020500400f0,
	0x28444000400010,
	0x20000a1004100014,
	0x20010c090202006,
	0x8408008200810004,
	0x1746000808002,
	0x2200098000808201,
	0x12c0002080200041,
	0x104000208e480804,
	0x8084014008281008,
	0x4200810910500410,
	0x100014481c20400c,
	0x4014a4040020808,
	0x401002001010a4,
	0x202000500010001,
	0x8112808005810081,
	0x40902108802020,
	0x42002101008101,
	0x459442200810c202,
	0x81001103309808,
	0x8110000080102,
	0x8812806008080404,
	0x104020000800101,
	0x40a1048000028201,
	0x4100ba0000004081,
	0x44803a4003400109,
	0xa010a00000030443,
	0x91021a000100409,
	0x4201e8040880a012,
	0x22a000440201802,
	0x30890a72000204,
	0x10411402a0c482,
	0x40004841102088,
	0x40230000100040,
	0x40100010000a0488,
	0x1410100200050844,
	0x100090808508411,
	0x1410040024001142,
	0x8840018001214002,
	0x410201000098001,
	0x8400802120088848,
	0x2060080000021004,
	0x82101002000d0022,
	0x1001101001008241,
	0x9040411808040102,
	0x600800480009042,
	0x1a020000040205,
	0x4200404040505199,
	0x2020081040080080,
	0x40a3002000544108,
	0x4501100800148402,
	0x81440280100224,
	0x88008000000804,
	0x8084060000002812,
	0x1840201000108312,
	0x5080202000000141,
	0x1042a180880281,
	0x900802900c01040,
	0x8205104104120,
	0x9004220000440a,
	0x8029510200708,
	0x8008440100404241,
	0x2420001111000bd,
	0x4000882304000041,
}

var BMagics = [64]uint64{
	0x100420000431024,
	0x280800101073404,
	0x42000a00840802,
	0xca800c0410c2,
	0x81004290941c20,
	0x400200450020250,
	0x444a019204022084,
	0x88610802202109a,
	0x11210a0800086008,
	0x400a08c08802801,
	0x1301a0500111c808,
	0x1280100480180404,
	0x720009020028445,
	0x91880a9000010a01,
	0x31200940150802b2,
	0x5119080c20000602,
	0x242400a002448023,
	0x4819006001200008,
	0x222c10400020090,
	0x302008420409004,
	0x504200070009045,
	0x210071240c02046,
	0x1182219000022611,
	0x400c50000005801,
	0x4004010000113100,
	0x2008121604819400,
	0xc4a4010000290101,
	0x404a000888004802,
	0x8820c004105010,
	0x28280100908300,
	0x4c013189c0320a80,
	0x42008080042080,
	0x90803000c080840,
	0x2180001028220,
	0x1084002a040036,
	0x212009200401,
	0x128110040c84a84,
	0x81488020022802,
	0x8c0014100181,
	0x2222013020082,
	0xa00100002382c03,
	0x1000280001005c02,
	0x84801010000114c,
	0x480410048000084,
	0x21204420080020a,
	0x2020010000424a10,
	0x240041021d500141,
	0x420844000280214,
	0x29084a280042108,
	0x84102a8080a20a49,
	0x104204908010212,
	0x40a20280081860c1,
	0x3044000200121004,
	0x1001008807081122,
	0x50066c000210811,
	0xe3001240f8a106,
	0x940c0204030020d4,
	0x619204000210826a,
	0x2010438002b00a2,
	0x884042004005802,
	0xa90240000006404,
	0x500d082244010008,
	0x28190d00040014e0,
	0x825201600c082444,
}
