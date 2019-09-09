# Knuth's Elevator Simulator

### About

This is a [Go](https://golang.org/) implementation of the elevator simulator described by [Donald E. Knuth](https://en.wikipedia.org/wiki/Donald_Knuth) in *[The Art of Computer Programming](https://en.wikipedia.org/wiki/The_Art_of_Computer_Programming)* (*TAOCP*) Volume 1.  While reading that tome, I found his example application of [doubly linked lists](https://en.wikipedia.org/wiki/Doubly_linked_list) so long and ridiculous that it warranted further exploration. 

Knuth’s example is a [discrete-event simulation](https://en.wikipedia.org/wiki/Discrete-event_simulation) involving concurrently executing entities that interact with each other.  It demonstrates how a sorted pending-event queue and a single thread (in the case of Go, one goroutine) can imitate parallel processing.   

While that concept could have been explained with a basic, contrived system, Knuth elaborately details the elevator system in the Mathematics building of the [California Institute of Technology](https://en.wikipedia.org/wiki/California_Institute_of_Technology) across 15 pages (ignoring the exercises that follow).  And, as is the convention of *TAOCP*, the lengthy algorithm is conveyed through headache-inducing blocks of text and commented assembly language instead of pseudocode or a high-level programming language.

### Example Output

```
TIME    STATE   FLOOR   D1      D2      D3      step    action
0000    N       2       0       0       0       U1      User 1 arrives at floor 2, destination is 0.
0000    N       2       0       0       0       U2      User 1 presses down button.
0000    N       2       0       0       0       U3      User 1 stands in queue in front of elevator.
0020    N       2       0       0       0       E3      Elevator doors start to open.
0040    N       2       X       X       0       E4      Doors are open. Users about to enter.
0040    N       2       X       X       0       U5      User 1 gets in.
0065    D       2       X       X       0       E4      Doors are open. Nobody outside elevator.
0065    D       2       0       X       X       E5      Elevator doors start to close.
0079    D       2       0       X       0       U1      User 2 arrives at floor 2, destination is 3.
0079    D       2       0       X       0       U2      User 2 arrives at doors closing and stop them.
0079    D       2       0       X       0       U3      User 2 stands in queue in front of elevator.
0079    D       2       0       X       0       E3      Elevator doors start to open.
0099    D       2       X       X       0       E4      Doors are open. Users about to enter.
0099    D       2       X       X       0       U5      User 2 gets in.
0124    D       2       X       X       0       E4      Doors are open. Nobody outside elevator.
0155    D       2       0       X       X       E5      Elevator doors start to close.
0175    D       2       0       X       0       E6      Elevator about to go down
0190    D       2       0       X       0       E8      Elevator moving down
0251    D       1       0       X       0       E8      Elevator moving down
0335    D       0       0       X       0       E2      Elevator stops.
0335    U       0       0       X       0       E3      Elevator doors start to open.
0355    U       0       X       X       0       E4      Doors are open. Users about to exit.
0355    U       0       X       X       0       U6      User 1 gets out, leaves the system.
0380    U       0       X       X       0       E4      Doors are open. Nobody outside elevator.
0411    U       0       0       X       X       E5      Elevator doors start to close.
0431    U       0       0       X       0       E6      Elevator about to go up
0446    U       0       0       X       0       E7      Elevator moving up
0497    U       1       0       X       0       E7      Elevator moving up
0548    U       2       0       X       0       E7      Elevator moving up
0569    U       3       0       X       0       U1      User 3 arrives at floor 1, destination is 2.
0569    U       3       0       X       0       U2      User 3 presses up button.
0569    U       3       0       X       0       U3      User 3 stands in queue in front of elevator.
0613    U       3       0       X       0       E2      Elevator stops.
0613    D       3       0       X       0       E3      Elevator doors start to open.
0633    D       3       X       X       0       E4      Doors are open. Users about to exit.
0633    D       3       X       X       0       U6      User 2 gets out, leaves the system.
0658    D       3       X       X       0       E4      Doors are open. Nobody outside elevator.
0689    D       3       0       X       X       E5      Elevator doors start to close.
0709    D       3       0       X       0       E6      Elevator about to go down
0724    D       3       0       X       0       E8      Elevator moving down
0785    D       2       0       X       0       E8      Elevator moving down
0869    D       1       0       X       0       E2      Elevator stops.
0869    N       1       0       X       0       E3      Elevator doors start to open.
0889    N       1       X       X       0       E4      Doors are open. Users about to enter.
0889    N       1       X       X       0       U5      User 3 gets in.
0914    U       1       X       X       0       E4      Doors are open. Nobody outside elevator.
0914    U       1       0       X       X       E5      Elevator doors start to close.
0934    U       1       0       X       0       E6      Elevator about to go up
0949    U       1       0       X       0       E7      Elevator moving up
0960    U       2       0       X       0       U1      User 4 arrives at floor 1, destination is 0.
0960    U       2       0       X       0       U2      User 4 presses down button.
0960    U       2       0       X       0       U3      User 4 stands in queue in front of elevator.
1014    U       2       0       X       0       E2      Elevator stops.
1014    D       2       0       X       0       E3      Elevator doors start to open.
1034    D       2       X       X       0       E4      Doors are open. Users about to exit.
1034    D       2       X       X       0       U6      User 3 gets out, leaves the system.
1059    D       2       X       X       0       E4      Doors are open. Nobody outside elevator.
1090    D       2       0       X       X       E5      Elevator doors start to close.
1110    D       2       0       X       0       E6      Elevator about to go down
1125    D       2       0       X       0       E8      Elevator moving down
1209    D       1       0       X       0       E2      Elevator stops.
1209    N       1       0       X       0       E3      Elevator doors start to open.
...
7447    N       2       0       X       0       E1      Elevator dormant
7651    N       2       0       X       0       E9      Elevator not active
7899    N       2       0       0       0       U1      User 18 arrives at floor 1, destination is 4.
7899    N       2       0       0       0       U2      User 18 presses up button.
7899    D       2       0       0       0       U3      User 18 stands in queue in front of elevator.
7919    D       2       0       0       0       E6      Elevator about to go down
7934    D       2       0       0       0       E8      Elevator moving down
8018    D       1       0       0       0       E2      Elevator stops.
8018    N       1       0       0       0       E3      Elevator doors start to open.
8038    N       1       X       X       0       E4      Doors are open. Users about to enter.
8038    N       1       X       X       0       U5      User 18 gets in.
8063    U       1       X       X       0       E4      Doors are open. Nobody outside elevator.
8063    U       1       0       X       X       E5      Elevator doors start to close.
8083    U       1       0       X       0       E6      Elevator about to go up
8098    U       1       0       X       0       E7      Elevator moving up
8149    U       2       0       X       0       E7      Elevator moving up
8200    U       3       0       X       0       E7      Elevator moving up
8265    U       4       0       X       0       E2      Elevator stops.
8265    N       4       0       X       0       E3      Elevator doors start to open.
8285    N       4       X       X       0       E4      Doors are open. Users about to exit.
8285    N       4       X       X       0       U6      User 18 gets out, leaves the system.
8310    N       4       X       X       0       E4      Doors are open. Nobody outside elevator.
8341    N       4       0       X       X       E5      Elevator doors start to close.
8361    D       4       0       X       0       E6      Elevator about to go down
8376    D       4       0       X       0       E8      Elevator moving down
8437    D       3       0       X       0       E8      Elevator moving down
8521    D       2       0       X       0       E2      Elevator stops.
8521    N       2       0       X       0       E3      Elevator doors start to open.
8523    N       2       X       X       0       U1      User 19 arrives at floor 4, destination is 1.
8523    N       2       X       X       0       U2      User 19 presses down button.
8523    N       2       X       X       0       U3      User 19 stands in queue in front of elevator.
8541    N       2       X       X       0       E4      Doors are open. Nobody outside elevator.
8597    N       2       0       X       X       E5      Elevator doors start to close.
8617    U       2       0       X       0       E6      Elevator about to go up
8632    U       2       0       X       0       E7      Elevator moving up
8683    U       3       0       X       0       E7      Elevator moving up
8748    U       4       0       X       0       E2      Elevator stops.
8748    N       4       0       X       0       E3      Elevator doors start to open.
8757    N       4       X       X       0       U1      User 20 arrives at floor 4, destination is 1.
8757    N       4       X       X       0       U2      User 20 presses down button.
8757    N       4       X       X       0       U3      User 20 stands in queue in front of elevator.
8768    N       4       X       X       0       E4      Doors are open. Users about to enter.
8768    N       4       X       X       0       U5      User 19 gets in.
8793    D       4       X       X       0       E4      Doors are open. Users about to enter.
8793    D       4       X       X       0       U5      User 20 gets in.
8793    D       4       X       X       0       E5      Doors flutter.
8818    D       4       X       X       0       E4      Doors are open. Nobody outside elevator.
8833    D       4       0       X       X       E5      Elevator doors start to close.
8853    D       4       0       X       0       E6      Elevator about to go down
8868    D       4       0       X       0       E8      Elevator moving down
8929    D       3       0       X       0       E8      Elevator moving down
8990    D       2       0       X       0       E8      Elevator moving down
9074    D       1       0       X       0       E2      Elevator stops.
9074    N       1       0       X       0       E3      Elevator doors start to open.
9094    N       1       X       X       0       E4      Doors are open. Users about to exit.
9094    N       1       X       X       0       U6      User 20 gets out, leaves the system.
9119    N       1       X       X       0       E4      Doors are open. Users about to exit.
9119    N       1       X       X       0       U6      User 19 gets out, leaves the system.
9144    N       1       X       X       0       E4      Doors are open. Nobody outside elevator.
9150    N       1       0       X       X       E5      Elevator doors start to close.
9170    U       1       0       X       0       E6      Elevator about to go up
9185    U       1       0       X       0       E7      Elevator moving up
9250    U       2       0       X       0       E2      Elevator stops.
9250    N       2       0       X       0       E3      Elevator doors start to open.
9270    N       2       X       X       0       E4      Doors are open. Nobody outside elevator.
9326    N       2       0       X       X       E5      Elevator doors start to close.
9346    N       2       0       X       0       E6      Elevator about to go dormant
9346    N       2       0       X       0       E1      Elevator dormant
9495    N       2       0       X       0       U1      User 21 arrives at floor 0, destination is 4.
9495    N       2       0       X       0       U2      User 21 presses up button.
9495    D       2       0       X       0       U3      User 21 stands in queue in front of elevator.
9515    D       2       0       X       0       E6      Elevator about to go down
9530    D       2       0       X       0       E8      Elevator moving down
9561    D       1       0       X       0       U1      User 22 arrives at floor 1, destination is 0.
9561    D       1       0       X       0       U2      User 22 presses down button.
9561    D       1       0       X       0       U3      User 22 stands in queue in front of elevator.
9614    D       1       0       X       0       E2      Elevator stops.
9614    D       1       0       X       0       E3      Elevator doors start to open.
9634    D       1       X       X       0       E4      Doors are open. Users about to enter.
9634    D       1       X       X       0       U5      User 22 gets in.
9659    D       1       X       X       0       E4      Doors are open. Nobody outside elevator.
9690    D       1       0       X       X       E5      Elevator doors start to close.
9710    D       1       0       X       0       E6      Elevator about to go down
9725    D       1       0       X       0       E8      Elevator moving down
9809    D       0       0       X       0       E2      Elevator stops.
9809    N       0       0       X       0       E3      Elevator doors start to open.
9829    N       0       X       X       0       E4      Doors are open. Users about to exit.
9829    N       0       X       X       0       U6      User 22 gets out, leaves the system.
9854    N       0       X       X       0       E4      Doors are open. Users about to enter.
9854    N       0       X       X       0       U5      User 21 gets in.
9879    U       0       X       X       0       E4      Doors are open. Nobody outside elevator.
9879    U       0       0       X       X       E5      Elevator doors start to close.
9899    U       0       0       X       0       E6      Elevator about to go up
9914    U       0       0       X       0       E7      Elevator moving up
9965    U       1       0       X       0       E7      Elevator moving up
```

### Bonus Fact

One of the references on [Wikipedia’s Elevator Paradox page](https://en.wikipedia.org/wiki/Elevator_paradox) is [a book by Donald E. Knuth](https://www.amazon.com/Selected-Papers-Games-Lecture-Notes/dp/157586584X).