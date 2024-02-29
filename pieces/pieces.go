// Move generation
package pieces

import (
	"bufio"
	"engine2/board"
	"fmt"
	"log"
	_ "math"
	"math/bits"
	"os"
	"strings"
)

/*
bb = bitboard, cb = chessboard
Magic numbers 0, ..., 63 and 1<<0, ..., 1<<63 are squares of the chessboard.
*/

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
var rookMagics = [64]uint64{
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

var bishopMagics = [64]uint64{
	0x100420000431024, 0x280800101073404, 0x42000a00840802, 0xca800c0410c2,
	0x81004290941c20, 0x400200450020250, 0x444a019204022084, 0x88610802202109a,
	0x11210a0800086008, 0x400a08c08802801, 0x1301a0500111c808, 0x1280100480180404,
	0x720009020028445, 0x91880a9000010a01, 0x31200940150802b2, 0x5119080c20000602,
	0x242400a002448023, 0x4819006001200008, 0x222c10400020090, 0x302008420409004,
	0x504200070009045, 0x210071240c02046, 0x1182219000022611, 0x400c50000005801,
	0x4004010000113100, 0x2008121604819400, 0xc4a4010000290101, 0x404a000888004802,
	0x8820c004105010, 0x28280100908300, 0x4c013189c0320a80, 0x42008080042080,
	0x90803000c080840, 0x2180001028220, 0x1084002a040036, 0x212009200401,
	0x128110040c84a84, 0x81488020022802, 0x8c0014100181, 0x2222013020082,
	0xa00100002382c03, 0x1000280001005c02, 0x84801010000114c, 0x480410048000084,
	0x21204420080020a, 0x2020010000424a10, 0x240041021d500141, 0x420844000280214,
	0x29084a280042108, 0x84102a8080a20a49, 0x104204908010212, 0x40a20280081860c1,
	0x3044000200121004, 0x1001008807081122, 0x50066c000210811, 0xe3001240f8a106,
	0x940c0204030020d4, 0x619204000210826a, 0x2010438002b00a2, 0x884042004005802,
	0xa90240000006404, 0x500d082244010008, 0x28190d00040014e0, 0x825201600c082444,
}

var rookAttackBBs [64][4096]uint64 = buildRookMagicBB()
var rookRelevantOccs [64]uint64
var rookOneBitCounts [64]int

var bishopAttackBBs [64][512]uint64 = buildBishopMagicBB()
var bishopRelevantOccs [64]uint64
var bishopOneBitCounts [64]int

func MovePiece(move board.Move, cb *board.Board) {
	// TODO: Refactor to remove switch. Maybe make a parent array board.Occupied
	fromBB := uint64(1 << move.From)
	toBB := uint64(1 << move.To)

	if toBB&(cb.Pieces[1^cb.WToMove]^cb.Kings[1^cb.WToMove]) != 0 {
		capturePiece(toBB, cb)
	}

	cb.Pieces[cb.WToMove] ^= fromBB + toBB
	switch move.Piece {
	case "p":
		cb.Pawns[cb.WToMove] ^= fromBB + toBB
		if move.To-move.From == 16 || move.To-move.From == -16 {
			cb.EpSquare = (move.To + move.From) / 2
		} else if move.To < 8 || move.To > 55 {
			promotePawn(toBB, cb, move.PromoteTo)
			cb.EpSquare = 100
		} else if move.To == cb.EpSquare {
			captureSq := move.To + 8
			if cb.WToMove == 1 {
				captureSq = move.To - 8
			}
			cb.Pawns[1^cb.WToMove] ^= uint64(1 << captureSq)
			cb.Pieces[1^cb.WToMove] ^= uint64(1 << captureSq)
			cb.EpSquare = 100
		} else {
			cb.EpSquare = 100
		}
	case "n":
		cb.Knights[cb.WToMove] ^= fromBB + toBB
		cb.EpSquare = 100
	case "b":
		cb.Bishops[cb.WToMove] ^= fromBB + toBB
		cb.EpSquare = 100
	case "r":
		cb.Rooks[cb.WToMove] ^= fromBB + toBB
		if move.From == 0 || move.From == 56 {
			cb.CastleRights[cb.WToMove][0] = false
		} else if move.From == 7 || move.From == 63 {
			cb.CastleRights[cb.WToMove][1] = false
		}
		cb.EpSquare = 100
	case "q":
		cb.Queens[cb.WToMove] ^= fromBB + toBB
		cb.EpSquare = 100
	case "k":
		if move.To-move.From == 2 || move.To-move.From == -2 {
			if cb.CastleRights[cb.WToMove][0] && (move.To == 2 || move.To == 58) {
				cb.Rooks[cb.WToMove] ^= uint64(1<<(move.To-2) + 1<<(move.To+1))
				cb.Pieces[cb.WToMove] ^= uint64(1<<(move.To-2) + 1<<(move.To+1))
			} else if cb.CastleRights[cb.WToMove][1] && (move.To == 6 || move.To == 62) {
				cb.Rooks[cb.WToMove] ^= uint64(1<<(move.To+1) + 1<<(move.To-1))
				cb.Pieces[cb.WToMove] ^= uint64(1<<(move.To+1) + 1<<(move.To-1))
			} else {
				panic("king moving two squares, but is not castling")
			}
		}
		cb.Kings[cb.WToMove] ^= fromBB + toBB
		cb.KingSqs[cb.WToMove] = move.To
		cb.CastleRights[cb.WToMove][0] = false
		cb.CastleRights[cb.WToMove][1] = false
		cb.EpSquare = 100
	default:
		panic("empty or invalid piece type")
	}

	cb.PrevMove = move
	cb.WToMove ^= 1
}

func capturePiece(squareBB uint64, cb *board.Board) {
	opponent := 1 ^ cb.WToMove
	cb.Pieces[opponent] ^= squareBB

	switch {
	case squareBB&cb.Pawns[opponent] != 0:
		cb.Pawns[opponent] ^= squareBB
	case squareBB&cb.Knights[opponent] != 0:
		cb.Knights[opponent] ^= squareBB
	case squareBB&cb.Bishops[opponent] != 0:
		cb.Bishops[opponent] ^= squareBB
	case squareBB&cb.Rooks[opponent] != 0:
		// int type mixing here seems ok based on investigation
		if opponent == 0 && squareBB == 1<<56 {
			cb.CastleRights[opponent][0] = false
		} else if opponent == 0 && squareBB == 1<<63 {
			cb.CastleRights[opponent][1] = false
		} else if opponent == 1 && squareBB == 0 {
			cb.CastleRights[opponent][0] = false
		} else if opponent == 1 && squareBB == 1<<7 {
			cb.CastleRights[opponent][1] = false
		}
		cb.Rooks[opponent] ^= squareBB
	case squareBB&cb.Queens[opponent] != 0:
		cb.Queens[opponent] ^= squareBB
	default:
		panic("no captured piece bitboard matches")
	}
}

func promotePawn(toBB uint64, cb *board.Board, promoteTo ...string) {
	// TODO: Else never triggers b/c move.promoteTo always has a string
	// Change to 'if promotoTo != ""'
	if len(promoteTo) == 1 {
		switch {
		case promoteTo[0] == "q":
			cb.Queens[cb.WToMove] ^= toBB
		case promoteTo[0] == "n":
			cb.Knights[cb.WToMove] ^= toBB
		case promoteTo[0] == "b":
			cb.Bishops[cb.WToMove] ^= toBB
		case promoteTo[0] == "r":
			cb.Rooks[cb.WToMove] ^= toBB
		default:
			panic("invalid promoteTo")
		}
	} else {
		fmt.Print("promote pawn to N, B, R, or Q: ")
		userPromote := getUserInput()

		if userPromote == "q" || userPromote == "n" || userPromote == "b" || userPromote == "r" {
			promotePawn(toBB, cb, userPromote)
		} else {
			fmt.Println("invalid promotion type, try again")
			promotePawn(toBB, cb)
		}
	}

	cb.Pawns[cb.WToMove] ^= toBB
}

func getUserInput() string {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	err := scanner.Err()
	if err != nil {
		log.Println("failed to get input:", err)
		return getUserInput()
	}
	return strings.ToLower(scanner.Text())
}

// Use for user-submitted moves only?
// Checks for blocking pieces and disallows captures of friendly pieces.
// Does not consider check, pins, or legality of a pawn movement direction.
func isValidMove(from, to int, pieceType string, cb *board.Board) bool {
	if from < 0 || from > 63 || to < 0 || to > 63 || to == from {
		return false
	}
	toBB := uint64(1 << to)
	diff := to - from
	// to == from already excluded, no 0 move bugs from pawnDirections.
	pawnDirections := [2][4]int{{-7, -8, -9, -16},
		{7, 8, 9, 16},
	}

	switch pieceType {
	case "p":
		if !board.ContainsN(diff, pawnDirections[cb.WToMove]) {
			return false
		}
	case "n":
		if toBB&cb.NAttacks[from] == 0 {
			return false
		}
	case "b":
		if toBB&calculateBishopMoves(from, cb) == 0 {
			return false
		}
	case "r":
		if toBB&lookupRookMoves(from, cb) == 0 {
			return false
		}
	case "q":
		if toBB&(lookupRookMoves(from, cb)|calculateBishopMoves(from, cb)) == 0 {
			return false
		}
	case "k":
		cb.Pieces[cb.WToMove] ^= uint64(1 << cb.KingSqs[cb.WToMove])
		cb.WToMove ^= 1
		attkSquares := getAttackedSquares(cb)
		cb.WToMove ^= 1
		cb.Pieces[cb.WToMove] ^= uint64(1 << cb.KingSqs[cb.WToMove])
		if toBB&getKingMoves(from, attkSquares, cb) == 0 {
			return false
		}
	default:
		// pieceType is not valid
		return false
	}

	// Friendly piece collision
	if toBB&cb.Pieces[cb.WToMove] != 0 {
		return false
	}

	return true
}

// Captures and protection are included in move gen.
func calculateRookMoves(square int, cb *board.Board) uint64 {
	occupied := cb.Pieces[0] | cb.Pieces[1]
	// North
	moves := cb.SlidingAttacks[0][square]
	blockers := cb.SlidingAttacks[0][square] & occupied
	blockerSq := bits.TrailingZeros64(blockers | uint64(1<<63))
	moves ^= cb.SlidingAttacks[0][blockerSq]
	// East
	moves |= cb.SlidingAttacks[2][square]
	blockers = cb.SlidingAttacks[2][square] & occupied
	blockerSq = bits.TrailingZeros64(blockers | uint64(1<<63))
	moves ^= cb.SlidingAttacks[2][blockerSq]
	// South
	moves |= cb.SlidingAttacks[4][square]
	blockers = cb.SlidingAttacks[4][square] & occupied
	blockerSq = 63 - bits.LeadingZeros64(blockers|uint64(1))
	moves ^= cb.SlidingAttacks[4][blockerSq]
	// West
	moves |= cb.SlidingAttacks[6][square]
	blockers = cb.SlidingAttacks[6][square] & occupied
	blockerSq = 63 - bits.LeadingZeros64(blockers|uint64(1))
	moves ^= cb.SlidingAttacks[6][blockerSq]

	return moves
}

func lookupRookMoves(square int, cb *board.Board) uint64 {
	occupied := cb.Pieces[0] | cb.Pieces[1]
	masked_blockers := rookRelevantOccs[square] & occupied
	idx := (masked_blockers * rookMagics[square]) >> (64 - rookOneBitCounts[square])
	// Do not exclude piece protection (no `& ^cb.Pieces[cb.WToMove]`)
	return rookAttackBBs[square][idx]
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
		cb.Pieces[0] = 0
		empty_board_attack_bb := calculateRookMoves(square, cb)
		for _, line := range [4]uint64{rank_1, rank_8, file_a, file_h} {
			// if square not in the rank/file
			if square_bb|line != line {
				empty_board_attack_bb &= ^line
			}
		}
		count_1_bits := bits.OnesCount64(empty_board_attack_bb)

		rookRelevantOccs[square] = empty_board_attack_bb
		rookOneBitCounts[square] = count_1_bits

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
			fmt.Println("perm_idx != possible occupancies (", perm_idx, "!=", possible_occupancies_count, ")")
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
			rookAttackBBs[square][idx] = calculateRookMoves(square, cb)
		}
		cb.Pieces[0] = 0
	}

	return rookAttackBBs
}

