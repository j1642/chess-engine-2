// Pre-calculated move bitboards, including "fancy magic" bitboards for sliding pieces
package moves

import (
	"github.com/j1642/chess-engine-2/board"
	"math/bits"
)

var Sliding = makeSlidingAttackBBs()
var Pawn = makePawnBBs()
var Knight = makeKnightBBs()
var King = makeKingBBs()

var RookRelevantOccs [64]uint64
var RookOneBitCounts [64]int

// TODO: only a few squares have 4096 possible blocker configurations.
// Try to use [64][]uint64 slices instead
var RookMagicAttacks [64][4096]uint64 = buildRookMagicBB()

var BishopRelevantOccs [64]uint64
var BishopOneBitCounts [64]int
var BishopMagicAttacks [64][512]uint64 = buildBishopMagicBB()

func makeSlidingAttackBBs() [8][64]uint64 {
	bbs := [8][64]uint64{}
	files := board.GetFiles()
	fileAForbidden := [3]int8{-9, -1, 7}
	fileHForbidden := [3]int8{9, 1, -7}

	// Movement directions are ordered clockwise, starting from north
	for i, dir := range [8]int8{8, 9, 1, -7, -8, -9, -1, 7} {
		for sq := int8(0); sq < 64; sq++ {
			if board.ContainsN(sq, files[0]) && board.ContainsN(dir, fileAForbidden) {
				continue
			} else if board.ContainsN(sq, files[3]) && board.ContainsN(dir, fileHForbidden) {
				continue
			}

			for j := int8(1); j < 8; j++ {
				newSq := j*dir + sq
				if newSq < 0 || newSq > 63 {
					break
				}
				bbs[i][sq] += 1 << newSq
				// Found board edge
				if dir != 8 && dir != -8 &&
					(board.ContainsN(newSq, files[0]) || board.ContainsN(newSq, files[3])) {
					break
				}
			}
		}
	}

	return bbs
}

