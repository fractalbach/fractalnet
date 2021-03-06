/*
_________________________________________________________________________
|                                                                       |
|                                                                       |
|                                Locations                              |
|                 ______________________________________                |
|                                                                       |
|                             Chris Achenbach                           |
|                               28 Feb 2018                             |
|_______________________________________________________________________|
*/

package game

/*
_________________________________________________________________________
                        2 Dimensional Locations
=========================================================================
*/

type Location struct {
	X, Y int
}

// MoveTo directly and instantly changes the location to given x and y.
func (l *Location) Set(x, y int) {
	l.X, l.Y = x, y
	return
}

/*
_________________________________________________________________________
                        3 Dimensional Locations
=========================================================================
*/
