package network

import (
	"net"
	"time"
)

type TCPConn struct {
	net.Conn
}

func NewTCPConn(conn net.Conn, config *ServiceConfig) (*TCPConn, error) {
	t := &TCPConn{
		Conn: conn,
	}

	var err error
	c := conn.(*net.TCPConn)
	if config.Linger > 0 {
		if err = c.SetLinger(config.Linger); err != nil {
			return nil, err
		}
	}
	if err = c.SetKeepAlive(config.KeepAlive); err != nil {
		return nil, err
	}
	if config.KeepAlive && config.KeepAlivePeriod > 0 {
		if err = c.SetKeepAlivePeriod(config.KeepAlivePeriod); err != nil {
			return nil, err
		}
	}
	if config.ReadBuffer > 0 {
		if err = c.SetReadBuffer(config.ReadBuffer); err != nil {
			return nil, err
		}
	}
	if config.WriteBuffer > 0 {
		if err = c.SetWriteBuffer(config.WriteBuffer); err != nil {
			return nil, err
		}
	}
	return t, err
}

func (c *TCPConn) SetReadDeadline(t time.Time) error {
	return c.Conn.(*net.TCPConn).SetReadDeadline(t)
}

func (c *TCPConn) SetWriteDeadline(t time.Time) error {
	return c.Conn.(*net.TCPConn).SetWriteDeadline(t)
}

func (c *TCPConn) Close() error {
	c.Conn.(*net.TCPConn).SetLinger(0)
	return c.Conn.Close()
}
