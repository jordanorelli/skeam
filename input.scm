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
