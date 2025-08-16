package util

import (
	"time"
)

func Now() int64 {
	return time.Now().UnixNano()
}

func GetCurrentTime() uint64 {
	return uint64(time.Now().UnixNano())
}
