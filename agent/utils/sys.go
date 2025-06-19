package utils

import (
	"os"
	"time"
)

func DelayExit(code int, delay time.Duration) {
	time.Sleep(delay)
	os.Exit(code)
}

func IsDnf() bool {
	_, err := os.Stat("/usr/bin/dnf")
	return err == nil
}
