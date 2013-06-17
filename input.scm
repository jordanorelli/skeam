; ------------------------------------------------------------------------------
; literals
; ------------------------------------------------------------------------------

; integer
1

; float
3.14

; string
"jordan"

; booleans.  I don't like the look of #t and #f.  They're dumb.
#t
#f

; ------------------------------------------------------------------------------
; basic math
; ------------------------------------------------------------------------------

(+ 1 1)
(- 1 1)
(* 1 1)
(/ 1 1)

; ------------------------------------------------------------------------------
; grammar
; ------------------------------------------------------------------------------

(define x 5)
x
(set! x "steve")
x

(quote (1 2 3))

(if #f (quote "true-value") (quote "false-value"))
(if #t (quote "true-value") (quote "false-value"))

(define plusone (lambda (x) (+ x 1)))
(plusone 1)

((lambda (x) (* x x)) 4)

((lambda (x y) (+ x y)) 10 25)

(define make-account
  (lambda (balance)
    (lambda (amt)
      (begin (set! balance (+ balance amt)) balance))))

(define a1 (make-account 100.00))
(a1 -20.00)

(not "dave")
(if (not #f) (quote "true-condition") (quote "false-condition"))

(length (quote (1 2 3)))

(list 1 2 3)
(length (list 1 2 3))

(null? null)

; this one I don't get.  It's supposed to evaluate to false?  What?  Why?
; Because null is an expression that evaluates to an empty sexp, but it's not,
; itself, actually null?  What the fuck, lisp?
(null? (quote null))

(null? (quote ()))
(null? (list))

(symbol? (quote null))
(symbol? 1)

; ahahhahahah this is so ugly.
(define counter
  ((lambda ()
    (begin
      (define count 0)
      (lambda ()
        (begin
          (set! count (+ 1 count))
          count))))))

; hmm, some kind of looping construct would be nice.
(counter)
(counter)
(counter)
(counter)
(counter)
(counter)
(counter)
(counter)
(counter)
(counter)

; ------------------------------------------------------------------------------
; norving examples
; ------------------------------------------------------------------------------

(define area (lambda (r) (* 3.141592653 (* r r))))
(area 3)

; <= isn't defined yet
(define fact (lambda (n) (if (<= n 1) 1 (* n (fact (- n 1))))))
(fact 10)

; this one is an overflow error :(
; (fact 100)

(area (fact 10))

; (define first car)
; (define rest cdr)
; (define count
;   (lambda (item L)
;     (if L (+ (equal? item (first L)) (count item (rest L)))
;       0)))
; (count 0 (list 0 1 2 3 0 0))
; (count (quote the) (quote (the more the merrier the bigger the better)))
