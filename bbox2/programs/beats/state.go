package beats

type state [soundCount][beatCount]bool

func (s state) String() string {
	var str string
	for row := range soundCount {
		for col := range s[row] {
			if s[row][col] {
				str += "X"
			} else {
				str += "."
			}
		}
		str += "\n"
	}
	return str
}

func (s state) activeButtons() int {
	active := 0

	for _, row := range s {
		for _, beat := range row {
			if beat {
				active++
			}
		}
	}

	return active
}

func (s *state) allOff() bool {
	for _, row := range s {
		for _, beat := range row {
			if beat {
				return false
			}
		}
	}

	return true
}
