package cm

import (
	"net"
)

type Manager struct {
	active     map[net.Conn]bool
	connect    chan net.Conn
	disconnect chan net.Conn
}

func New() *Manager {
	m := &Manager{
		active:     make(map[net.Conn]bool, 10),
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
			m.active[conn] = true
		case conn := <-m.disconnect:
			delete(m.active, conn)
		}
	}
}

func (m *Manager) Add(conn net.Conn) {
	m.connect <- conn
}

func (m *Manager) Remove(conn net.Conn) {
	m.disconnect <- conn
}
