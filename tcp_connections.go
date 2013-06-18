package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/jordanorelli/skeam/cm"
	"io"
	"net"
)

const MAX_SEXP_LINES = 40

var manager = cm.New()
var errSexpTooLong = errors.New("error: sexp is too long")

type tcpInterpreter struct {
	fout   io.Writer        // buffered file-like output stream
	ferr   io.Writer        // buffered file-like error stream
	tokens chan token       // tokens returns from the lexer
	values chan interface{} // values returned from the interpreter
	errors chan error       // errors returned from the interpreter
}

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
		go startConnection(conn, manager)
	}
}

func newTcpInterpreter() *tcpInterpreter {
	return &tcpInterpreter{
		tokens: make(chan token),
		values: make(chan interface{}),
		errors: make(chan error),
	}
}

func (t *tcpInterpreter) send() {
	for {
		select {
		case v := <-t.values:
			if _, err := fmt.Fprintln(t.fout, v); err != nil {
				fmt.Println("can't write out to client: ", err)
			}
		case e := <-t.errors:
			if _, err := fmt.Fprintln(t.ferr, e); err != nil {
				fmt.Println("can't write error to client: ", err)
			}
		}
	}
}

func (t *tcpInterpreter) Run(in io.Reader, out, errors io.Writer) {
	go lex(bufio.NewReader(in), t.tokens)

	t.fout = out
	t.ferr = errors
	go t.send()

	evalall(t.tokens, t.values, t.errors, universe)
}

func startConnection(conn net.Conn, m *cm.Manager) {
	m.Add(conn)
	defer m.Remove(conn)

	i := newTcpInterpreter()
	i.Run(conn, conn, conn)
}
