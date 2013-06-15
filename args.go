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

	c := make(chan token, 32)
	go lex(bufio.NewReader(f), c)
	evalall(c, universe)
}

func printErrorMsg(message string) {
	io.WriteString(os.Stderr, message)
}

func die(message string) {
	printErrorMsg(message)
	os.Exit(2)
}

func startConnection(conn net.Conn, c, d chan net.Conn) {
	c <- conn
	defer func() { d <- conn }()
	disconnect := func() {
		fmt.Println("disconnected")
	}
	r := bufio.NewReader(conn)
	for {
		if _, err := io.WriteString(conn, "> "); err != nil {
			disconnect()
			return
		}
		line, prefix, err := r.ReadLine()
		if prefix {
			fmt.Println("(prefix)")
		}
		switch err {
		case nil:
			break
		case io.EOF:
			disconnect()
			return
		default:
			printErrorMsg(err.Error())
			disconnect()
			return
		}

		tokens := make(chan token, 32)
		go lexs(string(line)+"\n", tokens)
		evalall(tokens, universe)
	}
}

func manageConnections(connect, disconnect chan net.Conn) {
	for {
		select {
		case <-connect:

		case <-disconnect:
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
