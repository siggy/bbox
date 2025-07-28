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
					strip: 0,
					start: 0,
					end:   143,
				},
				{
					strip: 4,
					start: 0,
					end:   100,
				},
			},
			buttons: [program.Cols]coord{
				{0, 1}, {0, 10}, {0, 15}, {0, 20},
				{0, 30}, {0, 40}, {0, 50}, {0, 60},
				{0, 80}, {0, 100}, {0, 120}, {0, 140},
				{4, 0}, {4, 10}, {4, 20}, {4, 50},
			},
		},
		{
			segments: []segment{
				{
					strip: 1,
					start: 0,
					end:   143,
				},
				{
					strip: 5,
					start: 0,
					end:   100,
				},
			},
			buttons: [program.Cols]coord{
				{1, 1}, {1, 10}, {1, 15}, {1, 20},
				{1, 30}, {1, 40}, {1, 50}, {1, 60},
				{1, 80}, {1, 100}, {1, 120}, {1, 140},
				{5, 0}, {5, 10}, {5, 20}, {5, 50},
			},
		},
		{
			segments: []segment{
				{
					strip: 2,
					start: 0,
					end:   143,
				},
				{
					strip: 6,
					start: 0,
					end:   100,
				},
			},
			buttons: [program.Cols]coord{
				{2, 1}, {2, 10}, {2, 15}, {2, 20},
				{2, 30}, {2, 40}, {2, 50}, {2, 60},
				{2, 80}, {2, 100}, {2, 120}, {2, 140},
				{6, 0}, {6, 10}, {6, 20}, {6, 50},
			},
		},
		{
			segments: []segment{
				{
					strip: 3,
					start: 0,
					end:   143,
				},
				{
					strip: 7,
					start: 0,
					end:   100,
				},
			},
			buttons: [program.Cols]coord{
				{3, 1}, {3, 10}, {3, 15}, {3, 20},
				{3, 30}, {3, 40}, {3, 50}, {3, 60},
				{3, 80}, {3, 100}, {3, 120}, {3, 140},
				{7, 0}, {7, 10}, {7, 20}, {7, 50},
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
