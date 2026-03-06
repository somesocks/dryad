package net

import (
	"dryad/diagnostics"
	stdnet "net"
	"time"
)

type Conn interface {
	stdnet.Conn
}

type Listener interface {
	Accept() (Conn, error)
	Close() error
	Addr() stdnet.Addr
}

type connWrap struct {
	conn stdnet.Conn
}

type listenerWrap struct {
	listener stdnet.Listener
}

var ErrClosed = stdnet.ErrClosed

func connKey(conn stdnet.Conn) string {
	if conn == nil {
		return ""
	}

	remote := ""
	if addr := conn.RemoteAddr(); addr != nil {
		remote = addr.String()
	}
	if remote != "" {
		return remote
	}

	if addr := conn.LocalAddr(); addr != nil {
		return addr.String()
	}

	return ""
}

func listenerKey(listener stdnet.Listener) string {
	if listener == nil {
		return ""
	}
	if addr := listener.Addr(); addr != nil {
		return addr.String()
	}
	return ""
}

var dial = diagnostics.BindA2R1(
	"net.dial",
	func(network string, address string) string {
		return network + ":" + address
	},
	func(network string, address string) (error, stdnet.Conn) {
		conn, err := stdnet.Dial(network, address)
		return err, conn
	},
)

var Dial = func(network string, address string) (Conn, error) {
	err, conn := dial(network, address)
	if err != nil {
		return nil, err
	}
	return &connWrap{conn: conn}, nil
}

var listen = diagnostics.BindA2R1(
	"net.listen",
	func(network string, address string) string {
		return network + ":" + address
	},
	func(network string, address string) (error, stdnet.Listener) {
		listener, err := stdnet.Listen(network, address)
		return err, listener
	},
)

var Listen = func(network string, address string) (Listener, error) {
	err, listener := listen(network, address)
	if err != nil {
		return nil, err
	}
	return &listenerWrap{listener: listener}, nil
}

var accept = diagnostics.BindA1R1(
	"net.accept",
	listenerKey,
	func(listener stdnet.Listener) (error, stdnet.Conn) {
		conn, err := listener.Accept()
		return err, conn
	},
)

var read = diagnostics.BindA2R1(
	"net.read",
	func(conn stdnet.Conn, _ []byte) string {
		return connKey(conn)
	},
	func(conn stdnet.Conn, b []byte) (error, int) {
		n, err := conn.Read(b)
		return err, n
	},
)

var write = diagnostics.BindA2R1(
	"net.write",
	func(conn stdnet.Conn, _ []byte) string {
		return connKey(conn)
	},
	func(conn stdnet.Conn, b []byte) (error, int) {
		n, err := conn.Write(b)
		return err, n
	},
)

var closeConn = diagnostics.BindA1R0(
	"net.close",
	connKey,
	func(conn stdnet.Conn) error {
		return conn.Close()
	},
)

var closeListener = diagnostics.BindA1R0(
	"net.listener_close",
	listenerKey,
	func(listener stdnet.Listener) error {
		return listener.Close()
	},
)

func (listener *listenerWrap) Accept() (Conn, error) {
	err, conn := accept(listener.listener)
	if err != nil {
		return nil, err
	}
	return &connWrap{conn: conn}, nil
}

func (listener *listenerWrap) Close() error {
	return closeListener(listener.listener)
}

func (listener *listenerWrap) Addr() stdnet.Addr {
	return listener.listener.Addr()
}

func (conn *connWrap) Read(b []byte) (int, error) {
	err, n := read(conn.conn, b)
	return n, err
}

func (conn *connWrap) Write(b []byte) (int, error) {
	err, n := write(conn.conn, b)
	return n, err
}

func (conn *connWrap) Close() error {
	return closeConn(conn.conn)
}

func (conn *connWrap) LocalAddr() stdnet.Addr {
	return conn.conn.LocalAddr()
}

func (conn *connWrap) RemoteAddr() stdnet.Addr {
	return conn.conn.RemoteAddr()
}

func (conn *connWrap) SetDeadline(t time.Time) error {
	return conn.conn.SetDeadline(t)
}

func (conn *connWrap) SetReadDeadline(t time.Time) error {
	return conn.conn.SetReadDeadline(t)
}

func (conn *connWrap) SetWriteDeadline(t time.Time) error {
	return conn.conn.SetWriteDeadline(t)
}
