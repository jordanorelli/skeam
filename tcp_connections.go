package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/jordanorelli/skeam/cm"
	"io"
	"net"
	"strings"
)

const MAX_SEXP_LINES = 40

var manager = cm.New()
var errSexpTooLong = errors.New("error: sexp is too long")

func depth(s string) int {
	n := 0
	for _, r := range s {
		switch r {
		case '(':
			n += 1
		case ')':
			n -= 1
		case ';':
			return n
		}
	}
	return n
}

func tcpInterpreter(conn net.Conn, userinput chan string, out chan interface{}, errors chan error) {
	lines := make([]string, 0, MAX_SEXP_LINES)
	currentDepth := 0
	addLine := func(line string) error {
		if len(lines) >= MAX_SEXP_LINES {
			return errSexpTooLong
		}
		lines = append(lines, line+"\n")
		return nil
	}
	errorMode := false
	skipLine := func(line string) {
		currentDepth += depth(line)
		if currentDepth == 0 {
			errorMode = false
		}
	}
	s := make(chan string)
	go func() {
		for program := range s {
			tokens := make(chan token, 32)
			go lexs(program, tokens)
			evalall(tokens, out, errors, universe)
		}
	}()
	go func() {
		for v := range out {
			fmt.Fprintln(conn, v)
		}
	}()
	for {
		select {
		case err := <-errors:
			fmt.Fprintf(conn, "error: %v\n", err)
		case line := <-userinput:
			if errorMode {
				skipLine(line)
				break
			}
			lineDepth := depth(line)
			currentDepth += lineDepth

			if currentDepth < 0 {
				lines = lines[:0]
				currentDepth = 0
				break
			}

			if len(lines) == 0 && lineDepth == 0 {
				s <- line + "\n"
				break
			}

			if err := addLine(line); err != nil {
				errorMode = true
				lines = lines[:0]
				currentDepth = 0
				break
			}

			if currentDepth == 0 {
				program := strings.Join(append(lines, "\n"), " ")
				lines = lines[:0]
				s <- program
			}
		}
	}
}

func runTCPServer() {
	ln, err := net.Listen("tcp", *tcpAddr)
	if err != nil {
		die(err.Error())
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			printErrorMsg(err.Error())
			continue
		}
		go startConnection(conn, manager)
	}
}

func startConnection(conn net.Conn, m *cm.Manager) {
	m.Add(conn)
	defer m.Remove(conn)

	out, errors := make(chan interface{}), make(chan error)
	userinput := make(chan string)

	go tcpInterpreter(conn, userinput, out, errors)

	r := bufio.NewReader(conn)
	for {
		line, _, err := r.ReadLine()
		switch err {
		case nil:
			break
		case io.EOF:
			io.WriteString(conn, "<eof>")
			return
		default:
			printErrorMsg(err.Error())
			return
		}
		userinput <- string(line)
	}
}
