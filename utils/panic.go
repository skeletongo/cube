package utils

import (
	"runtime"

	log "github.com/sirupsen/logrus"
)

func DumpStackIfPanic() {
	if err := recover(); err != nil {
		var buf [4096]byte
		n := runtime.Stack(buf[:], false)
		log.Errorf("panic: %s\n%s\n", err, string(buf[:n]))
	}
}
