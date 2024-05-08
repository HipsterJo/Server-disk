package conversion

import (
	"strconv"
)

func StringToInt(str string, defaultValue int) int {
	if str != "" {
		number, err := strconv.Atoi(str)
		if err != nil {
			return defaultValue
		}

		return number
	}

	return defaultValue
}
