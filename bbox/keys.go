package bbox

import (
	"github.com/nsf/termbox-go"
)

type Coord [2]int

type Key struct {
	Ch  rune        // a unicode character
	Key termbox.Key // one of Key* constants, invalid if 'Ch' is not 0
}

var keymaps = map[Key]*Coord{
	{'1', 0}: &Coord{0, 0},
	{'2', 0}: &Coord{0, 1},
	{'3', 0}: &Coord{0, 2},
	{'4', 0}: &Coord{0, 3},
	{'5', 0}: &Coord{0, 4},
	{'6', 0}: &Coord{0, 5},
	{'7', 0}: &Coord{0, 6},
	{'8', 0}: &Coord{0, 7},
	{'!', 0}: &Coord{0, 8},
	{'@', 0}: &Coord{0, 9},
	{'#', 0}: &Coord{0, 10},
	{'$', 0}: &Coord{0, 11},
	{'%', 0}: &Coord{0, 12},
	{'^', 0}: &Coord{0, 13},
	{'&', 0}: &Coord{0, 14},
	{'*', 0}: &Coord{0, 15},

	{'q', 0}: &Coord{1, 0},
	{'w', 0}: &Coord{1, 1},
	{'e', 0}: &Coord{1, 2},
	{'r', 0}: &Coord{1, 3},
	{'t', 0}: &Coord{1, 4},
	{'y', 0}: &Coord{1, 5},
	{'u', 0}: &Coord{1, 6},
	{'i', 0}: &Coord{1, 7},
	{'Q', 0}: &Coord{1, 8},
	{'W', 0}: &Coord{1, 9},
	{'E', 0}: &Coord{1, 10},
	{'R', 0}: &Coord{1, 11},
	{'T', 0}: &Coord{1, 12},
	{'Y', 0}: &Coord{1, 13},
	{'U', 0}: &Coord{1, 14},
	{'I', 0}: &Coord{1, 15},

	{'a', 0}: &Coord{2, 0},
	{'s', 0}: &Coord{2, 1},
	{'d', 0}: &Coord{2, 2},
	{'f', 0}: &Coord{2, 3},
	{'g', 0}: &Coord{2, 4},
	{'h', 0}: &Coord{2, 5},
	{'j', 0}: &Coord{2, 6},
	{'k', 0}: &Coord{2, 7},
	{'A', 0}: &Coord{2, 8},
	{'S', 0}: &Coord{2, 9},
	{'D', 0}: &Coord{2, 10},
	{'F', 0}: &Coord{2, 11},
	{'G', 0}: &Coord{2, 12},
	{'H', 0}: &Coord{2, 13},
	{'J', 0}: &Coord{2, 14},
	{'K', 0}: &Coord{2, 15},

	{'z', 0}: &Coord{3, 0},
	{'x', 0}: &Coord{3, 1},
	{'c', 0}: &Coord{3, 2},
	{'v', 0}: &Coord{3, 3},
	{'b', 0}: &Coord{3, 4},
	{'n', 0}: &Coord{3, 5},
	{'m', 0}: &Coord{3, 6},
	{',', 0}: &Coord{3, 7},
	{'Z', 0}: &Coord{3, 8},
	{'X', 0}: &Coord{3, 9},
	{'C', 0}: &Coord{3, 10},
	{'V', 0}: &Coord{3, 11},
	{'B', 0}: &Coord{3, 12},
	{'N', 0}: &Coord{3, 13},
	{'M', 0}: &Coord{3, 14},
	{'<', 0}: &Coord{3, 15},
}

