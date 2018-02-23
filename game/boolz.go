package game

import (
	"crypto/rand"
	"errors"
)

/*
Example For using Bool Grids Compression
    https://play.golang.org/p/WmS2RkgI8p6
    https://play.golang.org/p/_2nSGJHF2Sy
*/

func MakeRandomByteGrid(w, h int) [][]byte {
	x := make([]byte, h*w)
	rand.Read(x)
	grid := make([][]byte, w)
	for i := range grid {
		grid[i] = x[h*i : h*(i+1)]
	}
	return grid
}

func MakeRandomBoolGrid(w, h int) [][]bool {
	g := MakeRandomByteGrid(w, h)
	out := make([][]bool, w)
	for i := range g {
		out[i] = make([]bool, h)
		for j := range g[i] {
			out[i][j] = g[i][j] < 128
		}
	}
	return out
}

func MakeEmptyBoolGrid(w, h int) [][]bool {
	grid := make([][]bool, w)
	for i := 0; i < w; i++ {
		grid[i] = make([]bool, h)
	}
	return grid
}

// BoolArrayToByteArray returns byte array and number of bits in last byte.
//
// Iterating through the Boolean Array, each boolean is treated as a bit.
// If there are leftover bits at the end of the array, they will all be
// converted into a single byte.
//
// Note!  The Resulting Byte Array itself does not contain the information
// about how many bits were used to make the last byte!  When unraveling
// back into a boolean array, it is assumed that you already know this info.
//
func BoolArrayToByteArray(bo []bool) ([]byte, int) {
	var by []byte
	iters := len(bo) / 8
	lbylen := len(bo) % 8
	for i := 0; i < iters; i++ {
		var thisbyte byte
		slicey := bo[8*i : 8*(i+1)]
		for j, v := range slicey {
			if v {
				thisbyte += 1 << uint(j)
			}
			//fmt.Println(j, thisbyte, 1<<uint(j))
		}
		by = append(by, thisbyte)
	}
	if lbylen == 0 {
		return by, lbylen
	}

	slicey := bo[len(bo)-lbylen : len(bo)]
	//fmt.Println("slice indexes:", len(bo)-lbylen, len(bo), slicey)
	var lastbyte byte
	for i, v := range slicey {
		if v {
			lastbyte += 1 << uint(i)
		}
		//fmt.Println(i, v)
	}
	by = append(by, lastbyte)
	return by, lbylen
}

// ByteArrayToBoolArray converts each byte into 8 booleans.
//
// The number of bits in the last byte is used in the event that you
// want a boolean array that is not divisible by 8.  If you want the last byte
// to be treated the same as the others, set bitsInLastByte = 0.
//
// ByteArrayToBoolArray is useful for unraveling the compressed byte array
// from the function BoolArrayToByteArray.
//
func ByteArrayToBoolArray(bo []byte, bitsInLastByte int) []bool {
	var theBools []bool
	indexLastByte := len(bo) - 1
	for i, v := range bo {
		var thisSlice []bool
		val := int(v)
		start := 7
		if (i == indexLastByte) && (bitsInLastByte != 0) {
			start = bitsInLastByte - 1
		}
		for j := start; j >= 0; j-- {
			test := (1 << uint(j))
			//fmt.Println(j, val, test)
			if val >= test {
				thisSlice = append([]bool{true}, thisSlice...)
				val -= test
			} else {
				thisSlice = append([]bool{false}, thisSlice...)
			}
		}
		theBools = append(theBools, thisSlice...)
	}
	return theBools
}

// CompressBoolGrid takes a boolean matrix and converts it into bytes.
// Returns the array of bytes and the number of leftover bits in last byte.
// The leftover bits should be equal to (width * height) % 8.
//
// This byte stream is useful for sending messages over the web, because it
// is in a form that will take up less space.
//
// Must use be a m x n matrix, where each boolean array is the same length.
// Otherwise, CompressBoolGrid will return an error message.
//
// An empty grid will simply return a empty byte array, with no error.
//
func CompressBoolGrid(bg [][]bool) ([]byte, int, error) {
	w := len(bg)
	h := len(bg[0])
	if (w == 0) || (h == 0) {
		return []byte{}, 0, nil
	}

	giantArray := []bool{}
	for _, v := range bg {
		if len(v) != h {
			return []byte{}, 0, errors.New("Variable Lengths in Boolean Grid.")
		}
		giantArray = append(giantArray, v...)
	}

	by, n := BoolArrayToByteArray(giantArray)
	if n != (w * h % 8) {
		return []byte{}, 0, errors.New("Incorrect number of leftover bits.")
	}
	return by, (w * h % 8), nil
}

// ConvertBoolGridToLocationList takes a w * h boolean matrix, and returns the
// a list of x, y positions for the "true" values in the boolean matrix.
//
// Note: this is designed for grids with less than 128 dots.  Error will throw
// if you attempt to use anything above.
//
func ConvertBoolGridToLocationList(b [][]bool) ([]uint8, error) {
	if len(b) >= 256 {
		return nil, errors.New(
			"Width and Height must each be less than 256.")
	}
	if len(b) <= 0 {
		return nil, errors.New(
			"Width and Height must be positive and nonzero.")
	}
	var out []uint8
	for i, x := range b {
		if len(x) >= 256 {
			return nil, errors.New(
				"Width and Height must each be less than 256.")
		}
		if len(x) <= 0 {
			return nil, errors.New(
				"Width and Height must be positive and nonzero.")
		}
		for j, v := range x {
			if v {
				out = append(out, uint8(i), uint8(j))
			}
		}
	}
	return out, nil
}
