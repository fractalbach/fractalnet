/*
===========================================================================
>   Game of Life and War
___________________________________________________________________________

Grid Square Types

0 = Neutral
1 = Player 1
2 = Player 2
3 = Red
4 = Pink
[5 - 10] = Red, in different

Credits

- Joshua Fan - The Idea for the "Game of War"
- Golang.org - Code Snippets for Implementation of Conway's Game of Life
- John Horton Conway - For inventing the "Game of Life"
___________________________________________________________________________
===========================================================================
*/
package gameofwar

//package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	//"fmt"
	"log"
	"math/rand"
	"time"
)

// ===========================================================================
//      Modifiable Game Instance Settings
// ___________________________________________________________________________

const (
	MIN_GAME_SPEED = 1
	MAX_GAME_SPEED = 2
	maxEnumeration = 20
)

// GameInstance contains all the information that would be found in the game:
// including the map, game state, players, and settings.
type GameInstance struct {
	life              *Life
	gamespeed         time.Duration
	w, h              int
	firstBombIndex    uint8
	firstFalloutIndex uint8
	//observer chan string
}

// NewGameInstance initializes a fresh game, only asking for a map size.
// All of the other settings start at default.
func NewGameInstance(w, h int) *GameInstance {
	return &GameInstance{
		w:                 w,
		h:                 h,
		life:              NewLife(w, h),
		gamespeed:         1 * time.Second,
		firstBombIndex:    10,
		firstFalloutIndex: 100,
	}
}

// ===========================================================================
//      Player Interaction
// ___________________________________________________________________________

// DropBomb takes a player number (1 or 2), and a position on the grid.
// If that action is allowed, then it will do that action
func (g *GameInstance) DropBomb(x, y int) bool {
	g.life.doLaBomba(x, y)
	return true
	/*
		// Error Check: Player ID should only be 1 or 2.
		if (p != 1) && (p != 2) {
			log.Println("DropBomb: That is not a valid team number.")
			return false
		}

		// Checks to see if the player is actually allowed to bomb that spot.
		me := g.life.a.WhatIs(x, y)
		if (p != me) && (p != 0) {
			log.Println("Team", p, "not allowed to drop bomb at:", x, y)
			return false
		}
	*/
	//log.Println("team", p, "Dropped La Bomba! at", x, y)

}

// ===========================================================================
//      Do La Bomba!  💥  💣
// ___________________________________________________________________________

/*
doLaBomba is a private function that changes the value at x,y and all
of the values in the neighborhood (a total of 9 squares updated.)


The Configuration of la Bomba should look like this:

          ooo
         o+-+o
         o-@-o
         o+-+o
          ooo

The Symbols   o, +, -, @    represent different valued squares.

        o = 3
        + = 4
        - = 5
        @ = 5

*/
func (l *Life) doLaBombaFailz(x, y int) {

	// Iterate through the 5x5 grid, deciding a value for each square.
	for i := -2; i <= 2; i++ {
		for j := -2; j <= 2; j++ {

			// Determine new value for the square (see figure above)
			var val uint8
			switch {
			case (i == 2 || i == -2 || j == 2 || j == -2) &&
				(i != j) && (-i != j):
				val = 3

			case (i == 1 || i == -1) && (j == 1 || j == -1):
				val = 5

			case (i == j || -i == j) && (i == 1 || i == -1):
				val = 3

			case i == 0 && j == 0:
				val = 4

			default:
				val = 0 // this would actually be an error.
			}
			l.AlterAt(x+i, y+j, val)
		}
	}
}

func (l *Life) doLaBomba(x, y int) {

	// The 8 Neighbors
	l.AlterAt(x, y, 5)
	l.AlterAt(x+1, y, 5)
	l.AlterAt(x-1, y, 5)
	l.AlterAt(x, y+1, 5)
	l.AlterAt(x, y-1, 5)

	l.AlterAt(x+1, y+1, 4)
	l.AlterAt(x-1, y-1, 4)
	l.AlterAt(x-1, y+1, 4)
	l.AlterAt(x+1, y-1, 4)

	// ---------------------------
	// The Far Neighbors

	//
	/* The Corners
	l.AlterAt(x+2, y-2, 3)
	l.AlterAt(x+2, y+2, 3)
	l.AlterAt(x-2, y-2, 3)
	l.AlterAt(x-2, y+2, 3)
	*/
	l.AlterAt(x+2, y-1, 3)
	l.AlterAt(x+2, y, 3)
	l.AlterAt(x+2, y+1, 3)
	l.AlterAt(x-2, y-1, 3)
	l.AlterAt(x-2, y, 3)
	l.AlterAt(x-2, y+1, 3)
	l.AlterAt(x, y-2, 3)
	l.AlterAt(x, y+2, 3)
	l.AlterAt(x+1, y-2, 3)
	l.AlterAt(x-1, y-2, 3)
	l.AlterAt(x+1, y+2, 3)
	l.AlterAt(x-1, y+2, 3)

}

