package wave

import (
	"bytes"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"log"
	"math/rand"
	//"fmt"
	//"time"
)

var (
	MAX_VAL = 3
)

// Field represents a two-dimensional field of cells.
type Field struct {
	s    [][]uint8
	w, h int
}

// Neighborhood represents the total color counts in the surrounding cells.
func NewNeighborhood() map[uint8]uint8 {
	return map[uint8]uint8{
		0: 0, // Empty
		1: 0, // Red
		2: 0, // Green
		3: 0, // Blue
	}
}

// WhoEatsMe returns the color that consumes the specified color, if nearby.
// If will only return a 0 if it passes through all of the cases.
func WhoEatsMe(myColor uint8) uint8 {
	switch myColor {
	case 1:
		return 3
	case 2:
		return 1
	case 3:
		return 2
	}
	return 0
}

// Abundance the color(s), which corresponds to a value [1-3],
// and the number of counted neighbors of that color [0-8].
// Since there could be a tie for the highest count,
// multiple colors are returned in an array.
func Abundance(neighborhood map[uint8]uint8) ([]uint8, uint8) {
	var greatColor []uint8
	var greatValue uint8
	for k, v := range neighborhood {
		if v == greatValue {
			greatColor = append(greatColor, k)
			continue
		}
		if v > greatValue {
			greatColor = []uint8{k}
			greatValue = v
		}
	}
	return greatColor, greatValue
}

// NewField returns an empty field of the specified width and height.
func NewField(w, h int) *Field {
	s := make([][]uint8, h)
	for i := range s {
		s[i] = make([]uint8, w)
	}
	return &Field{s: s, w: w, h: h}
}

// Set sets the state of the specified cell to the given value.
func (f *Field) Set(x, y int, b uint8) {
	f.s[y][x] = b
}

// WhatsIs reports the number that is at the specified cell.
// NOTE: THIS WRAPS AROUND THE MAP.
func (f *Field) WhatIs(x, y int) uint8 {
	//x += f.w
	//x %= f.w
	//y += f.h
	//y %= f.h
	if (x < 0) || (y < 0) || (x >= f.w) || (y >= f.h) {
		return 0
	}
	return f.s[y][x]
}

// Next returns the state of the specified cell at the next time step.
func (f *Field) Next(x, y int) uint8 {
	// Count the adjacent cells that are alive.
	n := NewNeighborhood()
	for _, i := range []int{-1, 0, 1} {
		for _, j := range []int{-1, 0, 1} {
			if j != 0 || i != 0 {
				val := f.WhatIs(x+i, y+j)
				if val != 0 {
					n[val]++
				}
			}
		}
	}
	q := []uint8{
		f.WhatIs(x-2, y),
		f.WhatIs(x, y-2),
		f.WhatIs(x+2, y),
		f.WhatIs(x, y+2),
	}
	for _, val := range q {
		if val != 0 {
			n[val]++
		}
	}

	// Return next state according to the game rules:
	me := f.WhatIs(x, y)
	enemy := WhoEatsMe(me)

	if n[enemy] >= 3 {
		return enemy
	}

	if (n[me] == 2) || (n[me] == 3) || (n[me] == 5) {
		return me
	}
	greatColors, greatValue := Abundance(n)
	if greatValue >= 0 && len(greatColors) == 1 {
		return greatColors[0]
		//return greatColors[gimmeRandom(len(greatColors))]
	}

	// greatColors, greatValue := Abundance(n)

	return me
}

// Life stores the state of a round of Conway's Game of Life.
type Life struct {
	a, b *Field
	w, h int
}

// NewLife returns a new Life game state with a random initial state.
func NewLife(w, h int) *Life {
	a := NewField(w, h)
	for i := 0; i < (w * h / 1); i++ {
		a.Set(rand.Intn(w), rand.Intn(h), uint8(gimmeRandom(4)))
	}
	return &Life{
		a: a, b: NewField(w, h),
		w: w, h: h,
	}
}

// Step advances the game by one instant, recomputing and updating all cells.
func (l *Life) Step() {
	// Update the state of the next field (b) from the current field (a).
	for y := 0; y < l.h; y++ {
		for x := 0; x < l.w; x++ {
			l.b.Set(x, y, l.a.Next(x, y))
		}
	}
	// Swap fields a and b.
	l.a, l.b = l.b, l.a
}

// String returns the game board as a string.
func (l *Life) String() string {
	var buf bytes.Buffer
	for y := 0; y < l.h; y++ {
		for x := 0; x < l.w; x++ {
			b := ' '
			// █
			// ░
			// ▒
			// ▓
			//b = '💙'
			//b = '💚'
			//b = '💗🐱'
			switch l.a.WhatIs(x, y) {
			case 0:
				b = '░'
			case 1:
				b = '▒'
			case 2:
				b = '▓'
			case 3:
				b = '█'
			}

			buf.WriteRune(b)
		}
		buf.WriteByte('\n')
	}
	return buf.String()
}

type GridState struct {
	GridState string
}

// encodeFieldData converts the field of cells into an array, then encodes
// them in base64.  Then, it is encapsulated in a json message called
// "GridState".  The JSON is returned in the form of a byte array.
func (f *Field) encodeFieldData() []byte {
	arr := []byte{}
	for _, v := range f.s {
		arr = append(arr, v...)
	}
	b64 := base64.StdEncoding.EncodeToString(arr)
	msg, err := json.Marshal(GridState{b64})
	if err != nil {
		log.Println(err)
		return []byte{}
	}
	return msg
}

// LifeStateMessage returns an encoded Json message, ready to be sent.
func (l *Life) LifeStateMessage() []byte {
	return l.a.encodeFieldData()
}

// AlterAt changes the value at a specific position of the field.
func (l *Life) AlterAt(x, y int, val uint8) {
	if (x < 0) || (y < 0) || (x >= l.w) || (y >= l.h) || (val > 3) {
		return
	}
	l.a.Set(x, y, val)
}

func gimmeRandom(max int) int {
	a := make([]byte, 1)
	crand.Read(a)
	return int(a[0]) % max
}

/*
func main() {
	iters := 100

	l := NewLife(60, 30)
	for i := 0; i < iters; i++ {
		l.Step()
		fmt.Print("\x0c") // Clear screen and print field.
		fmt.Print(l)
		time.Sleep(time.Second / 10)
	}
}
*/
//░
// ▒
// ▓
