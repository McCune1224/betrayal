package util

import "strconv"

func Atoi64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func Atoi32(s string) (int32, error) {
	res, _ := strconv.ParseInt(s, 10, 32)
	if res > int64(int32(res)) {
		return 0, strconv.ErrRange
	}
	return int32(res), nil
}

func Itoa64(i int64) string {
	return strconv.FormatInt(i, 10)
}

func Itoa32(i int32) string {
	return strconv.FormatInt(int64(i), 10)
}
