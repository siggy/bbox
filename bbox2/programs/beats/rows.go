package beats

import "github.com/siggy/bbox/bbox2/program"

// define a set of pixels on a given strip, inclusive
type segment struct {
	strip int
	start int
	end   int
}

type coord struct {
	strip int
	pixel int
}

// TODO: make private?
// assume segments are in contiguous order w.r.t. buttons
type Row struct {
	segments []segment
	buttons  [program.Cols]coord
}

var (
	rows = [program.Rows]Row{
		// test strip 0-143
		{
			segments: []segment{
				{
					strip: 0,
					start: 0,
					end:   143,
				},
			},
			buttons: [program.Cols]coord{
				{0, 1}, {0, 10}, {0, 15}, {0, 20},
				{0, 25}, {0, 30}, {0, 35}, {0, 40},
				{0, 50}, {0, 60}, {0, 70}, {0, 80},
				{0, 95}, {0, 110}, {0, 125}, {0, 143},
			},
		},
		// rows 0 and 1 are LED strip 0
		// {
		// 	start: 71,
		// 	end:   0,
		// 	buttons: [program.Cols]int{
		// 		68, 64, 60, 56,
		// 		41, 37, 33, 29,
		// 		27, 23, 19, 15,
		// 		13, 9, 5, 1,
		// 	},
		// },
		{
			segments: []segment{
				{
					strip: 0,
					start: 72,
					end:   151,
				},
			},
			buttons: [program.Cols]coord{
				{0, 75}, {0, 79}, {0, 83}, {0, 88},
				{0, 103}, {0, 108}, {0, 112}, {0, 117},
				{0, 119}, {0, 124}, {0, 128}, {0, 133},
				{0, 136}, {0, 140}, {0, 145}, {0, 150},
			},
		},

		// rows 2 and 3 are LED strip 1
		{
			segments: []segment{
				{
					strip: 1,
					start: 83,
					end:   0,
				},
			},
			buttons: [program.Cols]coord{
				{0, 79}, {0, 74}, {0, 69}, {0, 64},
				{0, 53}, {0, 47}, {0, 42}, {0, 37},
				{0, 34}, {0, 29}, {0, 24}, {0, 18},
				{0, 16}, {0, 10}, {0, 5}, {0, 0},
			},
		},
		{
			segments: []segment{
				{
					strip: 1,
					start: 84,
					end:   176,
				},
			},
			buttons: [program.Cols]coord{
				{0, 88}, {0, 93}, {0, 99}, {0, 105},
				{0, 115}, {0, 121}, {0, 127}, {0, 133},
				{0, 136}, {0, 142}, {0, 148}, {0, 154},
				{0, 157}, {0, 163}, {0, 169}, {0, 174},
			},
		},
	}
)
