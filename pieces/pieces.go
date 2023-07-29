package pieces

import (
	_ "fmt"
	_ "math"
)

// bb = bit board

func makeFiles() [4][8]int {
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
func makeFiles() [4]map[int]bool {
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
	files := makeFiles()

	//fmt.Println(files)

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
			//fmt.Println(math.Pow(2, float64(sq + d)), d)
			bbs[sq] += 1 << (sq + d)
		}
	}

	return bbs
}

func makeKingBBs() [64]uint64 {
	bbs := [64]uint64{}
	directions := []int{}
	files := makeFiles()

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
	files := makeFiles()
	// TODO: make containsN() generic to remove wasted zeroes.
	// Or use slices instead of arrays.
	fileAForbidden := [8]int{-9, -1, 7, 0, 0, 0, 0, 0}
	fileHForbidden := [8]int{9, 1, -7, 0, 0, 0, 0, 0}

	// Clockwise ordering
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
