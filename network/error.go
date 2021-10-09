package network

import (
	"fmt"
)

type ErrorType int

const (
	ErrorTypeMsgID   ErrorType = iota // 未知的消息号
	ErrorTypeEncoder                  // 不支持的编码
)

type Error struct {
	Err  error
	Type ErrorType
	Data interface{}
}

func (e *Error) Error() string {
	return fmt.Sprintf("ErrorType: %v, Error: %v, Data: %v", e.Type, e.Err.Error(), e.Data)
}

func (e *Error) IsType(flags ErrorType) bool {
	return e.Type == flags
}

func NewError(err error, errorType ErrorType, data interface{}) *Error {
	return &Error{
		Err:  err,
		Type: errorType,
		Data: data,
	}
}
