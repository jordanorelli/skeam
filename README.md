# skeam

Skeam is a primitive
[Lisp](http://en.wikipedia.org/wiki/Lisp_(programming_language\)) interpreter.
I wrote this out of a curiosity to learn about the basics of writing
interpreters; it's not something that I'd recommend using, but it may be
helpful to look at if you're interested in writing your own.  The name comes
from [Scheme](http://en.wikipedia.org/wiki/Scheme_(programming_language\)) and
[Skream](http://en.wikipedia.org/wiki/Skream).

Skeam does not implement [tail-call](http://en.wikipedia.org/wiki/Tail_call)
elimination or [continuations](http://en.wikipedia.org/wiki/Continuation), so
it's not technically a Scheme implementation.

The `input.scm` file gives an example of what is currently understood by the interpreter.

## installing skeam

First make sure you have Go1, the current version of the Go programming
language.  If you don't have it, you can download it
[here](http://golang.org/doc/install).

Skeam is go-gettable, so installation only requires the following command:  `go
install github.com/jordanorelli/skeam`.  Make sure your `$GOBIN` is included in
your environment's `$PATH`.  E.g., on Mac OS X, this generally means adding
`export PATH=$PATH:/usr/local/go/bin` to your `.bashrc`.

Once installed, you can access the Skeam REPL by simply running the command
`skeam`.  To execute a Skeam file, pass the filename as a parameter to the
`skeam` command.  E.g., `skeam input.scm` would run the `input.scm` file.
