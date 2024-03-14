// Move generation
package pieces

import (
	"bufio"
	"engine2/board"
	"fmt"
	"log"
	"math/bits"
	"os"
	"strings"
)

/*
bb = bitboard, cb = chessboard
Magic numbers 0, ..., 63 and 1<<0, ..., 1<<63 are squares of the chessboard.
*/

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

var rookRelevantOccs [64]uint64
var rookOneBitCounts [64]int
var rookAttackBBs [64][4096]uint64 = buildRookMagicBB()

var bishopRelevantOccs [64]uint64
var bishopOneBitCounts [64]int
var bishopAttackBBs [64][512]uint64 = buildBishopMagicBB()

const (
	PAWN     = uint8(0)
	KNIGHT   = uint8(1)
	BISHOP   = uint8(2)
	ROOK     = uint8(3)
	QUEEN    = uint8(4)
	KING     = uint8(5)
	NO_PIECE = uint8(9)
)

func MovePiece(move board.Move, cb *board.Board) {
	// TODO: Refactor to remove switch?
	fromBB := uint64(1 << move.From)
	toBB := uint64(1 << move.To)
	if cb.EpSquare != 100 {
		cb.Zobrist ^= board.ZobristKeys.EpFile[cb.EpSquare%8]
	}

	if toBB&(cb.Pieces[1^cb.WToMove]^cb.Kings[1^cb.WToMove]) != 0 {
		capturePiece(toBB, move.To, cb)
	}

	cb.Pieces[cb.WToMove] ^= fromBB + toBB

	switch move.Piece {
	case PAWN:
		cb.Pawns[cb.WToMove] ^= fromBB + toBB
		if move.To-move.From == 16 || move.To-move.From == -16 {
			cb.EpSquare = (move.To + move.From) / 2
			cb.Zobrist ^= board.ZobristKeys.EpFile[cb.EpSquare%8]
			cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][0][move.To]
		} else if move.To < 8 || move.To > 55 {
			promotePawn(toBB, move.To, cb, move.PromoteTo)
			cb.EpSquare = 100
		} else if move.To == cb.EpSquare {
			captureSq := move.To + 8
			if cb.WToMove == 1 {
				captureSq = move.To - 8
			}
			cb.Pawns[1^cb.WToMove] ^= uint64(1 << captureSq)
			cb.Pieces[1^cb.WToMove] ^= uint64(1 << captureSq)
			cb.EpSquare = 100
			cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][0][move.To]
		} else {
			cb.EpSquare = 100
			cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][0][move.To]
		}
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][0][move.From]
		cb.HalfMoves = 1
	case KNIGHT:
		cb.Knights[cb.WToMove] ^= fromBB + toBB
		cb.EpSquare = 100
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][1][move.From]
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][1][move.To]
	case BISHOP:
		cb.Bishops[cb.WToMove] ^= fromBB + toBB
		cb.EpSquare = 100
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][2][move.From]
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][2][move.To]
	case ROOK:
		cb.Rooks[cb.WToMove] ^= fromBB + toBB
		if move.From == 0 || move.From == 56 {
			if cb.CastleRights[cb.WToMove][0] == true {
				cb.Zobrist ^= board.ZobristKeys.Castle[cb.WToMove][0]
			}
			cb.CastleRights[cb.WToMove][0] = false
		} else if move.From == 7 || move.From == 63 {
			if cb.CastleRights[cb.WToMove][1] == true {
				cb.Zobrist ^= board.ZobristKeys.Castle[cb.WToMove][1]
			}
			cb.CastleRights[cb.WToMove][1] = false
		}
		cb.EpSquare = 100
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][3][move.From]
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][3][move.To]
	case QUEEN:
		cb.Queens[cb.WToMove] ^= fromBB + toBB
		cb.EpSquare = 100
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][4][move.From]
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][4][move.To]
	case KING:
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
		if cb.CastleRights[cb.WToMove][0] == true {
			cb.Zobrist ^= board.ZobristKeys.Castle[cb.WToMove][0]
		}
		if cb.CastleRights[cb.WToMove][1] == true {
			cb.Zobrist ^= board.ZobristKeys.Castle[cb.WToMove][1]
		}
		cb.CastleRights[cb.WToMove][0] = false
		cb.CastleRights[cb.WToMove][1] = false
		cb.EpSquare = 100
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][5][move.From]
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][5][move.To]
	case NO_PIECE:
		break
	default:
		panic("empty or invalid piece type")
	}

	cb.PrevMove = move
	cb.WToMove ^= 1
	cb.Zobrist ^= board.ZobristKeys.BToMove
	cb.HalfMoves += 1
}