func getPawnMoves(square int, cb *board.Board) uint64 {
	opponent := 1 ^ cb.WToMove

	if square < 8 || square > 55 {
		panic("pawns can't be on the first or last rank")
	}

	moves := cb.PAttacks[cb.WToMove][square] & (cb.Pieces[opponent] | uint64(1<<cb.EpSquare))

	var dir, low, high int
	if cb.WToMove == 1 {
		dir = 8
		low = 7
		high = 16
	} else {
		dir = -8
		low = 47
		high = 56
	}
	occupied := cb.Pieces[0] | cb.Pieces[1]

	if low < square && square < high && 1<<(square+dir)&occupied == 0 {
		moves |= (1<<(square+dir) + 1<<(square+2*dir)) & ^occupied
	} else {
		moves |= 1 << (square + dir) & ^occupied
	}

	return moves
}

func getKnightMoves(square int, cb *board.Board) uint64 {
	return cb.NAttacks[square]
}

func calculateBishopMoves(square int, cb *board.Board) uint64 {
	occupied := cb.Pieces[0] | cb.Pieces[1]
	// Northeast
	moves := cb.SlidingAttacks[1][square]
	blockers := cb.SlidingAttacks[1][square] & occupied
	blockerSq := bits.TrailingZeros64(blockers | uint64(1<<63))
	moves ^= cb.SlidingAttacks[1][blockerSq]
	// Southeast
	moves |= cb.SlidingAttacks[3][square]
	blockers = cb.SlidingAttacks[3][square] & occupied
	blockerSq = 63 - bits.LeadingZeros64(blockers|uint64(1))
	moves ^= cb.SlidingAttacks[3][blockerSq]
	// Southwest
	moves |= cb.SlidingAttacks[5][square]
	blockers = cb.SlidingAttacks[5][square] & occupied
	blockerSq = 63 - bits.LeadingZeros64(blockers|uint64(1))
	moves ^= cb.SlidingAttacks[5][blockerSq]
	// Northwest
	moves |= cb.SlidingAttacks[7][square]
	blockers = cb.SlidingAttacks[7][square] & occupied
	blockerSq = bits.TrailingZeros64(blockers | uint64(1<<63))
	moves ^= cb.SlidingAttacks[7][blockerSq]

	return moves
}

