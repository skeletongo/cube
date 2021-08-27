package utils

import (
	"runtime"

	log "github.com/sirupsen/logrus"
)

func DumpStackIfPanic() {
	if err := recover(); err != nil {
		log.Errorf("panic: %s", err)
		var buf [4096]byte
		n := runtime.Stack(buf[:], false)
		log.Errorln(string(buf[:n]))
	}
}
