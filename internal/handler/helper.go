package handler

import "strconv"

func parseUintParam(value string) (uint, error) {
	id, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