func lookupBishopMoves(square int, cb *board.Board) uint64 {
	occupied := cb.Pieces[0] | cb.Pieces[1]
	masked_blockers := bishopRelevantOccs[square] & occupied
	idx := (masked_blockers * bishopMagics[square]) >> (64 - bishopOneBitCounts[square])
	// Do not exclude piece protection (no `& ^cb.Pieces[cb.WToMove]`)
	return bishopAttackBBs[square][idx]
}

func buildBishopMagicBB() [64][512]uint64 {
	// WIP bishop pre-calculated magic bitboard attacks
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
		empty_board_attack_bb := calculateBishopMoves(square, cb)
		for _, line := range [4]uint64{rank_1, rank_8, file_a, file_h} {
			// if square not in the rank/file
			if square_bb|line != line {
				empty_board_attack_bb &= ^line
			}
		}
		count_1_bits := bits.OnesCount64(empty_board_attack_bb)

		bishopRelevantOccs[square] = empty_board_attack_bb
		bishopOneBitCounts[square] = count_1_bits

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
			fmt.Println("perm_idx != possible occupancies (", perm_idx, "!=", possible_occupancies_count, ")")
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
			idx := (masked * bishopMagics[square]) >> (64 - count_1_bits)
			bishopAttackBBs[square][idx] = calculateBishopMoves(square, cb)
		}
		cb.Pieces[0] = 0
	}

	return bishopAttackBBs
}

