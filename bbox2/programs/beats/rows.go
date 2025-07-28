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

// row assumes segments are in contiguous order w.r.t. buttons
type row struct {
	segments []segment
	buttons  [program.Cols]coord
}

var (
	rows = [program.Rows]row{
		{
			segments: []segment{
				{
					strip: 4,
					start: 100,
					end:   0,
				},
				{
					strip: 0,
					start: 143,
					end:   0,
				},
			},
			buttons: [program.Cols]coord{
				{4, 30}, {4, 20}, {4, 10}, {4, 0},
				{0, 96}, {0, 87}, {0, 77}, {0, 68},
				{0, 62}, {0, 52}, {0, 43}, {0, 33},
				{0, 28}, {0, 19}, {0, 9}, {0, 0},
			},
		},
		{
			segments: []segment{
				{
					strip: 5,
					start: 100,
					end:   0,
				},
				{
					strip: 1,
					start: 143,
					end:   0,
				},
			},
			buttons: [program.Cols]coord{
				{5, 33}, {5, 23}, {5, 12}, {5, 0},
				{1, 109}, {1, 99}, {1, 88}, {1, 78},
				{1, 71}, {1, 60}, {1, 49}, {1, 39},
				{1, 32}, {1, 22}, {1, 11}, {1, 0},
			},
		},
		{
			segments: []segment{
				{
					strip: 6,
					start: 100,
					end:   0,
				},
				{
					strip: 2,
					start: 143,
					end:   0,
				},
			},
			buttons: [program.Cols]coord{
				{6, 38}, {6, 25}, {6, 13}, {6, 0},
				{2, 123}, {2, 112}, {2, 99}, {2, 87},
				{2, 80}, {2, 68}, {2, 55}, {2, 43},
				{2, 36}, {2, 24}, {2, 12}, {2, 0},
			},
		},
		{
			segments: []segment{
				{
					strip: 7,
					start: 100,
					end:   0,
				},
				{
					strip: 3,
					start: 143,
					end:   0,
				},
			},
			buttons: [program.Cols]coord{
				{7, 41}, {7, 27}, {7, 14}, {7, 0},
				{3, 137}, {3, 124}, {3, 111}, {3, 97},
				{3, 89}, {3, 75}, {3, 61}, {3, 48},
				{3, 41}, {3, 28}, {3, 14}, {3, 1},
			},
		},
	}

	flatRows = initRows(rows)
)

func initRows(rows [program.Rows]row) [program.Rows]flatRow {
	flatRows := [program.Rows]flatRow{}

	for i, row := range rows {
		buttonIndex := 0

		for _, segment := range row.segments {
			start := segment.start
			end := segment.end
			step := 1
			if segment.start > segment.end {
				step = -1
			}

			for j := start; ; j += step {
				if (step > 0 && j > end) || (step < 0 && j < end) {
					break
				}

				if buttonIndex < program.Cols &&
					row.buttons[buttonIndex].strip == segment.strip &&
					row.buttons[buttonIndex].pixel == j {
					flatRows[i].buttons[buttonIndex] = len(flatRows[i].pixels)
					buttonIndex++
				}

				flatRows[i].pixels = append(flatRows[i].pixels, coord{
					strip: segment.strip,
					pixel: j,
				})
			}
		}
		if buttonIndex != program.Cols {
			panic(fmt.Sprintf("Row %d: expected %d buttons, got %d", i, program.Cols, buttonIndex))
		}
	}

	return flatRows
}

type flatRow struct {
	// pixels is a flat slice of pixels for an entire row
	pixels []coord
	// buttons maps button => index in pixels slice
	buttons [program.Cols]int
}