// mapping from keyboard box
var keymaps_rpi = map[Key]*Coord{
	// 2 x 21 = [volume down]
	// 2 x 24 = [mute]
	// 3 x 19 = ` (quit)

	{'1', 0}:            &Coord{0, 0}, // 3 x 20
	{'q', 0}:            &Coord{0, 1}, // 3 x 21
	{0, termbox.KeyTab}: &Coord{0, 2}, // 3 x 22
	{'a', 0}:            &Coord{0, 3}, // 3 x 23
	{'z', 0}:            &Coord{0, 4}, // 3 x 25
	{0, termbox.KeyF1}:  &Coord{0, 5}, // 4 x 19
	{'2', 0}:            &Coord{0, 6}, // 4 x 20
	{'w', 0}:            &Coord{0, 7}, // 4 x 21
	//{0, CAPS_LOCK}:    &Coord{0, 7}, // 4 x 22
	{'s', 0}: &Coord{0, 8}, // 4 x 23
	// 4 x 24 = §
	{'x', 0}:           &Coord{0, 9},  // 4 x 25
	{0, termbox.KeyF2}: &Coord{0, 10}, // 5 x 19
	{'3', 0}:           &Coord{0, 11}, // 5 x 20
	{'e', 0}:           &Coord{0, 12}, // 5 x 21
	{0, termbox.KeyF3}: &Coord{0, 13}, // 5 x 22
	{'d', 0}:           &Coord{0, 14}, // 5 x 23
	{0, termbox.KeyF4}: &Coord{0, 15}, // 5 x 24

	{'5', 0}: &Coord{1, 0},  // 6 x 19
	{'4', 0}: &Coord{1, 1},  // 6 x 20
	{'r', 0}: &Coord{1, 2},  // 6 x 21
	{'t', 0}: &Coord{1, 3},  // 6 x 22
	{'f', 0}: &Coord{1, 4},  // 6 x 23
	{'g', 0}: &Coord{1, 5},  // 6 x 24
	{'v', 0}: &Coord{1, 6},  // 6 x 25
	{'b', 0}: &Coord{1, 7},  // 6 x 26
	{'6', 0}: &Coord{1, 8},  // 7 x 19
	{'7', 0}: &Coord{1, 9},  // 7 x 20
	{'u', 0}: &Coord{1, 10}, // 7 x 21
	{'y', 0}: &Coord{1, 11}, // 7 x 22
	{'j', 0}: &Coord{1, 12}, // 7 x 23
	{'h', 0}: &Coord{1, 13}, // 7 x 24
	{'m', 0}: &Coord{1, 14}, // 7 x 25
	{'n', 0}: &Coord{1, 15}, // 7 x 26

	{'=', 0}:           &Coord{2, 0},  // 8 x 19
	{'8', 0}:           &Coord{2, 1},  // 8 x 20
	{'i', 0}:           &Coord{2, 2},  // 8 x 21
	{']', 0}:           &Coord{2, 3},  // 8 x 22
	{'k', 0}:           &Coord{2, 4},  // 8 x 23
	{0, termbox.KeyF6}: &Coord{2, 5},  // 8 x 24
	{',', 0}:           &Coord{2, 6},  // 8 x 25
	{0, termbox.KeyF8}: &Coord{2, 7},  // 9 x 19
	{'9', 0}:           &Coord{2, 8},  // 9 x 20
	{'o', 0}:           &Coord{2, 9},  // 9 x 21
	{0, termbox.KeyF7}: &Coord{2, 10}, // 9 x 22
	{'l', 0}:           &Coord{2, 11}, // 9 x 23
	{'.', 0}:           &Coord{2, 12}, // 9 x 25
	{'-', 0}:           &Coord{2, 13}, // 10 x 19
	{'0', 0}:           &Coord{2, 14}, // 10 x 20
	{'p', 0}:           &Coord{2, 15}, // 10 x 21

	{'[', 0}:                   &Coord{3, 0}, // 10 x 22
	{';', 0}:                   &Coord{3, 1}, // 10 x 23
	{'\'', 0}:                  &Coord{3, 2}, // 10 x 24
	{'\\', 0}:                  &Coord{3, 3}, // 10 x 25
	{'/', 0}:                   &Coord{3, 4}, // 10 x 26
	{0, termbox.KeyF9}:         &Coord{3, 5}, // 11 x 19
	{0, termbox.KeyF10}:        &Coord{3, 6}, // 11 x 20
	{0, termbox.KeyBackspace2}: &Coord{3, 7}, // 11 x 22 // 'delete' on mac keyboard
	// 11 x 23 = \ ***
	{0, termbox.KeyF5}:    &Coord{3, 8},  // 11 x 24
	{0, termbox.KeyEnter}: &Coord{3, 9},  // 11 x 25
	{0, termbox.KeySpace}: &Coord{3, 10}, // 11 x 26
	{0, termbox.KeyF12}:   &Coord{3, 11}, // 12 x 20
	// 12 x 21 = 8 ***
	// 12 x 22 = 5 ***
	// 12 x 23 = 2 ***
	// 12 x 24 = 0 ***
	// 12 x 25 = / ***
	{0, termbox.KeyArrowRight}: &Coord{3, 12}, // 12 x 26
	{0, termbox.KeyDelete}:     &Coord{3, 13}, // 13 x 19
	// 13 x 20 = [fn f11]
	// 13 x 21 = 7 ***
	// 13 x 22 = 4 ***
	// 13 x 23 = 1 ***
	{0, termbox.KeyArrowDown}: &Coord{3, 14}, // 13 x 26
	// 14 x 20 = termbox.KeyF11 *weird behavior on pi*
	// 14 x 21 = 9 ***
	// 14 x 22 = 6 ***
	// 14 x 23 = 3 ***
	// 14 x 24 = . ***
	// 14 x 25 = *
	// 14 x 26 = - ***
	{0, termbox.KeyHome}: &Coord{3, 15}, // 15 x 19
	// 15 x 20 = KeyEnd
	// 15 x 21 = +
	// 15 x 23 = KeyEnter ***
	// 15 x 24 = KeyArrowUp
	// 15 x 25 = [brightness up]
	// 15 x 26 = KeyArrowLeft
	// 16 x 21 = [brightness down]
	// 17 x 24 = [launch itunes?]
	// 18 x 22 = [volume up]
}

// pin counts
// * => >6, requires double bridge
// - => 0, no bridge
// 1   0 -
// 2   0 -
// 3   6
// 4   5
// 5   6
// 6   8 *
// 7   8 *
// 8   7 *
// 9   6
// 10  8 *
// 11  6
// 12  2
// 13  2
// 14  0
// 15  1
// 16  0 -
// 17  0 -
// 18  0 -
// 19 11 *
// 20 10 *
// 21  8 *
// 22  8 *
// 23  8 *
// 24  6
// 25  8 *
// 26  6
// 27  0 -
// 28  0 -
// 29  0 -
// 30  0 -