func getQueenMoves(square int, cb *board.Board) uint64 {
	return lookupRookMoves(square, cb) | calculateBishopMoves(square, cb)
	//return lookupRookMoves(square, cb) | lookupBishopMoves(square, cb)
}

// Return legal king moves.
func getKingMoves(square int, oppAttackedSquares uint64, cb *board.Board) uint64 {
	occupied := cb.Pieces[0] | cb.Pieces[1]
	moves := cb.KAttacks[square] & ^oppAttackedSquares & ^cb.Pieces[cb.WToMove]

	if cb.WToMove == 0 {
		if cb.CastleRights[0][0] && (1<<57+1<<58+1<<59)&occupied == 0 &&
			(1<<58+1<<59+1<<60)&oppAttackedSquares == 0 {
			moves += 1 << 58
		}
		if cb.CastleRights[0][1] && (1<<61+1<<62)&occupied == 0 &&
			(1<<60+1<<61+1<<62)&oppAttackedSquares == 0 {
			moves += 1 << 62
		}
	} else {
		if cb.CastleRights[1][0] && (1<<1+1<<2+1<<3)&occupied == 0 &&
			(1<<2+1<<3+1<<4)&oppAttackedSquares == 0 {
			moves += 1 << 2
		}
		if cb.CastleRights[1][1] && (1<<5+1<<6)&occupied == 0 &&
			(1<<4+1<<5+1<<6)&oppAttackedSquares == 0 {
			moves += 1 << 6
		}
	}

	return moves
}

