package tools

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"sync"
)

var RecoverPanicFunc func(args ...interface{})

func init() {
	bufPool := &sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}
	RecoverPanicFunc = func(args ...interface{}) {
		if r := recover(); r != nil {
			buf := bufPool.Get().(*bytes.Buffer)
			defer bufPool.Put(buf)
			buf.Reset()
			buf.WriteString(fmt.Sprintf("panic: %v\n", r))
			for _, v := range args {
				buf.WriteString(fmt.Sprintf("%v\n", v))
			}
			pcs := make([]uintptr, 10)
			n := runtime.Callers(3, pcs)
			frames := runtime.CallersFrames(pcs[:n])
			for f, again := frames.Next(); again; f, again = frames.Next() {
				buf.WriteString(fmt.Sprintf("%v:%v %v\n", f.File, f.Line, f.Function))
			}
			fmt.Fprint(os.Stderr, buf.String())
		}
	}
}
