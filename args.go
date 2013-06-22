package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var (
	tcpAddr = flag.String("tcp", "", "tcp ip:port to listen on")
)

func runfile() {
	filename := flag.Args()[0]
	f, err := os.Open(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to read file ", filename)
		os.Exit(1)
	}
	defer f.Close()

	i := newInterpreter(f, os.Stdout, os.Stderr)
	i.run(universe)
}

func printErrorMsg(message string) {
	io.WriteString(os.Stderr, message)
}

func die(message string) {
	printErrorMsg(message)
	os.Exit(2)
}