// Return the set of squares attacked by color cb.WToMove
func getAttackedSquares(cb *board.Board) uint64 {
	// TODO: Is there a way to avoid reading 1 bits when accumulating moves?
	var pieces []int
	attackSquares := uint64(0)

	pieces = Read1BitsPawns(cb.Pawns[cb.WToMove])
	for _, square := range pieces {
		// Do not include pawn pushes.
		attackSquares |= cb.PAttacks[cb.WToMove][square]
	}

	pieces = read1Bits(cb.Knights[cb.WToMove])
	for _, square := range pieces {
		attackSquares |= cb.NAttacks[square]
	}
	pieces = read1Bits(cb.Bishops[cb.WToMove])
	for _, square := range pieces {
		attackSquares |= calculateBishopMoves(square, cb)
	}
	pieces = read1Bits(cb.Rooks[cb.WToMove])
	for _, square := range pieces {
		attackSquares |= lookupRookMoves(square, cb)
	}
	pieces = read1Bits(cb.Queens[cb.WToMove])
	for _, square := range pieces {
		attackSquares |= getQueenMoves(square, cb)
	}
	// Do not include castling.
	attackSquares |= cb.KAttacks[cb.KingSqs[cb.WToMove]]

	return attackSquares
}

type moveGenFunc func(int, *board.Board) uint64
type readBitsFunc func(uint64) []int

func GetAllMoves(cb *board.Board) []board.Move {
	// Return slice of all pseudo-legal moves for color cb.WToMove (king moves
	// are strictly legal)
	cb.Pieces[cb.WToMove] ^= uint64(1 << cb.KingSqs[cb.WToMove])
	cb.WToMove ^= 1
	// TODO: could inline getAttackedSquares() here to prevent 2x calls to read1Bits()
	attackedSquares := getAttackedSquares(cb)
	cb.WToMove ^= 1
	cb.Pieces[cb.WToMove] ^= uint64(1 << cb.KingSqs[cb.WToMove])

	var capturesBlks uint64
	var attackerCount int
	if cb.Kings[cb.WToMove]&attackedSquares != 0 {
		capturesBlks, attackerCount = getCheckingSquares(cb)
	}

	kingSq := cb.KingSqs[cb.WToMove]
	moves := read1Bits(getKingMoves(kingSq, attackedSquares, cb) & ^cb.Pieces[cb.WToMove])

	allMoves := make([]board.Move, len(moves), 35)
	for i, toSq := range moves {
		allMoves[i] = board.Move{From: kingSq, To: toSq, Piece: "k", PromoteTo: ""}
	}
	// If attackerCount > 1 and king has no moves, it is checkmate
	if attackerCount > 1 {
		return allMoves
	}

	pieces := []uint64{cb.Pawns[cb.WToMove], cb.Knights[cb.WToMove],
		cb.Bishops[cb.WToMove], cb.Rooks[cb.WToMove], cb.Queens[cb.WToMove],
	}
	moveFuncs := []moveGenFunc{getPawnMoves, getKnightMoves, calculateBishopMoves,
		lookupRookMoves, getQueenMoves,
	}
	symbols := []string{"p", "n", "b", "r", "q"}

	// 29% perft() speed up and -40% malloc from having this loop in this function
	for i, piece := range pieces {
		for _, fromSq := range read1Bits(piece) {
			moves := read1Bits(moveFuncs[i](fromSq, cb) & ^cb.Pieces[cb.WToMove])
			for _, toSq := range moves {
				if i == 0 && (toSq < 8 || toSq > 55) &&
					(capturesBlks == 0 || uint64(1<<toSq)&capturesBlks != 0) {
					allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: "p", PromoteTo: "n"})
					allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: "p", PromoteTo: "b"})
					allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: "p", PromoteTo: "r"})
					allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: "p", PromoteTo: "q"})
				} else if capturesBlks == 0 || uint64(1<<toSq)&capturesBlks != 0 {
					allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: symbols[i], PromoteTo: ""})
				}
			}
		}
	}

	return allMoves
}

