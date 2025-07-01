package beats

import (
	"fmt"

	"github.com/siggy/bbox/bbox2/program"
)

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
					end:   10,
				},
				{
					strip: 0,
					start: 11,
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
					strip: 1,
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
				{1, 79}, {1, 74}, {1, 69}, {1, 64},
				{1, 53}, {1, 47}, {1, 42}, {1, 37},
				{1, 34}, {1, 29}, {1, 24}, {1, 18},
				{1, 16}, {1, 10}, {1, 5}, {1, 0},
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
				{1, 88}, {1, 93}, {1, 99}, {1, 105},
				{1, 115}, {1, 121}, {1, 127}, {1, 133},
				{1, 136}, {1, 142}, {1, 148}, {1, 154},
				{1, 157}, {1, 163}, {1, 169}, {1, 174},
			},
		},
	}

	flatRows = initRows(rows)
)

func initRows(rows [program.Rows]Row) [program.Rows]flatRow {
	flatRows := [program.Rows]flatRow{}

	for i, row := range rows {
		buttonIndex := 0

		for _, segment := range row.segments {
			start := segment.start
			end := segment.end
			if segment.start > segment.end {
				start = segment.end
				end = segment.start
			}

			for j := start; j <= end; j++ {
				button := false
				if buttonIndex < program.Cols &&
					row.buttons[buttonIndex].strip == segment.strip &&
					row.buttons[buttonIndex].pixel == j {

					flatRows[i].buttons[buttonIndex] = len(flatRows[i].pixels)

					button = true
					buttonIndex++
				}

				flatRows[i].pixels = append(flatRows[i].pixels, pixel{
					strip:  segment.strip,
					pixel:  j,
					button: button,
				})
			}
		}
	}
	fmt.Printf("initRows() ROWS:     %+v\n", rows)
	fmt.Printf("initRows() FLATROWS: %+v\n", flatRows)

	return flatRows

}

type pixel struct {
	strip  int
	pixel  int
	button bool // TODO: needed?
}

// TODO: need a map from button to pixel index?

type flatRow struct {
	// pixels is a flat slice of pixels for an entire row
	pixels []pixel
	// buttons maps button => index in pixels slice
	buttons [program.Cols]int
}
