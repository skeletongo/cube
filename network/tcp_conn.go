package network

import (
	"net"
)

type TCPConn struct {
	net.Conn
	config *ServiceConfig
}

func NewTCPConn(conn net.Conn, config *ServiceConfig) *TCPConn {
	t := &TCPConn{
		Conn:   conn,
		config: config,
	}
	return t
}