// ===========================================================================
//      Cellular Automata Grid and Stuff
// ___________________________________________________________________________

// Field represents a two-dimensional field of cells.
type Field struct {
	s    [][]uint8
	w, h int
}

// Neighborhood stores 2 useful maps:
// 1. The cell values based on location,
// 2. The total number of cells found nearby, based on a specific cell value.
type Neighborhood struct {
	direct map[uint8]uint8
	totals map[uint8]uint8
}

// North returns the value at coordinate (+0, -1) relative to center.
func (n *Neighborhood) North() uint8 {
	return n.direct[0]
}

// South returns the value at coordinate (+0, +1) relative to center.
func (n *Neighborhood) South() uint8 {
	return n.direct[1]
}

// East returns the value at coordinate (+1, +0) relative to center.
func (n *Neighborhood) East() uint8 {
	return n.direct[2]
}

// West returns the value at coordinate (-1, +0) relative to center.
func (n *Neighborhood) West() uint8 {
	return n.direct[3]
}

// NorthEast returns the value at coordinate (+1, -1) relative to center.
func (n *Neighborhood) NorthEast() uint8 {
	return n.direct[4]
}

// NorthWest returns the value at coordinate (-1, -1) relative to center.
func (n *Neighborhood) NorthWest() uint8 {
	return n.direct[5]
}

// SouthEast returns the value at coordinate (+1, +1) relative to center.
func (n *Neighborhood) SouthEast() uint8 {
	return n.direct[6]
}

// SouthWest returns the value at coordinate (+1, -1) relative to center.
func (n *Neighborhood) SouthWest() uint8 {
	return n.direct[7]
}

// Count returns the total number of cells in the neighborhood that contain
// the specified value.
func (n *Neighborhood) Count(cellValue uint8) uint8 {
	return n.totals[cellValue]
}

// NewMooreNeighborhood returns a neighborhood of cells that include the 4
// cardinal directions and the 4 diagonals, for a total of 8 nearby cells.
func NewMooreNeighborhood(f *Field, x, y int) *Neighborhood {
	n := Neighborhood{
		direct: map[uint8]uint8{
			0: f.WhatIs(x, y-1),   // North
			1: f.WhatIs(x, y+1),   // South
			2: f.WhatIs(x+1, y),   //East
			3: f.WhatIs(x-1, y),   // West
			4: f.WhatIs(x+1, y-1), // North-East
			5: f.WhatIs(x-1, y-1), // North-West
			6: f.WhatIs(x+1, y+1), // South-East
			7: f.WhatIs(x-1, y+1), // South-West
		},
		totals: map[uint8]uint8{
			0: 0, // Empty Neutral
			1: 0, // player 1
			2: 0, // player 2
			3: 0, // A
			4: 0, // B
			5: 0, // C
			6: 0, // Static? (may be not used)
			7: 0,
		},
	}
	for _, cellValue := range n.direct {
		if _, ok := n.totals[cellValue]; ok {
			n.totals[cellValue]++
		}
	}
	return &n
}

// WhoEatsMe returns the color that consumes the specified color, if nearby.
// If will only return a 0 if it passes through all of the cases.
/*
func WhoEatsMe(myColor uint8) uint8 {
	switch myColor {
	case 0:
		return 3 // empty  <- red
	case 1:
		return 2 // 1 <- 2
	case 2:
		return 1 // 2 <- 1
	case 3:
		return 4
	case 4:
		return 5
	case 5:
		return 6
	case 6:
		return 3
	}
	return 0 // empty eaten by empty
	// uint8(rand.Intn(5))
}
*/
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

// Life stores the state of a round of Conway's Game of Life.
type Life struct {
	a, b *Field
	w, h int
}

// NewLife returns a new Life game state with a random initial state.
func NewLife(w, h int) *Life {
	a := NewField(w, h)

	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			a.Set(i, j, uint8(rand.Intn(2)+1))
		}
	}
	/*
		for i := 0; i < (w * h / 4); i++ {
			a.Set(rand.Intn(w), rand.Intn(h), uint8(rand.Intn(2)+1))
		}
	*/
	return &Life{
		a: a, b: NewField(w, h),
		w: w, h: h,
	}
}

// ResetGameInstance will randomly create player squares.
// The amount of player squares to make is specified by the "numberToMake".
func (g *GameInstance) RandomizeGameBoard(numberToMake int) {
	if numberToMake <= 0 {
		numberToMake = g.w * g.h / 4
	}
	a := NewField(g.w, g.h)
	for i := 0; i < numberToMake; i++ {
		a.Set(rand.Intn(g.w), rand.Intn(g.h), uint8(rand.Intn(2)+1))
	}
	g.life.a = a
}

