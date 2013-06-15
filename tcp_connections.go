package main

import (
	"./cm"
	"bufio"
	"fmt"
	"io"
	"net"
)

var manager = cm.New()

func tcpInterpreter(conn net.Conn, userinput chan string, out chan interface{}, errors chan error) {
	prompt := func() {
		io.WriteString(conn, "> ")
	}
	prompt()
	for {
		select {
		case v := <-out:
			fmt.Fprintln(conn, v)
			prompt()
		case err := <-errors:
			fmt.Fprintf(conn, "error: %v", err)
		case line := <-userinput:
			tokens := make(chan token, 32)
			go lexs(line+"\n", tokens)
			go evalall(tokens, out, errors, universe)
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
