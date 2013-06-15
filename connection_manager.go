package main

import (
	"net"
)

var manager = newConnectionManager()

type connectionManager struct {
	active     []net.Conn
	connect    chan net.Conn
	disconnect chan net.Conn
}

func newConnectionManager() *connectionManager {
	m := &connectionManager{
		active:     make([]net.Conn, 0, 10),
		connect:    make(chan net.Conn),
		disconnect: make(chan net.Conn),
	}
	go m.run()
	return m
}

func (m *connectionManager) run() {
	for {
		select {
		case conn := <-m.connect:
			m.active = append(m.active, conn)
		case conn := <-m.disconnect:
			m.removeConnection(conn)
		}
	}
}

func (m *connectionManager) Add(conn net.Conn) {
	m.connect <- conn
}

func (m *connectionManager) Remove(conn net.Conn) {
	m.disconnect <- conn
}

func (m *connectionManager) removeConnection(conn net.Conn) {
	for i, other := range m.active {
		if conn.RemoteAddr() == other.RemoteAddr() {
			m.active = append(m.active[:i], m.active[i+1:]...)
			return
		}
	}
}
