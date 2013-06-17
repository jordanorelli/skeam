package main

import (
	"./cm"
	"bufio"
	"errors"
	"fmt"
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
		lines = append(lines, line)
		return nil
	}
	errorMode := false
	skipLine := func(line string) {
		currentDepth += depth(line)
		if currentDepth == 0 {
			errorMode = false
		}
	}
	for {
		select {
		case v := <-out:
			fmt.Fprintln(manager, v)
		case err := <-errors:
			fmt.Fprintf(conn, "error: %v\n", err)
		case line := <-userinput:
			if errorMode {
				skipLine(line)
				break
			}
			lineDepth := depth(line)
			currentDepth += lineDepth

			if len(lines) == 0 && lineDepth == 0 {
				tokens := make(chan token, 32)
				go lexs(line+"\n", tokens)
				go evalall(tokens, out, errors, universe)
				break
			}

			if err := addLine(line); err != nil {
				errorMode = true
				lines = lines[:0]
				currentDepth = 0
				break
			}

			if currentDepth == 0 {
				runnable := strings.Join(lines, " ")
				lines = lines[:0]
				tokens := make(chan token, 32)
				go lexs(runnable+"\n", tokens)
				go evalall(tokens, out, errors, universe)
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
		line, prefix, err := r.ReadLine()
		if prefix {
			fmt.Println("(prefix)")
		}
		switch err {
		case nil:
			break
		case io.EOF:
			return
		default:
			printErrorMsg(err.Error())
			return
		}
		userinput <- string(line)
	}
}
