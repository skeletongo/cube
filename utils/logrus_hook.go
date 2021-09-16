package utils

import (
	"fmt"
	"runtime"

	"github.com/sirupsen/logrus"
)

// FileLineHook 设置打印文件路径及行号
type FileLineHook struct {
	LogLevels []logrus.Level // 需要打印的日志级别
	Skip      int            // 跳过几层调用栈
	Test      bool           // 打印所有调用栈信息，找出合适的 Skip 配置
}

func (e *FileLineHook) Levels() []logrus.Level {
	return e.LogLevels
}

func (e *FileLineHook) Fire(entry *logrus.Entry) error {
	_, filename, line, _ := runtime.Caller(e.Skip)
	entry.Data["source"] = fmt.Sprintf("%s:%d", filename, line)
	if e.Test {
		buf := [4096]byte{}
		n := runtime.Stack(buf[:], false)
		fmt.Println(string(buf[:n]))
	}
	return nil
}

func NewFileLineHook(levels ...logrus.Level) logrus.Hook {
	return &FileLineHook{
		LogLevels: levels,
		Skip:      8,
	}
}
