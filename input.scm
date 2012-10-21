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
