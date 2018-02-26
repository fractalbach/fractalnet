// TreeGen
//
//  Fun Experiments from the playground:
/*
Compressing Boolean Matrices to Base 64:
	https://play.golang.org/p/mX9XQY-5IdN

*/
package game

import (
	"crypto/rand"
	"fmt"
)

// tree is basically just an x and y location,
// later, a tree will represent an entity in the game.
type tree struct {
	x int
	y int
}

// TreeList is designed to be easily converted into a list of objects.
//
// It's easier to convert TreeList into an array of in-game objects,
// than it would be to convert the 2-dimensional boolean array
type TreeList struct {
	w     int
	h     int
	Trees []tree
}

type BoolGrid struct {
	w, h int
	grid [][]bool
}

func (b *BoolGrid) Set(x, y int, v bool) {
	b.grid[x][y] = v
}

// CreateInitialTrees outputs a Randomized Boolean Grid pointer.
//
// The boolean grid is actually a matrix, of size w * h.
//
func CreateRandomInitialTrees(w, h int) *BoolGrid {
	bg := ConvertToBoolGridType(MakeRandomBoolGrid(w, h))
	bg.NextGeneration()
	return bg
}

func GenerateRandomTree(w, h int) tree {
	return tree{
		x: gimmeRandom(w),
		y: gimmeRandom(h),
	}
}

func MakeTreeList(w, h, TreesToMake int) *TreeList {
	var treeMap []tree
	for i := 0; i < TreesToMake; i++ {
		treeMap = append(treeMap, GenerateRandomTree(w, h))
	}
	return &TreeList{
		w:     w,
		h:     h,
		Trees: treeMap,
	}
}

// MakeBoolGrid creates a 2-d array, where each array is a horizontal!
// Keep in mind that each grid[x] would appear as a column.
//
func (tl *TreeList) MakeBoolGrid() *BoolGrid {
	grid := make([][]bool, tl.w)
	for i := range grid {
		grid[i] = make([]bool, tl.h)
	}
	for _, v := range tl.Trees {
		grid[v.x][v.y] = true
	}
	return &BoolGrid{
		w:    tl.w,
		h:    tl.h,
		grid: grid,
	}
}

/*
The Rules of the Game of Life

- Any live cell with 2 or 3 live neighbors lives
- Any live cell with fewer than 2 live neighbors dies
- Any live cell with more than 3 live neighbors dies
- Any dead cell with 3 live neighbors becomes a live cell


+-------------+---------------------+
|  Number of  | Current Cell Status |
|   Living    |----------+----------|
|  Neighbors  |   Alive  |   Dead   |
|===================================|
|      0      |          |          |
|-------------+----------+----------|
|      1      |          |          |
|-------------+----------+----------|
|      2      |  Lives!  |          |
|-------------+----------+----------|
|      3      |  Lives!  |   Lives! |
|-------------+----------+----------|
|      4      |          |          |
+-------------+----------+----------+

*/
func (b *BoolGrid) NextGeneration() {
	future := BoolGrid{
		w:    b.w,
		h:    b.h,
		grid: MakeEmptyBoolGrid(b.w, b.h),
	}
	for i := 0; i < b.w; i++ {
		for j := 0; j < b.h; j++ {
			c := b.CountLivingNeighbors(i, j)
			if (c == 3) || ((c == 2) && (b.grid[i][j])) {
				future.grid[i][j] = true
			} else {
				future.grid[i][j] = false
			}
			//fmt.Println("(x,y):",i,j,"(neighbors):", c,"(past):", b.grid[i][j],"(future):", future.grid[i][j])
		}
	}
	b.grid = future.grid
}

// CountLivingNeighbors returns an integer in [0, 4].
// represents the number of other living trees that surround square (x, y),
// including those that wrap around the grid.
func (b *BoolGrid) CountLivingNeighbors(x, y int) int {
	lives := 0
	neighbors := []bool{
		b.Alive(x, y+1),
		b.Alive(x, y-1),
		b.Alive(x+1, y),
		b.Alive(x-1, y),
		b.Alive(x+1, y+1),
		b.Alive(x+1, y-1),
		b.Alive(x-1, y+1),
		b.Alive(x-1, y-1),
	}
	for _, v := range neighbors {
		if v {
			lives++
		}
	}
	return lives
}

// Alive reports whether the specified cell is alive, it simply calls another
// function: one of the other "Alive" functions with specific rules.
func (b *BoolGrid) Alive(x, y int) bool {
	return b.AliveNoWrap(x, y)
}

// AliveNoWrap reports whether the specified cell is alive. It does NOT wrap.
// If the value is outside of the boundaries, it will return false.
func (b *BoolGrid) AliveNoWrap(x, y int) bool {
	if (x < 0) || (y < 0) || (x >= b.w) || (y >= b.h) {
		return false
	}
	return b.grid[x][y]
}

// AliveWrap reports whether the specified cell is alive.  It treats the world
// as if it wraps around.
//
// If the x or y coordinates are outside the field boundaries they are wrapped
// toroidally. For instance, an x value of -1 is treated as width-1.
//
// CREDITS: golang.org homepage example
func (b *BoolGrid) AliveWrap(x, y int) bool {
	x += b.w
	x %= b.w
	y += b.h
	y %= b.h
	return b.grid[x][y]
}

// gimmeRandom returns a random int \in [0, max], max <= 1024
func gimmeRandom(max int) int {
	a := make([]byte, 3)
	rand.Read(a)
	return int(a[0]+a[1]+a[2]) % max
}

func (tl *TreeList) prettyPrint() {
	tl.MakeBoolGrid().prettyPrint()
}

func (b *BoolGrid) prettyPrint() {
	for i := 0; i < b.w; i++ {
		for j := 0; j < b.h; j++ {
			if b.grid[i][j] {
				fmt.Print("█")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Print("\n")
	}
}

func (b *BoolGrid) test() {
	for i := 0; i < b.w; i++ {
		for j := 0; j < b.h; j++ {
			fmt.Print(b.CountLivingNeighbors(i, j))
		}
		fmt.Print("\n")
	}
}

func makeMyExample() *TreeList {
	return &TreeList{
		w: 6,
		h: 80,
		Trees: []tree{
			tree{0, 34}, tree{0, 35},
			tree{1, 33}, tree{1, 36},
			tree{2, 34}, tree{2, 35},
			tree{3, 33}, tree{3, 36},
			tree{4, 34}, tree{4, 35},
		},
	}
}

// Just a hack, need to get rid of the boolgrid type at some point.
func ConvertToBoolGridType(bg [][]bool) *BoolGrid {
	return &BoolGrid{
		w:    len(bg),
		h:    len(bg[0]),
		grid: bg,
	}
}

/*
func main() {
	iters := 200

	//w := 20
	//h := 100
	//TreesToMake := w * h
	//tl := MakeTreeList(w, h, TreesToMake)

	tl := makeMyExample()
	bg := tl.MakeBoolGrid()
	//fmt.Println(bg)

	for i := 0; i < iters; i++ {
		fmt.Print("\x0c")
		bg.prettyPrint()
		//bg.test()
		bg = bg.NextGeneration()
		time.Sleep(time.Second / 8)
	}
}
*/
