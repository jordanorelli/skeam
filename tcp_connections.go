package main

import (
	"github.com/jordanorelli/skeam/cm"
	"net"
)

const MAX_SEXP_LINES = 40

var manager = cm.New()

func runTCPServer() {
	addr, err := net.ResolveTCPAddr("tcp", *tcpAddr)
	if err != nil {
		die(err.Error())
	}
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		die(err.Error())
	}
	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			printErrorMsg(err.Error())
			continue
		}
		go startTCPConnection(conn, manager)
	}
}

func startTCPConnection(conn net.Conn, m *cm.Manager) {
	m.Add(conn)
	defer m.Remove(conn)

	i := newInterpreter(conn, conn, conn)
	i.run(universe)
}
