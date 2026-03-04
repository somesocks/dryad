package net

import (
	"dryad/diagnostics"
	stdnet "net"
)

type Conn = stdnet.Conn
type Listener = stdnet.Listener

var ErrClosed = stdnet.ErrClosed

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

var Dial = func(network string, address string) (stdnet.Conn, error) {
	err, conn := dial(network, address)
	return conn, err
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

var Listen = func(network string, address string) (stdnet.Listener, error) {
	err, listener := listen(network, address)
	return listener, err
}