func capturePiece(squareBB uint64, square int8, cb *board.Board) {
	opponent := 1 ^ cb.WToMove
	cb.Pieces[opponent] ^= squareBB
	cb.HalfMoves = 1

	switch {
	case squareBB&cb.Pawns[opponent] != 0:
		cb.Pawns[opponent] ^= squareBB
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[opponent][0][square]
	case squareBB&cb.Knights[opponent] != 0:
		cb.Knights[opponent] ^= squareBB
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[opponent][1][square]
	case squareBB&cb.Bishops[opponent] != 0:
		cb.Bishops[opponent] ^= squareBB
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[opponent][2][square]
	case squareBB&cb.Rooks[opponent] != 0:
		// int type mixing here seems ok based on investigation
		if opponent == 0 && squareBB == 1<<56 {
			cb.CastleRights[opponent][0] = false
			cb.Zobrist ^= board.ZobristKeys.Castle[opponent][0]
		} else if opponent == 0 && squareBB == 1<<63 {
			cb.CastleRights[opponent][1] = false
			cb.Zobrist ^= board.ZobristKeys.Castle[opponent][1]
		} else if opponent == 1 && squareBB == 0 {
			cb.CastleRights[opponent][0] = false
			cb.Zobrist ^= board.ZobristKeys.Castle[opponent][0]
		} else if opponent == 1 && squareBB == 1<<7 {
			cb.CastleRights[opponent][1] = false
			cb.Zobrist ^= board.ZobristKeys.Castle[opponent][1]
		}
		cb.Rooks[opponent] ^= squareBB
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[opponent][3][square]
	case squareBB&cb.Queens[opponent] != 0:
		cb.Queens[opponent] ^= squareBB
		cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[opponent][4][square]
	default:
		panic("no captured piece bitboard matches")
	}
}

func promotePawn(toBB uint64, square int8, cb *board.Board, promoteTo ...uint8) {
	// TODO: Else never triggers b/c move.promoteTo always has a string
	// Change to 'if promotoTo != ""'
	if len(promoteTo) == 1 {
		switch {
		case promoteTo[0] == QUEEN:
			cb.Queens[cb.WToMove] ^= toBB
			cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][4][square]
		case promoteTo[0] == KNIGHT:
			cb.Knights[cb.WToMove] ^= toBB
			cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][1][square]
		case promoteTo[0] == BISHOP:
			cb.Bishops[cb.WToMove] ^= toBB
			cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][2][square]
		case promoteTo[0] == ROOK:
			cb.Rooks[cb.WToMove] ^= toBB
			cb.Zobrist ^= board.ZobristKeys.ColorPieceSq[cb.WToMove][3][square]
		default:
			panic("invalid promoteTo")
		}
	} else {
		fmt.Print("promote pawn to N, B, R, or Q: ")
		userPromote := getUserInput()

		if userPromote == QUEEN || userPromote == KNIGHT || userPromote == BISHOP ||
			userPromote == ROOK {
			promotePawn(toBB, square, cb, userPromote)
		} else {
			fmt.Println("invalid promotion type, try again")
			promotePawn(toBB, square, cb)
		}
	}

	cb.Pawns[cb.WToMove] ^= toBB
}

func getUserInput() uint8 {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	err := scanner.Err()
	if err != nil {
		log.Println("failed to get input:", err)
		return getUserInput()
	}
	// TODO: make sure this returns the first rune, not the first byte
	var piece uint8
	switch strings.ToLower(scanner.Text())[0] {
	case byte('n'):
		piece = KNIGHT
	case byte('b'):
		piece = BISHOP
	case byte('r'):
		piece = ROOK
	case byte('q'):
		piece = QUEEN
	default:
		return getUserInput()
	}
	return piece
}

