package cm

import (
	"net"
)

type Manager struct {
	active     []net.Conn
	connect    chan net.Conn
	disconnect chan net.Conn
}

func New() *Manager {
	m := &Manager{
		active:     make([]net.Conn, 0, 10),
		connect:    make(chan net.Conn),
		disconnect: make(chan net.Conn),
	}
	go m.run()
	return m
}

func (m *Manager) run() {
	for {
		select {
		case conn := <-m.connect:
			m.active = append(m.active, conn)
		case conn := <-m.disconnect:
			m.removeConnection(conn)
		}
	}
}

func (m *Manager) Add(conn net.Conn) {
	m.connect <- conn
}

func (m *Manager) Remove(conn net.Conn) {
	m.disconnect <- conn
}

func (m *Manager) removeConnection(conn net.Conn) {
	for i, other := range m.active {
		if conn.RemoteAddr() == other.RemoteAddr() {
			m.active = append(m.active[:i], m.active[i+1:]...)
			return
		}
	}
}
