package utils

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"sync"

	log "github.com/sirupsen/logrus"
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

// FileLineHook 设置打印文件路径及行号
type FileLineHook struct {
	LogLevels []log.Level // 需要打印的日志级别
	Skip      int         // 跳过几层调用栈
	Test      bool        // 打印所有调用栈信息，找出合适的 Skip 配置
}

func (e *FileLineHook) Levels() []log.Level {
	return e.LogLevels
}

func (e *FileLineHook) Fire(entry *log.Entry) error {
	_, filename, line, _ := runtime.Caller(e.Skip)
	entry.Data["source"] = fmt.Sprintf("%s:%d", filename, line)
	if e.Test {
		buf := [4096]byte{}
		n := runtime.Stack(buf[:], false)
		fmt.Println(string(buf[:n]))
	}
	return nil
}
