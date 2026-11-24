package api

import "strconv"

func integerValue(value string) (int64, error) {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, badRequest()
	}
	return parsed, nil
}
