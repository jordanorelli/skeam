package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

var tcpAddr = flag.String("tcp", "", "foo")

func args() {
	flag.Parse()
	if *tcpAddr != "" {
		runTCPServer()
		return
	}
	filename := flag.Args()[1]
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to read file ", filename)
		os.Exit(1)
	}
	defer f.Close()

	out, errors := make(chan interface{}), make(chan error)
	go defaultInterpreter(out, errors)

	c := make(chan token, 32)
	go lex(bufio.NewReader(f), c)
	evalall(c, out, errors, universe)
}

func printErrorMsg(message string) {
	io.WriteString(os.Stderr, message)
}

func die(message string) {
	printErrorMsg(message)
	os.Exit(2)
}

func tcpInterpreter(conn net.Conn, out chan interface{}, errors chan error) {
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
		}
	}
}

func startConnection(conn net.Conn, c, d chan net.Conn) {
	c <- conn
	defer func() {
		d <- conn
		fmt.Println("disconnected")
	}()

	out, errors := make(chan interface{}), make(chan error)
	go tcpInterpreter(conn, out, errors)

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

		tokens := make(chan token, 32)
		go lexs(string(line)+"\n", tokens)
		evalall(tokens, out, errors, universe)
	}
}

var activeConnections = make([]net.Conn, 0, 10)

func removeConnection(conn net.Conn) {
	for i, other := range activeConnections {
		if conn.RemoteAddr() == other.RemoteAddr() {
			activeConnections = append(activeConnections[:i], activeConnections[i+1:]...)
			return
		}
	}
}

func manageConnections(connect, disconnect chan net.Conn) {
	for {
		select {
		case conn := <-connect:
			activeConnections = append(activeConnections, conn)
		case conn := <-disconnect:
			removeConnection(conn)
		}
	}
}

func runTCPServer() {
	connect, disconnect := make(chan net.Conn), make(chan net.Conn)
	go manageConnections(connect, disconnect)
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
		go startConnection(conn, connect, disconnect)
	}
}
