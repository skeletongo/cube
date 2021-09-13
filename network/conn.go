package network

import (
	"io"
	"net"
	"time"
)

// Conn 网络连接
// tcp,udp,websocket等
type Conn interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
	io.ReadWriteCloser
}
