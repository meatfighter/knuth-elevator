# knuth-elevator

### About

This is a [Go](https://golang.org/) implementation of the elevator simulator described by Donald E. Knuth in The Art of Computer Programming (TAOCP) Volume 1.  While reading that tome, I found his example application of doubly linked lists so long and ridiculous that it deserved further exploration. 

Knuth’s example is a discrete-event simulation involving concurrently executing entities that interact with each other.  It demonstrates how a sorted pending-event queue and a single thread (in the case of Go, one goroutine) can imitate parallel processing.   

While that concept could have been explained with a basic, contrived system, Knuth elaborately details the elevator system in the Mathematics building of the California Institute of Technology across 15 pages (ignoring the exercises that follow).  And, as is the convention of TAOCP, the lengthy algorithm is conveyed through headache-inducing blocks of text and commented assembly language instead of pseudocode or a high-level programming language.