// Use for user-submitted moves only?
// Checks for blocking pieces and disallows captures of friendly pieces.
// Does not consider check, pins, or legality of a pawn movement direction.
func isValidMove(from, to int8, pieceType string, cb *board.Board) bool {
	if from < 0 || from > 63 || to < 0 || to > 63 || to == from {
		return false
	}
	toBB := uint64(1 << to)
	diff := to - from
	// to == from already excluded, no 0 move bugs from pawnDirections.
	pawnDirections := [2][4]int8{{-7, -8, -9, -16},
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
		if toBB&lookupBishopMoves(from, cb) == 0 {
			return false
		}
	case "r":
		if toBB&lookupRookMoves(from, cb) == 0 {
			return false
		}
	case "q":
		if toBB&(lookupRookMoves(from, cb)|lookupBishopMoves(from, cb)) == 0 {
			return false
		}
	case "k":
		cb.Pieces[cb.WToMove] ^= uint64(1 << cb.KingSqs[cb.WToMove])
		cb.WToMove ^= 1
		attkSquares := GetAttackedSquares(cb)
		cb.WToMove ^= 1
		cb.Pieces[cb.WToMove] ^= uint64(1 << cb.KingSqs[cb.WToMove])
		if toBB&GetKingMoves(from, attkSquares, cb) == 0 {
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

func lookupRookMoves(square int8, cb *board.Board) uint64 {
	occupied := cb.Pieces[0] | cb.Pieces[1]
	masked_blockers := rookRelevantOccs[square] & occupied
	idx := (masked_blockers * rookMagics[square]) >> (64 - rookOneBitCounts[square])
	// Do not exclude piece protection (no `& ^cb.Pieces[cb.WToMove]`)
	return rookAttackBBs[square][idx]
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

func GetPawnMoves(square int8, cb *board.Board) uint64 {
	opponent := 1 ^ cb.WToMove

	if square < 8 || square > 55 {
		panic("pawns can't be on the first or last rank")
	}

	moves := cb.PAttacks[cb.WToMove][square] & (cb.Pieces[opponent] | uint64(1<<cb.EpSquare))

	var dir, low, high int8
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

func getKnightMoves(square int8, cb *board.Board) uint64 {
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

func lookupBishopMoves(square int8, cb *board.Board) uint64 {
	occupied := cb.Pieces[0] | cb.Pieces[1]
	masked_blockers := bishopRelevantOccs[square] & occupied
	idx := (masked_blockers * bishopMagics[square]) >> (64 - bishopOneBitCounts[square])
	// Do not exclude piece protection (no `& ^cb.Pieces[cb.WToMove]`)
	return bishopAttackBBs[square][idx]
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

func getQueenMoves(square int8, cb *board.Board) uint64 {
	return lookupRookMoves(square, cb) | lookupBishopMoves(square, cb)
}

// Return legal king moves.
func GetKingMoves(square int8, oppAttackedSquares uint64, cb *board.Board) uint64 {
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
func GetAttackedSquares(cb *board.Board) uint64 {
	// TODO: Is there a way to avoid reading 1 bits when accumulating moves?
	attackSquares := uint64(0)

	bb := cb.Pawns[cb.WToMove]
	for bb > 0 {
		attackSquares |= cb.PAttacks[cb.WToMove][bits.TrailingZeros64(bb)]
		bb &= bb - 1
	}
	bb = cb.Knights[cb.WToMove]
	for bb > 0 {
		attackSquares |= cb.NAttacks[bits.TrailingZeros64(bb)]
		bb &= bb - 1
	}
	bb = cb.Bishops[cb.WToMove]
	for bb > 0 {
		attackSquares |= lookupBishopMoves(int8(bits.TrailingZeros64(bb)), cb)
		bb &= bb - 1
	}
	bb = cb.Rooks[cb.WToMove]
	for bb > 0 {
		attackSquares |= lookupRookMoves(int8(bits.TrailingZeros64(bb)), cb)
		bb &= bb - 1
	}
	bb = cb.Queens[cb.WToMove]
	for bb > 0 {
		attackSquares |= getQueenMoves(int8(bits.TrailingZeros64(bb)), cb)
		bb &= bb - 1
	}
	// Do not include castling.
	attackSquares |= cb.KAttacks[cb.KingSqs[cb.WToMove]]

	return attackSquares
}

type moveGenFunc func(int8, *board.Board) uint64

// Return slice of all pseudo-legal moves for color cb.WToMove, where any king
// moves are strictly legal. However, if the king is in check, only legal moves
// are returned
func GetAllMoves(cb *board.Board) []board.Move {
	cb.Pieces[cb.WToMove] ^= 1 << cb.KingSqs[cb.WToMove]
	cb.WToMove ^= 1
	attackedSquares := GetAttackedSquares(cb)
	cb.WToMove ^= 1
	cb.Pieces[cb.WToMove] ^= 1 << cb.KingSqs[cb.WToMove]

	var capturesBlks uint64
	var attackerCount int
	if cb.Kings[cb.WToMove]&attackedSquares != 0 {
		capturesBlks, attackerCount = GetCheckingSquares(cb)
	}

	// TODO: Trying to use a global allMoves did not work well
	allMoves := make([]board.Move, 0, 35)
	kingSq := cb.KingSqs[cb.WToMove]
	kingMovesBB := GetKingMoves(kingSq, attackedSquares, cb) & ^cb.Pieces[cb.WToMove]

	var toSq int8
	for kingMovesBB > 0 {
		toSq = int8(bits.TrailingZeros64(kingMovesBB))
		allMoves = append(allMoves, board.Move{From: kingSq, To: toSq, Piece: KING, PromoteTo: NO_PIECE})
		kingMovesBB &= kingMovesBB - 1
	}

	// If attackerCount > 1 and king has no moves, it is checkmate
	if attackerCount > 1 {
		return allMoves
	}

	pieces := [5]uint64{cb.Pawns[cb.WToMove], cb.Knights[cb.WToMove],
		cb.Bishops[cb.WToMove], cb.Rooks[cb.WToMove], cb.Queens[cb.WToMove],
	}
	moveFuncs := [5]moveGenFunc{GetPawnMoves, getKnightMoves, lookupBishopMoves,
		lookupRookMoves, getQueenMoves,
	}
	symbols := [5]uint8{PAWN, KNIGHT, BISHOP, ROOK, QUEEN}

	// 29% perft() speed up and -40% malloc from having this loop in this function
	var fromSq int8
	for i, pieceBB := range pieces {
		for pieceBB > 0 {
			fromSq = int8(bits.TrailingZeros64(pieceBB))
			pieceBB &= pieceBB - 1

			movesBB := moveFuncs[i](fromSq, cb) & ^cb.Pieces[cb.WToMove]
			for movesBB > 0 {
				toSq = int8(bits.TrailingZeros64(movesBB))
				movesBB &= movesBB - 1

				if capturesBlks == 0 || uint64(1<<toSq)&capturesBlks != 0 {
					if i != 0 || (7 < toSq && toSq < 56) {
						allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: symbols[i], PromoteTo: NO_PIECE})
					} else {
						allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: symbols[i], PromoteTo: KNIGHT})
						allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: symbols[i], PromoteTo: BISHOP})
						allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: symbols[i], PromoteTo: ROOK})
						allMoves = append(allMoves, board.Move{From: fromSq, To: toSq, Piece: symbols[i], PromoteTo: QUEEN})
					}
				}
			}
		}
	}

	return allMoves
}

