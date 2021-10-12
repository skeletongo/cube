package tools

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

// FileLineHook 设置打印文件路径及行号
type FileLineHook struct {
	LogLevels []logrus.Level // 需要打印的日志级别
	FieldName string         // 名称
	Skip      int            // 跳过几层调用栈
	Num       int            // Skip后的查找范围
	Test      bool           // 打印所有调用栈信息，找出合适的 Skip 配置
	filename  string
	line      int
}

func (e *FileLineHook) Levels() []logrus.Level {
	return e.LogLevels
}

func (e *FileLineHook) Fire(entry *logrus.Entry) error {
	for i := 0; i < e.Num; i++ {
		_, e.filename, e.line, _ = runtime.Caller(e.Skip + i)
		if !strings.Contains(e.filename, "logrus") {
			break
		}
	}
	entry.Data[e.FieldName] = fmt.Sprintf("%s:%d", e.filename, e.line)
	if e.Test {
		buf := [4096]byte{}
		n := runtime.Stack(buf[:], false)
		fmt.Println(string(buf[:n]))
	}
	return nil
}

// NewFileLineHook 打印文件路径及行号
// levels 指定日志级别
func NewFileLineHook(levels ...logrus.Level) logrus.Hook {
	return &FileLineHook{
		LogLevels: levels,
		FieldName: "source",
		Skip:      8,
		Num:       2,
	}
}
