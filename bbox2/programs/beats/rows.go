package beats

import "github.com/siggy/bbox/bbox2/program"

type Row struct {
	start   int
	end     int
	buttons [program.Cols]int
}

var (
	rows = [program.Rows]Row{
		// test strip 0-143
		{
			start: 0,
			end:   143,
			buttons: [program.Cols]int{
				1, 10, 15, 20,
				25, 30, 35, 40,
				50, 60, 70, 80,
				95, 110, 125, 143,
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
			start: 72,
			end:   151,
			buttons: [program.Cols]int{
				75, 79, 83, 88,
				103, 108, 112, 117,
				119, 124, 128, 133,
				136, 140, 145, 150,
			},
		},

		// rows 2 and 3 are LED strip 1
		{
			start: 83,
			end:   0,
			buttons: [program.Cols]int{
				79, 74, 69, 64,
				53, 47, 42, 37,
				34, 29, 24, 18,
				16, 10, 5, 0,
			},
		},
		{
			start: 84,
			end:   176,
			buttons: [program.Cols]int{
				88, 93, 99, 105,
				115, 121, 127, 133,
				136, 142, 148, 154,
				157, 163, 169, 174,
			},
		},
	}
)
