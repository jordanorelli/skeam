package cm

import (
	"net"
	"strings"
)

type Manager struct {
	active     map[net.Conn]bool
	connect    chan net.Conn
	disconnect chan net.Conn
	write      chan writeOp
}

func New() *Manager {
	m := &Manager{
		active:     make(map[net.Conn]bool, 10),
		connect:    make(chan net.Conn),
		disconnect: make(chan net.Conn),
		write:      make(chan writeOp),
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
			conn.Close()
			delete(m.active, conn)
		case op := <-m.write:
			m.broadcast(op)
		}
	}
}

func (m *Manager) broadcast(op writeOp) {
	var res writeResponse
	for conn, _ := range m.active {
		res.add(conn.Write(op.data))
	}
	op.reply <- res
}

func (m *Manager) Add(conn net.Conn) {
	m.connect <- conn
}

func (m *Manager) Remove(conn net.Conn) {
	m.disconnect <- conn
}

func (m *Manager) Write(b []byte) (int, error) {
	op := *newWriteOp(b)
	m.write <- op
	res := <-op.reply
	return res.i, res.e
}

type multiError []error

func (e multiError) add(err error) {
	if e == nil {
		e = make([]error, 0, 4)
	}
	e = append(e, err)
}

func (e multiError) Error() string {
	messages := make([]string, len(e))
	for i, _ := range e {
		messages[i] = e[i].Error()
	}
	return strings.Join(messages, " | ")
}

type writeResponse struct {
	i int
	e multiError
}

func (res writeResponse) add(n int, e error) {
	if e != nil {
		res.e.add(e)
	}
	res.i += n
}

type writeOp struct {
	data  []byte
	reply chan writeResponse
}

func newWriteOp(b []byte) *writeOp {
	return &writeOp{
		data:  b,
		reply: make(chan writeResponse),
	}
}