// Return pawn attack bitboards so attacks aren't repeatedly calculated on the fly
func makePawnBBs() [2][64]uint64 {
	// First index is cb.WToMove: 1 for white pawns, 0 for black pawns.
	bbs := [2][64]uint64{}
	for sq := 0; sq < 64; sq++ {
		switch {
		// Used in pieces.getCheckingSquares() for pawn checks on the black king on the eighth rank
		// NOT FOR PAWN ATTACKS. Pawns on the first and eighth ranks checked in pieces.getPawnMoves()
		case sq > 56:
			if sq%8 == 0 {
				bbs[0][sq] += 1 << (sq - 7)
			} else if sq%8 == 7 {
				bbs[0][sq] += 1 << (sq - 9)
			} else {
				bbs[0][sq] += 1<<(sq-7) + 1<<(sq-9)
			}
		// Used in pieces.getCheckingSquares() for pawn checks on the black king on the eighth rank
		// NOT FOR PAWN ATTACKS. Pawns on the first and eighth ranks checked in pieces.getPawnMoves()
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
	var directions []int8
	files := board.GetFiles()

	for sq := int8(0); sq < 64; sq++ {
		switch {
		case board.ContainsN(sq, files[0]):
			directions = []int8{17, 10, -6, -15}
		case board.ContainsN(sq, files[1]):
			directions = []int8{17, 15, 10, -6, -17, -15}
		case board.ContainsN(sq, files[2]):
			directions = []int8{17, 15, -17, -15, 6, -10}
		case board.ContainsN(sq, files[3]):
			directions = []int8{15, -17, 6, -10}
		default:
			directions = []int8{17, 15, 10, -6, -17, -15, 6, -10}
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

func makeKingBBs() [64]uint64 {
	bbs := [64]uint64{}
	var directions []int8
	files := board.GetFiles()

	for sq := int8(0); sq < 64; sq++ {
		switch {
		// file A
		case board.ContainsN(sq, files[0]):
			directions = []int8{8, 9, 1, -7, -8}
		// file H
		case board.ContainsN(sq, files[3]):
			directions = []int8{8, 7, -1, -9, -8}
		default:
			directions = []int8{7, 8, 9, -1, 1, -9, -8, -7}
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

var RookMagics = [64]uint64{
	0xa8002c000108020, 0x6c00049b0002001, 0x100200010090040, 0x2480041000800801,
	0x280028004000800, 0x900410008040022, 0x280020001001080, 0x2880002041000080,
	0xa000800080400034, 0x4808020004000, 0x2290802004801000, 0x411000d00100020,
	0x402800800040080, 0xb000401004208, 0x2409000100040200, 0x1002100004082,
	0x22878001e24000, 0x1090810021004010, 0x801030040200012, 0x500808008001000,
	0xa08018014000880, 0x8000808004000200, 0x201008080010200, 0x801020000441091,
	0x800080204005, 0x1040200040100048, 0x120200402082, 0xd14880480100080,
	0x12040280080080, 0x100040080020080, 0x9020010080800200, 0x813241200148449,
	0x491604001800080, 0x100401000402001, 0x4820010021001040, 0x400402202000812,
	0x209009005000802, 0x810800601800400, 0x4301083214000150, 0x204026458e001401,
	0x40204000808000, 0x8001008040010020, 0x8410820820420010, 0x1003001000090020,
	0x804040008008080, 0x12000810020004, 0x1000100200040208, 0x430000a044020001,
	0x280009023410300, 0xe0100040002240, 0x200100401700, 0x2244100408008080,
	0x8000400801980, 0x2000810040200, 0x8010100228810400, 0x2000009044210200,
	0x4080008040102101, 0x40002080411d01, 0x2005524060000901, 0x502001008400422,
	0x489a000810200402, 0x1004400080a13, 0x4000011008020084, 0x26002114058042,
}

var BishopMagics = [64]uint64{
	0x89a1121896040240, 0x2004844802002010, 0x2068080051921000, 0x62880a0220200808,
	0x4042004000000, 0x100822020200011, 0xc00444222012000a, 0x28808801216001,
	0x400492088408100, 0x201c401040c0084, 0x840800910a0010, 0x82080240060,
	0x2000840504006000, 0x30010c4108405004, 0x1008005410080802, 0x8144042209100900,
	0x208081020014400, 0x4800201208ca00, 0xf18140408012008, 0x1004002802102001,
	0x841000820080811, 0x40200200a42008, 0x800054042000, 0x88010400410c9000,
	0x520040470104290, 0x1004040051500081, 0x2002081833080021, 0x400c00c010142,
	0x941408200c002000, 0x658810000806011, 0x188071040440a00, 0x4800404002011c00,
	0x104442040404200, 0x511080202091021, 0x4022401120400, 0x80c0040400080120,
	0x8040010040820802, 0x480810700020090, 0x102008e00040242, 0x809005202050100,
	0x8002024220104080, 0x431008804142000, 0x19001802081400, 0x200014208040080,
	0x3308082008200100, 0x41010500040c020, 0x4012020c04210308, 0x208220a202004080,
	0x111040120082000, 0x6803040141280a00, 0x2101004202410000, 0x8200000041108022,
	0x21082088000, 0x2410204010040, 0x40100400809000, 0x822088220820214,
	0x40808090012004, 0x910224040218c9, 0x402814422015008, 0x90014004842410,
	0x1000042304105, 0x10008830412a00, 0x2520081090008908, 0x40102000a0a60140,
}

func buildRookMagicBB() [64][4096]uint64 {
	var rookAttackBBs [64][4096]uint64
	cb, err := board.FromFen("8/8/8/8/8/8/8/8 w - 0 1")
	if err != nil {
		panic(err)
	}
	rank_1 := uint64(0xff)
	rank_8 := uint64(0xff << 56)
	file_a := uint64(0x101010101010101)
	file_h := uint64(0x8080808080808080)

	for square := 0; square < 64; square++ {
		square_bb := uint64(1 << square)
		cb.Pieces[0] = 0
		empty_board_attack_bb := CalculateRookMoves(square, cb)
		for _, line := range [4]uint64{rank_1, rank_8, file_a, file_h} {
			// if square not in the rank/file
			if square_bb|line != line {
				empty_board_attack_bb &= ^line
			}
		}
		count_1_bits := bits.OnesCount64(empty_board_attack_bb)

		RookRelevantOccs[square] = empty_board_attack_bb
		RookOneBitCounts[square] = count_1_bits

		possible_occupancies_count := 1 << bits.OnesCount64(empty_board_attack_bb) //(2**...)
		permutations := make([]uint64, possible_occupancies_count)
		blockers := uint64(0)
		perm_idx := 0

		for {
			permutations[perm_idx] = blockers
			perm_idx++
			blockers = (blockers - empty_board_attack_bb) & empty_board_attack_bb
			if blockers == 0 {
				break
			}
		}
		if perm_idx != possible_occupancies_count {
			panic("some occupancies were not calculated")
		}

		for _, occupancy := range permutations {
			cb.Pieces[0] = occupancy
			masked := occupancy & empty_board_attack_bb

			for _, line := range [4]uint64{rank_1, rank_8, file_a, file_h} {
				// if square not in the rank/file
				if square_bb|line != line {
					masked &= ^line
				}
			}
			idx := (masked * RookMagics[square]) >> (64 - count_1_bits)
			rookAttackBBs[square][idx] = CalculateRookMoves(square, cb)
		}
		cb.Pieces[0] = 0
	}

	return rookAttackBBs
}

// Captures and protection are included in move gen.
func CalculateRookMoves(square int, cb *board.Board) uint64 {
	occupied := cb.Pieces[0] | cb.Pieces[1]
	// North
	moves := Sliding[0][square]
	blockers := Sliding[0][square] & occupied
	blockerSq := bits.TrailingZeros64(blockers | uint64(1<<63))
	moves ^= Sliding[0][blockerSq]
	// East
	moves |= Sliding[2][square]
	blockers = Sliding[2][square] & occupied
	blockerSq = bits.TrailingZeros64(blockers | uint64(1<<63))
	moves ^= Sliding[2][blockerSq]
	// South
	moves |= Sliding[4][square]
	blockers = Sliding[4][square] & occupied
	blockerSq = 63 - bits.LeadingZeros64(blockers|uint64(1))
	moves ^= Sliding[4][blockerSq]
	// West
	moves |= Sliding[6][square]
	blockers = Sliding[6][square] & occupied
	blockerSq = 63 - bits.LeadingZeros64(blockers|uint64(1))
	moves ^= Sliding[6][blockerSq]

	return moves
}

func buildBishopMagicBB() [64][512]uint64 {
	var bishopAttackBBs [64][512]uint64
	cb, err := board.FromFen("8/8/8/8/8/8/8/8 w - 0 1")
	if err != nil {
		panic(err)
	}
	rank_1 := uint64(0xff)
	rank_8 := uint64(0xff << 56)
	file_a := uint64(0x101010101010101)
	file_h := uint64(0x8080808080808080)

	for square := 0; square < 64; square++ {
		square_bb := uint64(1 << square)
		cb.Pieces[0] = 0
		empty_board_attack_bb := CalculateBishopMoves(square, cb)
		for _, line := range [4]uint64{rank_1, rank_8, file_a, file_h} {
			// if square not in the rank/file
			if square_bb|line != line {
				empty_board_attack_bb &= ^line
			}
		}
		count_1_bits := bits.OnesCount64(empty_board_attack_bb)

		BishopRelevantOccs[square] = empty_board_attack_bb
		BishopOneBitCounts[square] = count_1_bits

		possible_occupancies_count := 1 << bits.OnesCount64(empty_board_attack_bb) //(2**...)
		permutations := make([]uint64, possible_occupancies_count)
		blockers := uint64(0)
		perm_idx := 0

		for {
			permutations[perm_idx] = blockers
			perm_idx++
			blockers = (blockers - empty_board_attack_bb) & empty_board_attack_bb
			if blockers == 0 {
				break
			}
		}
		if perm_idx != possible_occupancies_count {
			panic("some occupancies were not calculated")
		}

		for _, occupancy := range permutations {
			cb.Pieces[0] = occupancy
			masked := occupancy & empty_board_attack_bb

			for _, line := range [4]uint64{rank_1, rank_8, file_a, file_h} {
				// if square not in the rank/file
				if square_bb|line != line {
					masked &= ^line
				}
			}
			idx := (masked * BishopMagics[square]) >> (64 - count_1_bits)
			bishopAttackBBs[square][idx] = CalculateBishopMoves(square, cb)
		}
		cb.Pieces[0] = 0
	}

	return bishopAttackBBs
}

func CalculateBishopMoves(square int, cb *board.Board) uint64 {
	occupied := cb.Pieces[0] | cb.Pieces[1]
	// Northeast
	moves := Sliding[1][square]
	blockers := Sliding[1][square] & occupied
	blockerSq := bits.TrailingZeros64(blockers | uint64(1<<63))
	moves ^= Sliding[1][blockerSq]
	// Southeast
	moves |= Sliding[3][square]
	blockers = Sliding[3][square] & occupied
	blockerSq = 63 - bits.LeadingZeros64(blockers|uint64(1))
	moves ^= Sliding[3][blockerSq]
	// Southwest
	moves |= Sliding[5][square]
	blockers = Sliding[5][square] & occupied
	blockerSq = 63 - bits.LeadingZeros64(blockers|uint64(1))
	moves ^= Sliding[5][blockerSq]
	// Northwest
	moves |= Sliding[7][square]
	blockers = Sliding[7][square] & occupied
	blockerSq = bits.TrailingZeros64(blockers | uint64(1<<63))
	moves ^= Sliding[7][blockerSq]

	return moves
}
