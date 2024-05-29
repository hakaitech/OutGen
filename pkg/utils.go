package pkg

import (
	"fmt"
	"time"
)

func GetTimestamp(dstr, tstr string) (int64, error) {
	if tstr == "" {
		format := "02-01-2006"
		t, err := time.Parse(format, dstr)
		if err != nil {
			return -1, err
		}
		return t.Unix(), nil
	}
	format := "02-01-2006 15:04:05"
	preconv := fmt.Sprintf("%s %s", dstr, tstr)
	t, err := time.Parse(format, preconv)
	if err != nil {
		return -1, err
	}
	return t.Unix(), nil
}
