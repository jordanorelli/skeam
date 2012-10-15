(+ 1 (+ 1 1) (dave (sam 1)))

; alright I have comments now!!!!
; woooohooo!!!!
; nice nice nice

(+ 1
   22.3
   a
   ;here comes a comment
   3.
   (one two three)
   "this is a string"
   4.0
   ((one two) three)
   abøne
   (dave
     1
     "here's an escaped quote: \" how neat!!!"
     2
     "and here;'s an escaped \\, sweet!"
     albert-camus
     3
     (sam 3 2 2)))



(begin (set! x 1)
       (set! x (+ x 1))
       (* x 2))

; ------------------------------------------------------------------------------
; the following stuff comes directly from the norvig essay, instead of being
; contrived lexer tests.
; ------------------------------------------------------------------------------

; define a function and then execute it
(begin (define r 3) (* 3.141592653 (* r r)))

; same thing, alternative form without "begin"
(define area (lambda (r) (* 3.141592653 (* r r))))
(area 3)
