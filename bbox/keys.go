package bbox

import (
	"github.com/nsf/termbox-go"
)

type Coord [2]int

var keymaps = map[string]*Coord{
	"1": &Coord{0, 0},
	"2": &Coord{0, 1},
	"3": &Coord{0, 2},
	"4": &Coord{0, 3},
	"5": &Coord{0, 4},
	"6": &Coord{0, 5},
	"7": &Coord{0, 6},
	"8": &Coord{0, 7},
	"!": &Coord{0, 8},
	"@": &Coord{0, 9},
	"#": &Coord{0, 10},
	"$": &Coord{0, 11},
	"%": &Coord{0, 12},
	"^": &Coord{0, 13},
	"&": &Coord{0, 14},
	"*": &Coord{0, 15},

	"w": &Coord{1, 0},
	"e": &Coord{1, 1},
	"r": &Coord{1, 2},
	"t": &Coord{1, 3},
	"y": &Coord{1, 4},
	"u": &Coord{1, 5},
	"i": &Coord{1, 6},
	"o": &Coord{1, 7},
	"W": &Coord{1, 8},
	"E": &Coord{1, 9},
	"R": &Coord{1, 10},
	"T": &Coord{1, 11},
	"Y": &Coord{1, 12},
	"U": &Coord{1, 13},
	"I": &Coord{1, 14},
	"O": &Coord{1, 15},

	"a": &Coord{2, 0},
	"s": &Coord{2, 1},
	"d": &Coord{2, 2},
	"f": &Coord{2, 3},
	"g": &Coord{2, 4},
	"h": &Coord{2, 5},
	"j": &Coord{2, 6},
	"k": &Coord{2, 7},
	"A": &Coord{2, 8},
	"S": &Coord{2, 9},
	"D": &Coord{2, 10},
	"F": &Coord{2, 11},
	"G": &Coord{2, 12},
	"H": &Coord{2, 13},
	"J": &Coord{2, 14},
	"K": &Coord{2, 15},

	"z": &Coord{3, 0},
	"x": &Coord{3, 1},
	"c": &Coord{3, 2},
	"v": &Coord{3, 3},
	"b": &Coord{3, 4},
	"n": &Coord{3, 5},
	"m": &Coord{3, 6},
	",": &Coord{3, 7},
	"Z": &Coord{3, 8},
	"X": &Coord{3, 9},
	"C": &Coord{3, 10},
	"V": &Coord{3, 11},
	"B": &Coord{3, 12},
	"N": &Coord{3, 13},
	"M": &Coord{3, 14},
	"<": &Coord{3, 15},
}

type Key struct {
	Ch  rune        // a unicode character
	Key termbox.Key // one of Key* constants, invalid if 'Ch' is not 0
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
	{'z', 0}:            &Coord{0, 4}, // 3 x 24
	{0, termbox.KeyF1}:  &Coord{0, 5}, // 4 x 19
	{'2', 0}:            &Coord{0, 6}, // 4 x 20
	{'w', 0}:            &Coord{0, 7}, // 4 x 21
	{'S', 0}:            &Coord{0, 8}, // 4 x 23
	// 4 x 24 = §
	{'x', 0}:           &Coord{0, 9},  // 4 x 25
	{0, termbox.KeyF2}: &Coord{0, 10}, // 5 x 19
	{'3', 0}:           &Coord{0, 11}, // 5 x 20
	{'e', 0}:           &Coord{0, 12}, // 5 x 21
	{'d', 0}:           &Coord{0, 13}, // 5 x 22
	{'c', 0}:           &Coord{0, 14}, // 5 x 23
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
	{'K', 0}:           &Coord{2, 4},  // 8 x 23
	{0, termbox.KeyF6}: &Coord{2, 5},  // 8 x 24
	{',', 0}:           &Coord{2, 6},  // 8 x 25
	{0, termbox.KeyF8}: &Coord{2, 7},  // 9 x 19
	{'9', 0}:           &Coord{2, 8},  // 9 x 20
	{'o', 0}:           &Coord{2, 9},  // 9 x 21
	{'l', 0}:           &Coord{2, 10}, // 9 x 23
	{'.', 0}:           &Coord{2, 11}, // 9 x 25
	{'-', 0}:           &Coord{2, 12}, // 10 x 19
	{'0', 0}:           &Coord{2, 13}, // 10 x 20
	{'p', 0}:           &Coord{2, 14}, // 10 x 21
	{'[', 0}:           &Coord{2, 15}, // 10 x 22

	{';', 0}:                   &Coord{3, 0}, // 10 x 23
	{'\'', 0}:                  &Coord{3, 1}, // 10 x 24
	{'\\', 0}:                  &Coord{3, 2}, // 10 x 25
	{'/', 0}:                   &Coord{3, 3}, // 10 x 26
	{0, termbox.KeyF9}:         &Coord{3, 4}, // 11 x 19
	{0, termbox.KeyF10}:        &Coord{3, 5}, // 11 x 20
	{0, termbox.KeyBackspace2}: &Coord{3, 6}, // 11 x 22
	// 11 x 23 = \ ***
	{0, termbox.KeyF5}:    &Coord{3, 7},  // 11 x 24
	{0, termbox.KeyEnter}: &Coord{3, 8},  // 11 x 25
	{0, termbox.KeySpace}: &Coord{3, 9},  // 11 x 26
	{0, termbox.KeyF12}:   &Coord{3, 10}, // 12 x 20
	// 12 x 21 = 8 ***
	// 12 x 22 = 5 ***
	// 12 x 23 = 2 ***
	// 12 x 24 = 0 ***
	// 12 x 25 = / ***
	{0, termbox.KeyArrowRight}: &Coord{3, 11}, // 12 x 26
	{0, termbox.KeyDelete}:     &Coord{3, 12}, // 13 x 19
	// 13 x 20 = [fn f11]
	// 13 x 21 = 7 ***
	// 13 x 22 = 4 ***
	// 13 x 23 = 1 ***
	{0, termbox.KeyArrowDown}: &Coord{3, 13}, // 13 x 26
	{0, termbox.KeyPgup}:      &Coord{3, 14}, // 14 x 19
	{0, termbox.KeyPgdn}:      &Coord{3, 15}, // 14 x 20
	// 14 x 21 = 9 ***
	// 14 x 22 = 6 ***
	// 14 x 23 = 3 ***
	// 14 x 24 = . ***
	// 14 x 25 = *
	// 14 x 26 = - ***
	// 15 x 19 = KeyHome
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
