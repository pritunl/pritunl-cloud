package utils

import (
	"os"
	"time"
)

func DelayExit(code int, delay time.Duration) {
	time.Sleep(delay)
	os.Exit(code)
}