func (g *GameInstance) FreshGameBoard() {
	w, h := g.w, g.h
	a := NewField(w, h)
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			a.Set(i, j, uint8(rand.Intn(2)+1))
		}
	}
	g.life.a = a
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
			// █ ▒
			// ░▒▓▓░▒▓█
			//b = '💙'
			//b = '💚'
			//b = '💗'
			switch l.a.WhatIs(x, y) {
			case 0:
				b = '0' // empty
			case 1:
				b = '.' // player 1
			case 2:
				b = ',' // player 2
			case 3:
				b = '▓' //  ▓🔥
			case 4:
				b = '▒' // ▒🌊▒
			case 5:
				b = '░' //░
			case 6:
				b = '#'
			case 7:
				b = '&'
			default:
				b = '?'
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
func (g *GameInstance) LifeStateMessage() []byte {
	return g.life.a.encodeFieldData()
}

func (g *GameInstance) ChangeAt(x, y int, val uint8) {
	g.life.AlterAt(x, y, val)
}

func (g *GameInstance) LifeUpdate() {
	g.life.Step()
}

// AlterAt changes the value at a specific position of the field.
func (l *Life) AlterAt(x, y int, val uint8) {
	if (x < 0) || (y < 0) || (x >= l.w) || (y >= l.h) {
		return
	}
	l.a.Set(x, y, val)
}

//  _________________________________________________________________________
// /~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~\
//      *~-.,_,.-~-.,_ Unique Functions for Mechanics  _,.-~-.,_,.-~*
// |_________________________________________________________________________|
// |~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~|

// Next returns the state of the specified cell at the next time step.
func (f *Field) Next(x, y int) uint8 {
	/*
		cellsToCheck := []uint8{ // Count values in adjacent cells.
			f.WhatIs(x, y+1),
			f.WhatIs(x, y-1),

			f.WhatIs(x-1, y-1),
			f.WhatIs(x-1, y),
			f.WhatIs(x-1, y+1),

			f.WhatIs(x+1, y-1),
			f.WhatIs(x+1, y+1),
			f.WhatIs(x+1, y),

			// Von Neumann group
			f.WhatIs(x, y+2),
			f.WhatIs(x, y-2),
			f.WhatIs(x+2, y),
			f.WhatIs(x-2, y),
		}
		// Make a "neighborhood" HashMap to hold the (Key, Value) Pairs.
		// Count the cells values, and add them to the counters.
		n := NewNeighborhood()
		for _, cellValue := range cellsToCheck {
			if _, ok := n[cellValue]; ok {
				n[cellValue]++
			}
		}
	*/
	n := NewMooreNeighborhood(f, x, y)
	me := f.WhatIs(x, y) // What value is at this cell?``
	//enemy := WhoEatsMe(me) // What cell eats this value?

	if me >= 7 {
		return 0
	}

	if me >= 3 {
		return me + 1
	}

	switch me {

	// ~~~~ Special Fire Rule ~~~~

	/*	case me <= 2 && n[enemy] == 1 && rand.Intn(16) == 0:
		return enemy //probability of being consumed by flame
	*/
	// ~~~~ Game of Life rules ~~~~
	// Slightly modified from Conway's game; to deal with 2 players.

	/*	case me <= 2 && (n[1]+n[2]) >= 4: // overcrowded
		return 0
	*/

	case 0:
		points1 := 0
		points2 := 0

		// Specific Squares for Player 2 -------------------
		if n.North() == 2 {
			points2++
		}
		if n.NorthWest() == 2 {
			points2++
		}
		if n.NorthEast() == 2 {
			points2++
		}
		if n.West() == 2 {
			points2++
		}
		if n.East() == 2 {
			points2++
		}

		// Specific squares for Player 1 ---------------

		if n.South() == 1 {
			points1++
		}
		if n.SouthEast() == 1 {
			points1++
		}
		if n.SouthWest() == 1 {
			points1++
		}
		if n.West() == 1 {
			points1++
		}
		if n.East() == 1 {
			points1++
		}

		//Test...
		if n.South() == 2 {
			points2++
		}
		if n.North() == 1 {
			points1++
		}
		// ... End Test

		switch {
		case points1 == points2:
			return 7
		case points1 > points2:
			return 1
		case points1 < points2:
			return 2
		}

	} //end of switch
	return me
}

// \_________________________________________________________________________/

// Rules if you are an empty square. Contested squares are not filled.
/*  if me == 0 {
    greatColors, greatValue := Abundance(n)
    if ((greatValue == 3) || (greatValue == 5)) && len(greatColors) == 1 {
        return greatColors[0]
    } else {
        return 0
    }
}*/

/*
func main() {
    iters := 50
    g := NewGameInstance(48, 48)
    t := 5

    for i := 0; i < iters; i++ {
        g.life.Step()

        fmt.Print("\x0c") // Clear screen and print field.
        fmt.Println("Iteration =", i)
        if t == i {
            g.life.doLaBomba(10, 10)
			g.life.doLaBomba(30,10)
			g.life.doLaBomba(30,20)
			g.life.doLaBomba(20,40)
        }
        fmt.Print(g.life)
        time.Sleep(time.Second / 30)
    }
}
*/
/*
x := l.a.encodeFieldData()
fmt.Println(len(x))
fmt.Println(string(x)[:100])*/
