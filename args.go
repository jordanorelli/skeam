package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

var (
	tcpAddr  = flag.String("tcp", "", "tcp ip:port to listen on")
	httpAddr = flag.String("http", "", "http ip:port to listen on")
)

// executes a file on disk using the universe environment.  This will block
// until the entire file has been executed.  Vals and errors printed to stdout
// and stderr, respectively.
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
