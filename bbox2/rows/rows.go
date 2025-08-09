package rows

import (
	"fmt"

	"github.com/siggy/bbox/bbox2/program"
)

// define a set of pixels on a given strip, inclusive
type Segment struct {
	strip int
	start int
	end   int
}

type Coord struct {
	Strip int
	Pixel int
}

// Row assumes segments are in contiguous order w.r.t. buttons.
type Row struct {
	segments []Segment
	Buttons  [program.Cols]Coord
}

// FlatRow describes the physical order of pixels.
type FlatRow struct {
	// pixels is a flat slice of pixels for an entire row
	Pixels []Coord
	// buttons maps button => index in Pixels slice
	Buttons [program.Cols]int
}

var (
	Rows = [program.Rows]Row{
		{
			segments: []Segment{
				{
					strip: 4,
					start: 31,
					end:   0,
				},
				{
					strip: 0,
					start: 98,
					end:   0,
				},
			},
			Buttons: [program.Cols]Coord{
				{4, 30}, {4, 20}, {4, 10}, {4, 0},
				{0, 96}, {0, 87}, {0, 77}, {0, 68},
				{0, 62}, {0, 52}, {0, 43}, {0, 33},
				{0, 28}, {0, 19}, {0, 9}, {0, 0},
			},
		},
		{
			segments: []Segment{
				{
					strip: 5,
					start: 35,
					end:   0,
				},
				{
					strip: 1,
					start: 111,
					end:   0,
				},
			},
			Buttons: [program.Cols]Coord{
				{5, 33}, {5, 23}, {5, 12}, {5, 0},
				{1, 109}, {1, 99}, {1, 88}, {1, 78},
				{1, 71}, {1, 60}, {1, 49}, {1, 39},
				{1, 32}, {1, 22}, {1, 11}, {1, 0},
			},
		},
		{
			segments: []Segment{
				{
					strip: 6,
					start: 40,
					end:   0,
				},
				{
					strip: 2,
					start: 126,
					end:   0,
				},
			},
			Buttons: [program.Cols]Coord{
				{6, 38}, {6, 25}, {6, 13}, {6, 0},
				{2, 123}, {2, 112}, {2, 99}, {2, 87},
				{2, 80}, {2, 68}, {2, 55}, {2, 43},
				{2, 36}, {2, 24}, {2, 12}, {2, 0},
			},
		},
		{
			segments: []Segment{
				{
					strip: 7,
					start: 44,
					end:   0,
				},
				{
					strip: 3,
					start: 140,
					end:   0,
				},
			},
			Buttons: [program.Cols]Coord{
				{7, 41}, {7, 27}, {7, 14}, {7, 0},
				{3, 137}, {3, 124}, {3, 111}, {3, 97},
				{3, 89}, {3, 75}, {3, 61}, {3, 48},
				{3, 41}, {3, 28}, {3, 14}, {3, 1},
			},
		},
	}

	FlatRows = initFlatRows(Rows)
)

func initFlatRows(rows [program.Rows]Row) [program.Rows]FlatRow {
	flatRows := [program.Rows]FlatRow{}

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
					row.Buttons[buttonIndex].Strip == segment.strip &&
					row.Buttons[buttonIndex].Pixel == j {
					flatRows[i].Buttons[buttonIndex] = len(flatRows[i].Pixels)
					buttonIndex++
				}

				flatRows[i].Pixels = append(flatRows[i].Pixels, Coord{
					Strip: segment.strip,
					Pixel: j,
				})
			}
		}
		if buttonIndex != program.Cols {
			panic(fmt.Sprintf("Row %d: expected %d buttons, got %d", i, program.Cols, buttonIndex))
		}
	}

	return flatRows
}
