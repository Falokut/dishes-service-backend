package controller

import (
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

func stringToIntSlice(s string) ([]int32, error) {
	if s == "" {
		return []int32{}, nil
	}
	strList := strings.Split(s, ",")
	res := make([]int32, len(strList))

	for i, str := range strList {
		num, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return nil, errors.WithMessage(err, "parse num")
		}
		res[i] = int32(num)
	}
	return res, nil
}
