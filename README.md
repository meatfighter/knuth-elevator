# Knuth's Elevator Simulator

### About

This is a [Go](https://golang.org/) implementation of the elevator simulator described by [Donald E. Knuth](https://en.wikipedia.org/wiki/Donald_Knuth) in *[The Art of Computer Programming](https://en.wikipedia.org/wiki/The_Art_of_Computer_Programming)* (*TAOCP*) Volume 1.  While reading that tome, I found his example application of [doubly linked lists](https://en.wikipedia.org/wiki/Doubly_linked_list) so long and ridiculous that it deserved further exploration. 

Knuth’s example is a [discrete-event simulation](https://en.wikipedia.org/wiki/Discrete-event_simulation) involving concurrently executing entities that interact with each other.  It demonstrates how a sorted pending-event queue and a single thread (in the case of Go, one goroutine) can imitate parallel processing.   

While that concept could have been explained with a basic, contrived system, Knuth elaborately details the elevator system in the Mathematics building of the [California Institute of Technology](https://en.wikipedia.org/wiki/California_Institute_of_Technology) across 15 pages (ignoring the exercises that follow).  And, as is the convention of *TAOCP*, the lengthy algorithm is conveyed through headache-inducing blocks of text and commented assembly language instead of pseudocode or a high-level programming language.

### Example Output

```
TIME    STATE   FLOOR   D1      D2      D3      step    action
0000    N       2       0       0       0       U1      User 1 arrives at floor 1, destination is 3.
0000    N       2       0       0       0       U2      User 1 presses up button.
0000    D       2       0       0       0       U3      User 1 stands in queue in front of elevator.
0020    D       2       0       0       0       E6      Elevator about to go down
0026    D       2       0       0       0       U1      User 2 arrives at floor 3, destination is 1.
0026    D       2       0       0       0       U2      User 2 presses down button.
0026    D       2       0       0       0       U3      User 2 stands in queue in front of elevator.
0035    D       2       0       0       0       E8      Elevator moving down
0119    D       1       0       0       0       E2      Elevator stops.
0119    U       1       0       0       0       E3      Elevator doors start to open.
0139    U       1       X       X       0       E4      Doors are open. Users about to enter.
0139    U       1       X       X       0       U5      User 1 gets in.
0164    U       1       X       X       0       E4      Doors are open. Nobody outside elevator.
0195    U       1       0       X       X       E5      Elevator doors start to close.
0215    U       1       0       X       0       E6      Elevator about to go up
0230    U       1       0       X       0       E7      Elevator moving up
0281    U       2       0       X       0       E7      Elevator moving up
0299    U       3       0       X       0       U1      User 3 arrives at floor 4, destination is 2.
0299    U       3       0       X       0       U2      User 3 presses down button.
0299    U       3       0       X       0       U3      User 3 stands in queue in front of elevator.
0346    U       3       0       X       0       E2      Elevator stops.
0346    U       3       0       X       0       E3      Elevator doors start to open.
0366    U       3       X       X       0       E4      Doors are open. Users about to exit.
0366    U       3       X       X       0       U6      User 1 gets out, leaves the system.
0391    U       3       X       X       0       E4      Doors are open. Users about to enter.
0391    U       3       X       X       0       U5      User 2 gets in.
0416    U       3       X       X       0       E4      Doors are open. Nobody outside elevator.
0419    U       3       0       X       X       U1      User 4 arrives at floor 2, destination is 3.
0419    U       3       0       X       X       U2      User 4 presses up button.
0419    U       3       0       X       X       U3      User 4 stands in queue in front of elevator.
0422    U       3       0       X       X       E5      Elevator doors start to close.
0442    U       3       0       X       0       E6      Elevator about to go up
0457    U       3       0       X       0       E7      Elevator moving up
0522    U       4       0       X       0       E2      Elevator stops.
0522    D       4       0       X       0       E3      Elevator doors start to open.
0542    D       4       X       X       0       E4      Doors are open. Users about to enter.
0542    D       4       X       X       0       U5      User 3 gets in.
0567    D       4       X       X       0       E4      Doors are open. Nobody outside elevator.
0598    D       4       0       X       X       E5      Elevator doors start to close.
0618    D       4       0       X       0       E6      Elevator about to go down
0633    D       4       0       X       0       E8      Elevator moving down
0717    D       3       0       X       0       E2      Elevator stops.
0717    D       3       0       X       0       E3      Elevator doors start to open.
0737    D       3       X       X       0       E4      Doors are open. Nobody outside elevator.
0793    D       3       0       X       X       E5      Elevator doors start to close.
0813    D       3       0       X       0       E6      Elevator about to go down
0828    D       3       0       X       0       E8      Elevator moving down
0912    D       2       0       X       0       E2      Elevator stops.
0912    D       2       0       X       0       E3      Elevator doors start to open.
0932    D       2       X       X       0       E4      Doors are open. Users about to exit.
0932    D       2       X       X       0       U6      User 3 gets out, leaves the system.
0957    D       2       X       X       0       E4      Doors are open. Users about to enter.
0957    D       2       X       X       0       U5      User 4 gets in.
0982    D       2       X       X       0       E4      Doors are open. Nobody outside elevator.
0988    D       2       0       X       X       E5      Elevator doors start to close.
1008    D       2       0       X       0       E6      Elevator about to go down
1023    D       2       0       X       0       E8      Elevator moving down
1107    D       1       0       X       0       E2      Elevator stops.
1107    U       1       0       X       0       E3      Elevator doors start to open.
1127    U       1       X       X       0       E4      Doors are open. Users about to exit.
1127    U       1       X       X       0       U6      User 2 gets out, leaves the system.
1152    U       1       X       X       0       E4      Doors are open. Nobody outside elevator.
1183    U       1       0       X       X       E5      Elevator doors start to close.
1203    U       1       0       X       0       E6      Elevator about to go up
...
7865    N       2       0       X       0       E1      Elevator dormant
8069    N       2       0       X       0       E9      Elevator not active
8069    N       2       0       0       0       E1      Elevator dormant
8307    N       2       0       0       0       U1      User 19 arrives at floor 4, destination is 0.
8307    N       2       0       0       0       U2      User 19 presses down button.
8307    U       2       0       0       0       U3      User 19 stands in queue in front of elevator.
8327    U       2       0       0       0       E6      Elevator about to go up
8342    U       2       0       0       0       E7      Elevator moving up
8393    U       3       0       0       0       E7      Elevator moving up
8458    U       4       0       0       0       E2      Elevator stops.
8458    N       4       0       0       0       E3      Elevator doors start to open.
8478    N       4       X       X       0       E4      Doors are open. Users about to enter.
8478    N       4       X       X       0       U5      User 19 gets in.
8503    D       4       X       X       0       E4      Doors are open. Nobody outside elevator.
8503    D       4       0       X       X       E5      Elevator doors start to close.
8523    D       4       0       X       0       E6      Elevator about to go down
8538    D       4       0       X       0       E8      Elevator moving down
8599    D       3       0       X       0       E8      Elevator moving down
8622    D       2       0       X       0       U1      User 20 arrives at floor 0, destination is 1.
8622    D       2       0       X       0       U2      User 20 presses up button.
8622    D       2       0       X       0       U3      User 20 stands in queue in front of elevator.
8660    D       2       0       X       0       E8      Elevator moving down
8721    D       1       0       X       0       E8      Elevator moving down
8805    D       0       0       X       0       E2      Elevator stops.
8805    N       0       0       X       0       E3      Elevator doors start to open.
8822    N       0       X       X       0       U1      User 21 arrives at floor 0, destination is 2.
8822    N       0       X       X       0       U2      User 21 presses up button.
8822    N       0       X       X       0       U3      User 21 stands in queue in front of elevator.
8825    N       0       X       X       0       E4      Doors are open. Users about to exit.
8825    N       0       X       X       0       U6      User 19 gets out, leaves the system.
8850    N       0       X       X       0       E4      Doors are open. Users about to enter.
8850    N       0       X       X       0       U5      User 20 gets in.
8875    U       0       X       X       0       E4      Doors are open. Users about to enter.
8875    U       0       X       X       0       U5      User 21 gets in.
8875    U       0       X       X       0       E5      Doors flutter.
8900    U       0       X       X       0       E4      Doors are open. Nobody outside elevator.
8915    U       0       0       X       X       E5      Elevator doors start to close.
8935    U       0       0       X       0       E6      Elevator about to go up
8950    U       0       0       X       0       E7      Elevator moving up
9015    U       1       0       X       0       E2      Elevator stops.
9015    U       1       0       X       0       E3      Elevator doors start to open.
9035    U       1       X       X       0       E4      Doors are open. Users about to exit.
9035    U       1       X       X       0       U6      User 20 gets out, leaves the system.
9060    U       1       X       X       0       E4      Doors are open. Nobody outside elevator.
9091    U       1       0       X       X       E5      Elevator doors start to close.
9111    U       1       0       X       0       E6      Elevator about to go up
9126    U       1       0       X       0       E7      Elevator moving up
9191    U       2       0       X       0       E2      Elevator stops.
9191    N       2       0       X       0       E3      Elevator doors start to open.
9211    N       2       X       X       0       E4      Doors are open. Users about to exit.
9211    N       2       X       X       0       U6      User 21 gets out, leaves the system.
9236    N       2       X       X       0       E4      Doors are open. Nobody outside elevator.
9267    N       2       0       X       X       E5      Elevator doors start to close.
9287    N       2       0       X       0       E6      Elevator about to go dormant
9287    N       2       0       X       0       E1      Elevator dormant
9376    N       2       0       X       0       U1      User 22 arrives at floor 0, destination is 3.
9376    N       2       0       X       0       U2      User 22 presses up button.
9376    D       2       0       X       0       U3      User 22 stands in queue in front of elevator.
9396    D       2       0       X       0       E6      Elevator about to go down
9411    D       2       0       X       0       E8      Elevator moving down
9472    D       1       0       X       0       E8      Elevator moving down
9556    D       0       0       X       0       E2      Elevator stops.
9556    N       0       0       X       0       E3      Elevator doors start to open.
9576    N       0       X       X       0       E4      Doors are open. Users about to enter.
9576    N       0       X       X       0       U5      User 22 gets in.
9601    U       0       X       X       0       E4      Doors are open. Nobody outside elevator.
9601    U       0       0       X       X       E5      Elevator doors start to close.
9621    U       0       0       X       0       E6      Elevator about to go up
9636    U       0       0       X       0       E7      Elevator moving up
9687    U       1       0       X       0       E7      Elevator moving up
9738    U       2       0       X       0       E7      Elevator moving up
9803    U       3       0       X       0       E2      Elevator stops.
9803    N       3       0       X       0       E3      Elevator doors start to open.
9823    N       3       X       X       0       E4      Doors are open. Users about to exit.
9823    N       3       X       X       0       U6      User 22 gets out, leaves the system.
9848    N       3       X       X       0       E4      Doors are open. Nobody outside elevator.
9879    N       3       0       X       X       E5      Elevator doors start to close.
9899    D       3       0       X       0       E6      Elevator about to go down
9914    D       3       0       X       0       E8      Elevator moving down
9998    D       2       0       X       0       E2      Elevator stops.
9998    N       2       0       X       0       E3      Elevator doors start to open.
```

### Bonus Fact

One of the references on [Wikipedia’s Elevator Paradox page](https://en.wikipedia.org/wiki/Elevator_paradox) is [a book by Donald E. Knuth](https://www.amazon.com/Selected-Papers-Games-Lecture-Notes/dp/157586584X).