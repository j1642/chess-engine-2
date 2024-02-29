package main

import (
	"engine2/board"
	"engine2/pieces"
	"fmt"
	"math/bits"
	"time"
)

var rookMagics = [64]uint64{
	0x2080020500400f0, 0x28444000400010, 0x20000a1004100014, 0x20010c090202006,
	0x8408008200810004, 0x1746000808002, 0x2200098000808201, 0x12c0002080200041,
	0x104000208e480804, 0x8084014008281008, 0x4200810910500410, 0x100014481c20400c,
	0x4014a4040020808, 0x401002001010a4, 0x202000500010001, 0x8112808005810081,
	0x40902108802020, 0x42002101008101, 0x459442200810c202, 0x81001103309808,
	0x8110000080102, 0x8812806008080404, 0x104020000800101, 0x40a1048000028201,
	0x4100ba0000004081, 0x44803a4003400109, 0xa010a00000030443, 0x91021a000100409,
	0x4201e8040880a012, 0x22a000440201802, 0x30890a72000204, 0x10411402a0c482,
	0x40004841102088, 0x40230000100040, 0x40100010000a0488, 0x1410100200050844,
	0x100090808508411, 0x1410040024001142, 0x8840018001214002, 0x410201000098001,
	0x8400802120088848, 0x2060080000021004, 0x82101002000d0022, 0x1001101001008241,
	0x9040411808040102, 0x600800480009042, 0x1a020000040205, 0x4200404040505199,
	0x2020081040080080, 0x40a3002000544108, 0x4501100800148402, 0x81440280100224,
	0x88008000000804, 0x8084060000002812, 0x1840201000108312, 0x5080202000000141,
	0x1042a180880281, 0x900802900c01040, 0x8205104104120, 0x9004220000440a,
	0x8029510200708, 0x8008440100404241, 0x2420001111000bd, 0x4000882304000041,
}

// TODO: make reference tables (aka magic bbs)
// 1 - attacked squares mask (excluding edges) & blocker pumutation = "masked blockers"
// 2 - "masked blockers" * magics[attacker square] = "index mapping"
// 3 - "index mapping" >> (64-n, n = bits in index mapping, bits.Len64()) = hash key for
//
//	that piece on that square with those blockers
//
// 4 - stored_attacks[hash key] = uint64(moves possible with those blockers)
// attack():
//
//	1 - blockers = occupied & sliding_piece_attacks[square]
//	2 - hash key = blockers magics[square] >> (64 - (WHAT IS THIS OBJ)[square])
//	3 - return bishop_attacks[square][hash key]

func main() {
	// implement magic bitboards
	//var rookBlockers [4096]uint64
	//var bishopBlockers [512]uint64
	cb, err := board.FromFen("8/8/8/8/8/8/8/8 w - 0 1")
	if err != nil {
		panic(err)
	}

	/*
	   rook_a7_bb := pieces.GetRookMoves(48, cb) - (1 << 0) - (1 << 55) - (1<<56)
	   fmt.Println("partial moves on a7:", pieces.Read1Bits(rook_a7_bb))
	   fmt.Println("bin len partial moves:", bits.OnesCount64(rook_a7_bb))
	   fmt.Println("all moves on a7:", pieces.Read1Bits(0x01FE010101010101))

	   rook_a7_bb_mult := rook_a7_bb * 0x48FFFE99FECFAA00
	   fmt.Println("mult, bin_len mult:   ", rook_a7_bb_mult, bits.OnesCount64(rook_a7_bb_mult))
	   hash_key := rook_a7_bb_mult >> (64 - bits.OnesCount64(rook_a7_bb))
	   fmt.Println("hash key:", hash_key)
	   // then, store rook_moves[square][hash] = getRookMoves(sq)

	   blockers := uint64(1<<55) + uint64(1<<56) + uint64(1<<48)
	   retreived := (blockers * 0x48FFFE99FECFAA00) >> (64 - bits.OnesCount64(rook_a7_bb))
	   fmt.Println("retreived hash:", retreived)
	*/

	//fmt.Println(buildRookMagicBB())
	rookBB := buildRookMagicBB()
	kiwipete, err := board.FromFen("r3k2r/p1ppqpb1/bn2pnp1/3PN3/1p2P3/2N2Q1p/PPPBBPPP/R3K2R w KQkq - 0 1")
	if err != nil {
		panic("")
	}

	var x uint64
	square := 20
	start := time.Now()
	for i := 0; i < 10_000; i++ {
		x = pieces.GetRookMoves(square, kiwipete)
	}
	elapsed := time.Since(start)
	fmt.Println("elapsed 1:", elapsed)
	fmt.Println("moves:", x)
	fmt.Println("moves:", pieces.Read1Bits(x))

	rank_1 := uint64(0xff)
	rank_8 := uint64(0xff << 56)
	file_a := uint64(0x101010101010101)
	file_h := uint64(0x8080808080808080)
	square_bb := uint64(1 << square)
	empty_board_attack_bb := pieces.GetRookMoves(square, cb)
	for _, line := range [4]uint64{rank_1, rank_8, file_a, file_h} {
		// if square not in the rank/file
		if square_bb|line != line {
			empty_board_attack_bb &= ^line
		}
	}
	count_1_bits := bits.OnesCount64(empty_board_attack_bb)
	if count_1_bits != 10 {
		panic(count_1_bits)
	}

	start = time.Now()
	for i := 0; i < 10_000; i++ {
		masked := empty_board_attack_bb & (kiwipete.Pieces[0] | kiwipete.Pieces[1])
		idx := (masked * rookMagics[square]) >> (64 - 10)
		x = rookBB[square][idx]
	}
	elapsed = time.Since(start)
	fmt.Println("elapsed 2:", elapsed)
	fmt.Println("moves:", x)
	fmt.Println("moves:", pieces.Read1Bits(x))
}

func buildRookMagicBB() [64][4096]uint64 {
	// WIP rook pre-calculated magic bitboard attacks
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
		empty_board_attack_bb := pieces.GetRookMoves(square, cb)
		for _, line := range [4]uint64{rank_1, rank_8, file_a, file_h} {
			// if square not in the rank/file
			if square_bb|line != line {
				empty_board_attack_bb &= ^line
			}
		}
		count_1_bits := bits.OnesCount64(empty_board_attack_bb)

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
			idx := (masked * rookMagics[square]) >> (64 - count_1_bits)
			rookAttackBBs[square][idx] = pieces.GetRookMoves(square, cb)
		}
	}

	return rookAttackBBs
}
