package keyboard

import "github.com/siggy/bbox/bbox2/program"

var KeyMapsPC = map[rune]program.Coord{
	'1': {Row: 0, Col: 0},
	'2': {Row: 0, Col: 1},
	'3': {Row: 0, Col: 2},
	'4': {Row: 0, Col: 3},
	'5': {Row: 0, Col: 4},
	'6': {Row: 0, Col: 5},
	'7': {Row: 0, Col: 6},
	'8': {Row: 0, Col: 7},
	'!': {Row: 0, Col: 8},
	'@': {Row: 0, Col: 9},
	'#': {Row: 0, Col: 10},
	'$': {Row: 0, Col: 11},
	'%': {Row: 0, Col: 12},
	'^': {Row: 0, Col: 13},
	'&': {Row: 0, Col: 14},
	'*': {Row: 0, Col: 15},

	'q': {Row: 1, Col: 0},
	'w': {Row: 1, Col: 1},
	'e': {Row: 1, Col: 2},
	'r': {Row: 1, Col: 3},
	't': {Row: 1, Col: 4},
	'y': {Row: 1, Col: 5},
	'u': {Row: 1, Col: 6},
	'i': {Row: 1, Col: 7},
	'Q': {Row: 1, Col: 8},
	'W': {Row: 1, Col: 9},
	'E': {Row: 1, Col: 10},
	'R': {Row: 1, Col: 11},
	'T': {Row: 1, Col: 12},
	'Y': {Row: 1, Col: 13},
	'U': {Row: 1, Col: 14},
	'I': {Row: 1, Col: 15},

	'a': {Row: 2, Col: 0},
	's': {Row: 2, Col: 1},
	'd': {Row: 2, Col: 2},
	'f': {Row: 2, Col: 3},
	'g': {Row: 2, Col: 4},
	'h': {Row: 2, Col: 5},
	'j': {Row: 2, Col: 6},
	'k': {Row: 2, Col: 7},
	'A': {Row: 2, Col: 8},
	'S': {Row: 2, Col: 9},
	'D': {Row: 2, Col: 10},
	'F': {Row: 2, Col: 11},
	'G': {Row: 2, Col: 12},
	'H': {Row: 2, Col: 13},
	'J': {Row: 2, Col: 14},
	'K': {Row: 2, Col: 15},

	'z': {Row: 3, Col: 0},
	'x': {Row: 3, Col: 1},
	'c': {Row: 3, Col: 2},
	'v': {Row: 3, Col: 3},
	'b': {Row: 3, Col: 4},
	'n': {Row: 3, Col: 5},
	'm': {Row: 3, Col: 6},
	',': {Row: 3, Col: 7},
	'Z': {Row: 3, Col: 8},
	'X': {Row: 3, Col: 9},
	'C': {Row: 3, Col: 10},
	'V': {Row: 3, Col: 11},
	'B': {Row: 3, Col: 12},
	'N': {Row: 3, Col: 13},
	'M': {Row: 3, Col: 14},
	'<': {Row: 3, Col: 15},
}

//
// Overrides:
// space => 8
// enter => 9

// | Col 1 | Col 2 | Col 3 | Col 4 | Col 5 | Col 6 | Col 7 | Col 8 |
// | :-: | :-: | :-: | :-: | :-: | :-: | :-: | :-: |
// | a | b | c | d | e | f | g | h |
// | i | j | k | l | m | n | o | p |
// | q | r | s | t | u | v | w | x |
// | A | B | C | D | E | F | G | H |
// | I | J | K | L | M | N | O | P |
// | Q | R | S | T | U | V | W | X |
// | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |
// | - | = | . | , | / | ; | space | enter |

// mapping from keyboard box
var KeyMapsRPI = map[rune]program.Coord{
	// 2 x 21 = [volume down]
	// 2 x 24 = [mute]
	// 3 x 19 = ` (quit)

	'a': {Row: 0, Col: 0},
	'b': {Row: 0, Col: 1},
	'c': {Row: 0, Col: 2},
	'd': {Row: 0, Col: 3},
	'e': {Row: 0, Col: 4},
	'f': {Row: 0, Col: 5},
	'g': {Row: 0, Col: 6},
	'h': {Row: 0, Col: 7},
	'i': {Row: 0, Col: 8},
	'j': {Row: 0, Col: 9},
	'k': {Row: 0, Col: 10},
	'l': {Row: 0, Col: 11},
	'm': {Row: 0, Col: 12},
	'n': {Row: 0, Col: 13},
	'o': {Row: 0, Col: 14},
	'p': {Row: 0, Col: 15},

	'q': {Row: 1, Col: 0},
	'r': {Row: 1, Col: 1},
	's': {Row: 1, Col: 2},
	't': {Row: 1, Col: 3},
	'u': {Row: 1, Col: 4},
	'v': {Row: 1, Col: 5},
	'w': {Row: 1, Col: 6},
	'x': {Row: 1, Col: 7},
	'A': {Row: 1, Col: 8},
	'B': {Row: 1, Col: 9},
	'C': {Row: 1, Col: 10},
	'D': {Row: 1, Col: 11},
	'E': {Row: 1, Col: 12},
	'F': {Row: 1, Col: 13},
	'G': {Row: 1, Col: 14},
	'H': {Row: 1, Col: 15},

	'I': {Row: 2, Col: 0},
	'J': {Row: 2, Col: 1},
	'K': {Row: 2, Col: 2},
	'L': {Row: 2, Col: 3},
	'M': {Row: 2, Col: 4},
	'N': {Row: 2, Col: 5},
	'O': {Row: 2, Col: 6},
	'P': {Row: 2, Col: 7},
	'Q': {Row: 2, Col: 8},
	'R': {Row: 2, Col: 9},
	'S': {Row: 2, Col: 10},
	'T': {Row: 2, Col: 11},
	'U': {Row: 2, Col: 12},
	'V': {Row: 2, Col: 13},
	'W': {Row: 2, Col: 14},
	'X': {Row: 2, Col: 15},

	'0': {Row: 3, Col: 0},
	'1': {Row: 3, Col: 1},
	'2': {Row: 3, Col: 2},
	'3': {Row: 3, Col: 3},
	'4': {Row: 3, Col: 4},
	'5': {Row: 3, Col: 5},
	'6': {Row: 3, Col: 6},
	'7': {Row: 3, Col: 7},
	'-': {Row: 3, Col: 8},
	'=': {Row: 3, Col: 9},
	'.': {Row: 3, Col: 10},
	',': {Row: 3, Col: 11},
	'/': {Row: 3, Col: 12},
	';': {Row: 3, Col: 13},
	'8': {Row: 3, Col: 14},
	'9': {Row: 3, Col: 15},
}