// Return the set of squares of pieces checking the king and interposition
// squares, and the number of checking pieces.
func GetCheckingSquares(cb *board.Board) (uint64, int) {
	opponent := 1 ^ cb.WToMove
	attackerCount := 0

	kSquare := cb.KingSqs[cb.WToMove]
	pAttackers := cb.PAttacks[cb.WToMove][kSquare] & cb.Pawns[opponent]
	knightAttackers := cb.NAttacks[kSquare] & cb.Knights[opponent]
	bqAttackers := lookupBishopMoves(kSquare, cb) & (cb.Bishops[opponent] |
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

	panicMsgs := [2]string{">1 piece is checking king orthogonally",
		">1 piece is checking king diagonally"}
	attackers := [2]uint64{orthogAttackers, bqAttackers}

	// Add interposition squares if any exist.
	for i, attacker := range attackers {
		if attacker != 0 {
			attackerSquares := read1Bits(attacker)
			attackerCount += len(attackerSquares)
			if len(attackerSquares) > 1 {
				if i == 0 && cb.PrevMove.From == cb.KingSqs[cb.WToMove]-8 &&
					(cb.PrevMove.PromoteTo == ROOK || cb.PrevMove.PromoteTo == QUEEN) {
					// Two pieces can orthogonally check a king if one was just promoted
					// from a pawn, with the other piece previously protecting the pawn
				} else {
					panic(panicMsgs[i])
				}
			}
			dir := findDirection(cb.KingSqs[cb.WToMove], attackerSquares[0])
			attackers[i] = fillFromTo(cb.KingSqs[cb.WToMove], attackerSquares[0], dir)
		}
	}

	return pAttackers | knightAttackers | attackers[0] | attackers[1], attackerCount
}

// Return a bitboard of squares between `from and `to`, excluding `from`
// and including `to`
func fillFromTo(from, to, direction int8) uint64 {
	bb := uint64(0)
	for sq := from + direction; sq != to; sq += direction {
		bb += 1 << sq
	}
	bb += 1 << to

	return bb
}

// Return the direction from one square to another. Assumes (from, to) is an
// orthogonal or diagonal move
func findDirection(from, to int8) int8 {
	var dir int8
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

func read1Bits(bb uint64) []int8 {
	// Using TrailingZeros64() seems as fast as bitshifting right while bb>0.
	squares := make([]int8, 0, 4)
	for bb > 0 {
		squares = append(squares, int8(bits.TrailingZeros64(bb)))
		bb &= bb - 1
	}
	return squares
}

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
