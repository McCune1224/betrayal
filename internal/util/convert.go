package util

import (
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"
)

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

// https://github.com/sqlc-dev/sqlc/discussions/2910
func Numeric(number float64) (pgtype.Numeric, error) {
	value := pgtype.Numeric{}
	parse := strconv.FormatFloat(number, 'f', -1, 64)
	if err := value.Scan(parse); err != nil {
		return pgtype.Numeric{}, err
	}
	return value, nil
}

func NumericToString(num pgtype.Numeric) (string, error) {
	val, err := num.Value()
	if err != nil {
		return "", err
	}
	s := val.(string)
	return s, nil
}
