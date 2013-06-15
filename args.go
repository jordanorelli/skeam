package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

var tcpAddr = flag.String("tcp", "", "foo")

func runfile() {
	filename := flag.Args()[0]
	fmt.Println(filename)
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