// Return the set of squares of pieces checking the king and interposition
// squares, and the number of checking pieces.
func getCheckingSquares(cb *board.Board) (uint64, int) {
	opponent := 1 ^ cb.WToMove
	attackerCount := 0

	kSquare := cb.KingSqs[cb.WToMove]
	pAttackers := cb.PAttacks[cb.WToMove][kSquare] & cb.Pawns[opponent]
	knightAttackers := cb.NAttacks[kSquare] & cb.Knights[opponent]
	bqAttackers := calculateBishopMoves(kSquare, cb) & (cb.Bishops[opponent] |
		cb.Queens[opponent])
	orthogAttackers := lookupRookMoves(cb.KingSqs[cb.WToMove], cb) &
		(cb.Rooks[opponent] | cb.Queens[opponent])

		// TODO: Remove king check?
	if cb.Kings[opponent]&cb.KAttacks[cb.KingSqs[cb.WToMove]] != 0 {
		fmt.Println(cb.KingSqs)
		cb.Print()
		panic("king is checking the other king")
	}
	if bits.OnesCount64(knightAttackers) > 1 {
		cb.Print()
		panic(">1 knights are checking the king")
	}

	// There should be 0 or 1 attackers in each attack group.
	if pAttackers != 0 {
		attackerCount += 1
	}
	if knightAttackers != 0 {
		attackerCount += 1
	}

	panicMsgs := []string{">1 piece is checking king orthogonally",
		">1 piece is checking king diagonally"}
	attackers := []uint64{orthogAttackers, bqAttackers}

	// Add interposition squares if any exist.
	for i, attacker := range attackers {
		if attacker != 0 {
			attackerSquares := read1Bits(attacker)
			attackerCount += len(attackerSquares)
			// Possible optimization: check if attackerCount + len(attackers) > 1 before the loop
			if len(attackerSquares) > 1 {
				panic(panicMsgs[i])
			}
			dir := findDirection(cb.KingSqs[cb.WToMove], attackerSquares[0])
			attackers[i] = fillFromTo(cb.KingSqs[cb.WToMove], attackerSquares[0], dir)
		}
	}

	return pAttackers | knightAttackers | attackers[0] | attackers[1], attackerCount
}

// Return a bitboard of squares between `from and `to`, excluding `from`
// and including `to`
func fillFromTo(from, to, direction int) uint64 {
	bb := uint64(0)
	for sq := from + direction; sq != to; sq += direction {
		bb += 1 << sq
	}
	bb += 1 << to

	return bb
}

// Return the direction from one square to another. Assumes (from, to) is an
// orthogonal or diagonal move
func findDirection(from, to int) int {
	var dir int
	diff := to - from
	// TODO: Change to lookup table for files, or use bitboards.
	files := board.GetFiles()

	switch {
	case diff%8 == 0:
		dir = 8
	case diff%9 == 0:
		dir = 9
	case -6 <= diff && diff <= 6:
		dir = 1
	case diff%7 == 0:
		fromInAFile := board.ContainsN(from, files[0])
		fromInHFile := board.ContainsN(from, files[3])
		toInAFile := board.ContainsN(to, files[0])
		toInHFile := board.ContainsN(to, files[3])
		if (fromInAFile && toInHFile) || (fromInHFile && toInAFile) {
			dir = 1
		} else {
			dir = 7
		}
	default:
		panic("invalid toSquare-fromSquare difference")
	}
	if diff < 0 {
		dir *= -1
	}

	return dir
}

func read1Bits(bb uint64) []int {
	// Using TrailingZeros64() seems as fast as bitshifting right while bb>0.
	squares := make([]int, 0, 4)
	for bb > 0 {
		squares = append(squares, bits.TrailingZeros64(bb))
		bb &= bb - 1
	}
	return squares
}

var Read1Bits = read1Bits

func Read1BitsPawns(bb uint64) []int {
	squares := make([]int, 0, 8)
	for bb > 0 {
		squares = append(squares, bits.TrailingZeros64(bb))
		bb &= bb - 1
	}
	return squares
}

func binSearch(n int, nums [8]int) bool {
	l := 0
	r := len(nums) - 1
	var mid int
	for l <= r {
		mid = (l + r) / 2
		switch {
		case n > nums[mid]:
			l = mid + 1
		case n < nums[mid]:
			r = mid - 1
		default:
			return true
		}
	}

	return false
}
