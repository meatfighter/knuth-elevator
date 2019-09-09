// Go implementation of the simulator described in:

// The Art of Computer Programming, Volume 1: Fundamental Algorithms, Third Edition (Donald E. Knuth)
// Chapter 2 Information Structures
// 2.2.5 Doubly Linked Lists
// Pages 280–298

// As an example of the use of doubly linked lists, we will now consider the
// writing of a discrete simulation program. “Discrete simulation” means the
// simulation of a system in which all changes in the state of the system may
// be assumed to happen at certain discrete instants of time. The “system” being
// simulated is usually a set of individual activities that are largely independent
// although they interact with each other; examples are customers at a store, ships
// in a harbor, people in a corporation. In a discrete simulation, we proceed by
// doing whatever is to be done at a certain instant of simulated time, then advance
// the simulated clock to the next time when some action is scheduled to occur.

// By contrast, a “continuous simulation” would be simulation of activities that
// are under continuous changes, such as traffic moving on a highway, spaceships
// traveling to other planets, etc. Continuous simulation can often be satisfactorily
// approximated by discrete simulation with very small time intervals between
// steps; however, in such a case we usually have “synchronous” discrete simulation,
// in which many parts of the system are slightly altered at each discrete time
// interval, and such an application generally calls for a somewhat different type of
// program organization than the kind considered here.

// The program developed below simulates the elevator system in the Mathematics
// building of the California Institute of Technology. The results of such a
// simulation will perhaps be of use only to people who make reasonably frequent
// visits to Caltech; and even for them, it may be simpler just to try using the
// elevator several times instead of writing a computer program. But, as is usual
// with simulation studies, the methods we will use are of much more interest than
// the answers given by the program. The methods to be discussed below illustrate
// typical implementation techniques used with discrete simulation programs.

package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	// The Mathematics building has five floors: sub-basement, basement, first,
	// second, and third. There is a single elevator, which has automatic controls
	// and can stop at each floor. For convenience we will renumber the floors 0, 1,
	// 2, 3, and 4.
	floorSubbasement = 0
	floorBasement    = 1
	floorFirst       = 2
	floorSecond      = 3
	floorThird       = 4

	floorHome = floorFirst

	floors = 5

	// The elevator is in one of three states: GOINGUP, GOINGDOWN, or NEUTRAL.
	// (The current state is indicated to passengers by lighted arrows inside the
	// elevator.) If it is in NEUTRAL state and not on floor 2, the machine will close
	// its doors and (if no command is given by the time its doors are shut) it will
	// change to GOINGUP or GOINGDOWN, heading for floor 2. (This is the “home floor,”
	// since most passengers get in there.) On floor 2 in NEUTRAL state, the doors will
	// eventually close and the machine will wait silently for another command. The first
	// command received for another floor sets the machine GOINGUP or GOINGDOWN as
	// appropriate; it stays in this state until there are no commands waiting in the
	// same direction, and then it switches direction or switches to NEUTRAL just before
	// opening the doors, depending on what other commands are in the CALL variables. The
	// elevator takes a certain amount of time to open and close its doors, to accelerate
	// and decelerate, and to get from one floor to another.
	stateGoingUp = iota
	stateGoingDown
	stateNeutral

	// E1--E9
	stepWaitForCall          = 1
	stepChangeOfState        = 2
	stepOpenDoors            = 3
	stepLetPeopleOutIn       = 4
	stepCloseDoors           = 5
	stepPrepareToMove        = 6
	stepGoUpAFloor           = 7
	stepGoDownAFloor         = 8
	stepSetInactionIndicator = 9

	minGiveUpTime = 30 * 10     // 30 seconds
	maxGiveUpTime = 2 * 60 * 10 // 2 minutes

	minInterTime = 1 * 10  // 1 seconds
	maxInterTime = 90 * 10 // 90 seconds

	maxTime = 1000 * 10 // stop simulation after 1000 seconds
)

type node struct {
	info  interface{}
	llink *node
	rlink *node
}

// manipulations of doubly linked lists almost always become much easier if a list head node
// is part of each list
func newDoublyLinkedList() *node {
	n := &node{}
	n.llink = n
	n.rlink = n
	return n
}

func newNode(info interface{}) *node {
	n := &node{
		info: info,
	}
	return n
}

func (x *node) insertRight(p *node) {
	if x == nil || p == nil {
		return
	}
	p.llink = x
	p.rlink = x.rlink
	x.rlink.llink = p
	x.rlink = p
}

func (x *node) insertLeft(p *node) {
	if x == nil || p == nil {
		return
	}
	p.rlink = x
	p.llink = x.llink
	x.llink.rlink = p
	x.llink = p
}

func (x *node) delete() {
	if x == nil {
		return
	}
	x.llink.rlink = x.rlink
	x.rlink.llink = x.llink
	x.llink = x
	x.rlink = x
}

type waitListener interface {
	execute()
}

type waitFunc func()

func newWaitFunc(listener func()) *waitFunc {
	f := waitFunc(listener)
	return &f
}

func (w *waitFunc) execute() {
	(*w)()
}

type waitElement struct {
	nextTime int          // the time when the next action for this entity is to take place
	nextInst waitListener // where this entity is to start executing instructions
}

func newWaitElement(nextTime int, nextInst waitListener) *waitElement {
	return &waitElement{
		nextTime: nextTime,
		nextInst: nextInst,
	}
}

// Each entity waiting for time to pass is placed in a doubly linked list
func newWaitQueue() *node {
	n := newDoublyLinkedList()
	n.info = newWaitElement(0, nil)
	return n
}

// Subroutine SORTIN adds the current node to the WAIT list, sorting
// it into the right place based on its NEXTTIME field.
func (x *node) sortIn(w *waitElement) *node {
	for {
		x = x.llink
		if w.nextTime >= x.info.(*waitElement).nextTime {
			n := newNode(w)
			x.insertRight(n)
			return n
		}
	}
}

//  Subroutine IMMED inserts the current node at the front of the WAIT list.
func (x *node) immed(w *waitElement) *node {
	n := newNode(w)
	x.insertRight(n)
	return n
}

type elevator struct {
	// On each floor there are two call buttons, one for UP and one for DOWN.
	// (Actually floor 0 has only UP and floor 4 has only DOWN, but we may ignore
	// that anomaly since the excess buttons will never be used.) Corresponding to
	// these buttons, there are ten variables CALLUP[j] and CALLDOWN[j], 0 ≤ j ≤ 4.
	// There are also variables CALLCAR[j], 0 ≤ j ≤ 4, representing buttons within
	// the elevator car, which direct it to a destination floor. When a person presses a
	// button, the appropriate variable is set to 1; the elevator clears the variable to 0
	// after the request has been fulfilled.
	callUp   []bool
	callDown []bool
	callCar  []bool
	floor    int     // the current position of the elevator
	d1       bool    // false except during the time people are getting in or out of the elevator
	d2       bool    // becomes false if the elevator has sat on one floor without moving for 30 sec or more
	d3       bool    // false except when the doors are open but nobody is getting in or out of the elevator
	state    int     // the current state of the elevator (GOINGUP, GOINGDOWN, or NEUTRAL)
	step     int     // constants refer to steps E1--E9
	elev1    *node   // elevator actions, except for E5 and E9.
	elev2    *node   // independent elevator action at E5
	elev3    *node   // independent elevator action at E9
	stack    *node   // a stack-like list representing the people now on board the elevator.
	queue    []*node // linear lists representing the people waiting on each floor
}

// Initially FLOOR = 2, D1 = D2 = D3 = 0, and STATE = NEUTRAL.
func newElevator() *elevator {
	e := &elevator{
		callUp:   make([]bool, floors),
		callDown: make([]bool, floors),
		callCar:  make([]bool, floors),
		floor:    floorHome,
		state:    stateNeutral,
		step:     stepWaitForCall,
		stack:    newDoublyLinkedList(),
		queue:    make([]*node, floors),
	}
	for i := floors - 1; i >= 0; i-- {
		e.queue[i] = newDoublyLinkedList()
	}
	return e
}

type simulator struct {
	time   int // simulated time clock (tenths of seconds)
	userID int // user ID counter
	random *rand.Rand
	ele    *elevator

	// Each entity waiting for time to pass is placed in a doubly linked
	// list called the WAIT list; this “agenda” is sorted on the NEXTTIME fields of its
	// nodes, so that the actions may be processed in the correct sequence of simulated
	// times.
	wait *node
}

func newSimulator() *simulator {
	return &simulator{
		wait:   newWaitQueue(),
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
		ele:    newElevator(),
	}
}

func (s *simulator) scheduleElevator(elev **node, delay int, listener waitListener) {
	(*elev).delete()
	*elev = s.wait.sortIn(newWaitElement(s.time+delay, listener))
}

func (s *simulator) scheduleElevatorImmediately(elev **node, listener waitListener) {
	(*elev).delete()
	*elev = s.wait.immed(newWaitElement(s.time, listener))
}

type user struct {
	id         int
	in         int // the floor on which the new user has entered the system
	out        int // the floor to which this user wants to go (OUT ̸= IN)
	giveUpTime int // time user will wait for elevator before running out of patience and deciding to walk
	listNode   *node
	giveUp     *node
}

func newUser(id, in, out, giveUpTime int) *user {
	return &user{
		id:         id,
		in:         in,
		out:        out,
		giveUpTime: giveUpTime,
	}
}

// U1. [Enter, prepare for successor.] The following quantities are determined in
// some manner that will not be specified here:
// IN, the floor on which the new user has entered the system;
// OUT, the floor to which this user wants to go (OUT ̸= IN);
// GIVEUPTIME, the amount of time this user will wait for the elevator before
// running out of patience and deciding to walk;
// INTERTIME, the amount of time before another user will enter the system.
// After these quantities have been computed, the simulation program sets
// things up so that another user enters the system at TIME + INTERTIME.
func (s *simulator) userEnterPrepareForSuccessor() {
	s.userID++
	in := int(s.random.Int31n(floors))
	out := int(s.random.Int31n(floors - 1))
	if out >= in {
		out++
	}
	u := newUser(s.userID, in, out, int(minGiveUpTime+s.random.Int31n(maxGiveUpTime-minGiveUpTime)))
	s.wait.sortIn(newWaitElement(s.time+int(minInterTime+s.random.Int31n(maxInterTime-minInterTime)),
		newWaitFunc(s.userEnterPrepareForSuccessor)))
	s.wait.immed(newWaitElement(s.time, newWaitFunc(func() { s.userSignalAndWait(u) })))
	s.print("U1", "User %d arrives at floor %d, destination is %d.", u.id, u.in, u.out)
}

// U2. [Signal and wait.] (The purpose of this step is to call for the elevator; some
// special cases arise if the elevator is already on the right floor.) If FLOOR = IN
// and if the elevator’s next action is step E6 below (that is, if the elevator doors
// are now closing), send the elevator immediately to its step E3 and cancel its
// activity E6. (This means that the doors will open again before the elevator
// moves.) If FLOOR = IN and if D3 ̸= 0, set D3 ← 0, set D1 to a nonzero value,
// and start up the elevator’s activity E4 again. (This means that the elevator
// doors are open on this floor, but everyone else has already gotten on or
// off. Elevator step E4 is a sequencing step that grants people permission to
// enter the elevator according to normal laws of courtesy; therefore, restarting
// E4 gives this user a chance to get in before the doors close.) In all other
// cases, the user sets CALLUP[IN] ← 1 or CALLDOWN[IN] ← 1, according as
// OUT > IN or OUT < IN; and if D2 = 0 or the elevator is in its “dormant”
// position E1, the DECISION subroutine specified below is performed. (The
// DECISION subroutine is used to take the elevator out of NEUTRAL state at
// certain critical times.)
func (s *simulator) userSignalAndWait(u *user) {
	if s.ele.floor == u.in && s.ele.step == stepCloseDoors {
		s.print("U2", "User %d arrives at doors closing and stop them.", u.id)
		s.scheduleElevatorImmediately(&s.ele.elev1, newWaitFunc(s.executeOpenDoors))
	} else if s.ele.floor == u.in && s.ele.d3 {
		s.print("U2", "User %d arrives at open doors.", u.id)
		s.ele.d3 = false
		s.ele.d1 = true
		s.scheduleElevatorImmediately(&s.ele.elev1, newWaitFunc(s.executeLetPeopleOutIn))
	} else {
		if u.out > u.in {
			s.print("U2", "User %d presses up button.", u.id)
			s.ele.callUp[u.in] = true
		} else {
			s.print("U2", "User %d presses down button.", u.id)
			s.ele.callDown[u.in] = true
		}
		if !s.ele.d2 || s.ele.step == stepWaitForCall {
			s.decision()
		}
	}
	s.wait.immed(newWaitElement(s.time, newWaitFunc(func() { s.userEnterQueue(u) })))
}

// U3. [Enter queue.] Insert this user at the rear of QUEUE[IN], which is a linear
// list representing the people waiting on this floor. Now the user waits
// patiently for GIVEUPTIME units of time, unless the elevator arrives first—
// more precisely, unless step E4 of the elevator routine below sends this user
// to U5 and cancels the scheduled activity U4.
func (s *simulator) userEnterQueue(u *user) {
	s.print("U3", "User %d stands in queue in front of elevator.", u.id)
	u.listNode = newNode(u)
	s.ele.queue[u.in].insertLeft(u.listNode) // enqueue left
	u.giveUp = s.wait.sortIn(newWaitElement(s.time+u.giveUpTime, newWaitFunc(func() { s.userGiveUp(u) })))
}

// U4. [Give up.] If FLOOR ̸= IN or D1 = 0, delete this user from QUEUE[IN]
// and from the simulated system. (The user has decided that the elevator is
// too slow, or that a bit of exercise will be better than an elevator ride.) If
// FLOOR = IN and D1 ̸= 0, the user stays and waits (knowing that the wait
// won’t be long).
func (s *simulator) userGiveUp(u *user) {
	if s.ele.floor != u.in || !s.ele.d1 {
		s.print("U4", "User %d decides to give up, leaves the system.", u.id)
		u.listNode.delete()
	} else {
		s.print("U4", "User %d almost gave up, but stays and waits.", u.id)
	}
}

// U5. [Get in.] This user now leaves QUEUE[IN] and enters ELEVATOR, which is
// a stack-like list representing the people now on board the elevator. Set
// CALLCAR[OUT] ← 1.
// Now if STATE = NEUTRAL, set STATE ← GOINGUP or GOINGDOWN as
// appropriate, and set the elevator’s activity E5 to be executed after 25 units
// of time. (This is a special feature of the elevator, allowing the doors to close
// faster than usual if the elevator is in NEUTRAL state when the user selects a
// destination floor. The 25-unit time interval gives step E4 the opportunity
// to make sure that D1 is properly set up by the time step E5, the door-closing
// action, occurs.)
// Now the user waits until being sent to step U6 by step E4 below, when
// the elevator has reached the desired floor.
func (s *simulator) userGetIn(u *user) {
	s.print("U5", "User %d gets in.", u.id)
	u.listNode.delete()
	u.giveUp.delete()
	s.ele.stack.insertLeft(u.listNode) // push left
	s.ele.callCar[u.out] = true
	if s.ele.state == stateNeutral {
		if u.out > u.in {
			s.ele.state = stateGoingUp
		} else {
			s.ele.state = stateGoingDown
		}
		s.scheduleElevator(&s.ele.elev2, 25, newWaitFunc(s.executeCloseDoors))
	}
}

// U6. [Get out.] Delete this user from the ELEVATOR list and from the simulated
// system.
func (s *simulator) userGetOut(u *user) {
	s.print("U6", "User %d gets out, leaves the system.", u.id)
	u.listNode.delete()
}

// E1. [Wait for call.] (At this point the elevator is sitting at floor 2 with the doors
// closed, waiting for something to happen.) If someone presses a button, the
// DECISION subroutine will take us to step E3 or E6. Meanwhile, wait.
func (s *simulator) executeWaitForCall() {
	s.print("E1", "Elevator dormant")
	s.ele.step = stepWaitForCall
}

func (s *simulator) isAllCallsAboveFalse() bool {
	for j := s.ele.floor + 1; j < floors; j++ {
		if s.ele.callUp[j] || s.ele.callDown[j] || s.ele.callCar[j] {
			return false
		}
	}
	return true
}

func (s *simulator) isAllCallsBelowFalse() bool {
	for j := s.ele.floor - 1; j >= 0; j-- {
		if s.ele.callUp[j] || s.ele.callDown[j] || s.ele.callCar[j] {
			return false
		}
	}
	return true
}

// E2. [Change of state?] If STATE = GOINGUP and CALLUP[j] = CALLDOWN[j] =
// CALLCAR[j] = 0 for all j > FLOOR, then set STATE ← NEUTRAL or STATE ←
// GOINGDOWN, according as CALLCAR[j] = 0 for all j < FLOOR or not, and set
// all CALL variables for the current floor to zero. If STATE = GOINGDOWN, do
// similar actions with directions reversed.
func (s *simulator) executeChangeOfState() {
	s.print("E2", "Elevator stops.")
	s.ele.step = stepChangeOfState
	if s.ele.state == stateGoingUp && s.isAllCallsAboveFalse() {
		if s.isAllCallsBelowFalse() {
			s.ele.state = stateNeutral
		} else {
			s.ele.state = stateGoingDown
		}
		s.ele.callUp[s.ele.floor] = false
		s.ele.callDown[s.ele.floor] = false
		s.ele.callCar[s.ele.floor] = false
	} else if s.ele.state == stateGoingDown && s.isAllCallsBelowFalse() {
		if s.isAllCallsAboveFalse() {
			s.ele.state = stateNeutral
		} else {
			s.ele.state = stateGoingUp
		}
		s.ele.callUp[s.ele.floor] = false
		s.ele.callDown[s.ele.floor] = false
		s.ele.callCar[s.ele.floor] = false
	}
	s.scheduleElevatorImmediately(&s.ele.elev1, newWaitFunc(s.executeOpenDoors))
}

// E3. [Open doors.] Set D1 and D2 to any nonzero values. Set elevator activity
// E9 to start up independently after 300 units of time. (This activity may be
// canceled in step E6 below before it occurs. If it has already been scheduled
// and not canceled, we cancel it and reschedule it.) Also set elevator activity
// E5 to start up independently after 76 units of time. Then wait 20 units of
// time (to simulate opening of the doors) and go to E4.
func (s *simulator) executeOpenDoors() {
	s.print("E3", "Elevator doors start to open.")
	s.ele.step = stepOpenDoors
	s.ele.d1 = true
	s.ele.d2 = true
	s.scheduleElevator(&s.ele.elev3, 300, newWaitFunc(s.executeSetInactionIndicator))
	s.scheduleElevator(&s.ele.elev2, 76, newWaitFunc(s.executeCloseDoors))
	s.scheduleElevator(&s.ele.elev1, 20, newWaitFunc(s.executeLetPeopleOutIn))
}

// E4. [Let people out, in.] If anyone in the ELEVATOR list has OUT = FLOOR, send
// the user of this type who has most recently entered immediately to step U6,
// wait 25 units, and repeat step E4. If no such users exist, but QUEUE[FLOOR]
// is not empty, send the front person of that queue immediately to step U5
// instead of U4, wait 25 units, and repeat step E4. But if QUEUE[FLOOR]
// is empty, set D1 ← 0, make D3 nonzero, and wait for some other activity
// to initiate further action. (Step E5 will send us to E6, or step U2 will
// restart E4.)
func (s *simulator) executeLetPeopleOutIn() {
	s.ele.step = stepLetPeopleOutIn
	p := s.ele.stack
	for {
		p = p.llink // pop left
		if p == s.ele.stack {
			break
		} else {
			u := p.info.(*user)
			if u.out == s.ele.floor {
				s.print("E4", "Doors are open. Users about to exit.")
				s.wait.immed(newWaitElement(s.time, newWaitFunc(func() { s.userGetOut(u) })))
				s.scheduleElevator(&s.ele.elev1, 25, newWaitFunc(s.executeLetPeopleOutIn))
				return
			}
		}
	}
	p = s.ele.queue[s.ele.floor]
	for {
		p = p.rlink // dequeue right
		if p == s.ele.queue[s.ele.floor] {
			break
		} else {
			s.print("E4", "Doors are open. Users about to enter.")
			u := p.info.(*user)
			s.wait.immed(newWaitElement(s.time, newWaitFunc(func() { s.userGetIn(u) })))
			s.scheduleElevator(&s.ele.elev1, 25, newWaitFunc(s.executeLetPeopleOutIn))
			return
		}
	}
	s.print("E4", "Doors are open. Nobody outside elevator.")
	s.ele.d1 = false
	s.ele.d3 = true
}

// E5. [Close doors.] If D1 ̸= 0, wait 40 units and repeat this step (the doors flutter
// a little, but they spring open again, since someone is still getting out or in).
// Otherwise set D3 ← 0 and set the elevator to start at step E6 after 20 units
// of time. (This simulates closing the doors after people have finished getting
// in or out; but if a new user enters on this floor while the doors are closing,
// they will open again as stated in step U2.)
func (s *simulator) executeCloseDoors() {
	s.ele.step = stepCloseDoors
	if s.ele.d1 {
		s.print("E5", "Doors flutter.")
		s.scheduleElevator(&s.ele.elev2, 40, newWaitFunc(s.executeCloseDoors))
	} else {
		s.print("E5", "Elevator doors start to close.")
		s.ele.d3 = false
		s.scheduleElevator(&s.ele.elev1, 20, newWaitFunc(s.executePrepareToMove))
	}
}

// E6. [Prepare to move.] Set CALLCAR[FLOOR] to zero; also set CALLUP[FLOOR]
// to zero if STATE ̸= GOINGDOWN, and also set CALLDOWN[FLOOR] to zero if
// STATE ̸= GOINGUP. (Note: If STATE = GOINGUP, the elevator does not clear
// out CALLDOWN, since it assumes that people who are going down will not
// have entered; but see exercise 6.) Now perform the DECISION subroutine.
// If STATE = NEUTRAL even after the DECISION subroutine has acted, go
// to E1. Otherwise, if D2 ̸= 0, cancel the elevator activity E9. Finally, if
// STATE = GOINGUP, wait 15 units of time (for the elevator to build up speed)
// and go to E7; if STATE = GOINGDOWN, wait 15 units and go to E8.
func (s *simulator) executePrepareToMove() {
	s.ele.step = stepPrepareToMove
	s.ele.callCar[s.ele.floor] = false
	if s.ele.state != stateGoingDown {
		s.ele.callUp[s.ele.floor] = false
	}
	if s.ele.state != stateGoingUp {
		s.ele.callDown[s.ele.floor] = false
	}
	s.decision()
	if s.ele.state == stateNeutral {
		s.print("E6", "Elevator about to go dormant")
		s.scheduleElevatorImmediately(&s.ele.elev1, newWaitFunc(s.executeWaitForCall))
	} else {
		if s.ele.d2 {
			s.ele.elev3.delete()
		}
		if s.ele.state == stateGoingUp {
			s.print("E6", "Elevator about to go up")
			s.scheduleElevator(&s.ele.elev1, 15, newWaitFunc(s.executeGoUpAFloor))
		} else {
			s.print("E6", "Elevator about to go down")
			s.scheduleElevator(&s.ele.elev1, 15, newWaitFunc(s.executeGoDownAFloor))
		}
	}
}

// E7. [Go up a floor.] Set FLOOR ← FLOOR + 1 and wait 51 units of time. If
// now CALLCAR[FLOOR] = 1 or CALLUP[FLOOR] = 1, or if ((FLOOR = 2 or
// CALLDOWN[FLOOR] = 1) and CALLUP[j] = CALLDOWN[j] = CALLCAR[j] = 0
// for all j > FLOOR), wait 14 units (for deceleration) and go to E2. Otherwise,
// repeat this step.
func (s *simulator) executeGoUpAFloor() {
	s.print("E7", "Elevator moving up")
	s.ele.step = stepGoUpAFloor
	s.ele.floor++
	s.scheduleElevator(&s.ele.elev1, 51, newWaitFunc(s.executeGoUpAFloor2))
}

func (s *simulator) executeGoUpAFloor2() {
	if s.ele.callCar[s.ele.floor] || s.ele.callUp[s.ele.floor] ||
		((s.ele.floor == 2 || s.ele.callDown[s.ele.floor]) && s.isAllCallsAboveFalse()) {
		s.scheduleElevator(&s.ele.elev1, 14, newWaitFunc(s.executeChangeOfState))
	} else {
		s.scheduleElevatorImmediately(&s.ele.elev1, newWaitFunc(s.executeGoUpAFloor))
	}
}

// E8. [Go down a floor.] This step is like E7 with directions reversed, and also
// the times 51 and 14 are changed to 61 and 23, respectively. (It takes the
// elevator longer to go down than up.)
func (s *simulator) executeGoDownAFloor() {
	s.print("E8", "Elevator moving down")
	s.ele.step = stepGoDownAFloor
	s.ele.floor--
	s.scheduleElevator(&s.ele.elev1, 61, newWaitFunc(s.executeGoDownAFloor2))
}

func (s *simulator) executeGoDownAFloor2() {
	if s.ele.callCar[s.ele.floor] || s.ele.callDown[s.ele.floor] ||
		((s.ele.floor == 2 || s.ele.callUp[s.ele.floor]) && s.isAllCallsBelowFalse()) {
		s.scheduleElevator(&s.ele.elev1, 23, newWaitFunc(s.executeChangeOfState))
	} else {
		s.scheduleElevatorImmediately(&s.ele.elev1, newWaitFunc(s.executeGoDownAFloor))
	}
}

// E9. [Set inaction indicator.] Set D2 ← 0 and perform the DECISION subroutine.
// (This independent action is initiated in step E3 but it is almost always
// canceled in step E6. See exercise 4.)
func (s *simulator) executeSetInactionIndicator() {
	s.print("E9", "Elevator not active")
	s.ele.d2 = false
	s.decision()
}

// Subroutine D (DECISION subroutine). This subroutine is performed at certain
// critical times, as specified in the coroutines above, when a decision about the
// elevator’s next direction is to be made.
func (s *simulator) decision() {

	// D1. [Decision necessary?] If STATE ̸= NEUTRAL, exit from this subroutine.
	if s.ele.state != stateNeutral {
		return
	}

	// D2. [Should doors open?] If the elevator is positioned at E1 and if CALLUP[2],
	// CALLCAR[2], and CALLDOWN[2] are not all zero, cause the elevator to start
	// its activity E3 after 20 units of time, and exit from this subroutine. (If
	// the DECISION subroutine is currently being invoked by the independent
	// activity E9, it is possible for the elevator coroutine to be positioned at E1.)
	if s.ele.step == stepWaitForCall && (s.ele.callUp[floorHome] || s.ele.callCar[floorHome] || s.ele.callDown[floorHome]) {
		s.scheduleElevator(&s.ele.elev1, 20, newWaitFunc(s.executeOpenDoors))
		return
	}

	// D3. [Any calls?] Find the smallest j ̸= FLOOR for which CALLUP[j], CALLCAR[j],
	// or CALLDOWN[j] is nonzero, and go on to step D4. But if no such j exists,
	// then set j ← 2 if the DECISION subroutine is currently being invoked by
	// step E6; otherwise exit from this subroutine.
	j := 0
	for ; j < floors; j++ {
		if j != s.ele.floor && (s.ele.callUp[j] || s.ele.callCar[j] || s.ele.callDown[j]) {
			goto D4
		}
	}
	if s.ele.step == stepPrepareToMove {
		j = 2
	} else {
		return
	}

D4: // D4. [Set STATE.] If FLOOR > j, set STATE ← GOINGDOWN; if FLOOR < j, set
	// STATE ← GOINGUP.
	if s.ele.floor > j {
		s.ele.state = stateGoingDown
	} else if s.ele.floor < j {
		s.ele.state = stateGoingUp
	}

	// D5. [Elevator dormant?] If the elevator coroutine is positioned at step E1, and
	// if j ̸= 2, set the elevator to perform step E6 after 20 units of time. Exit
	// from the subroutine.
	if s.ele.step == stepWaitForCall && j != 2 {
		s.scheduleElevator(&s.ele.elev1, 20, newWaitFunc(s.executePrepareToMove))
		return
	}
}

func (s *simulator) print(step, action string, a ...interface{}) {
	var state, d1, d2, d3 rune
	switch s.ele.state {
	case stateGoingDown:
		state = 'D'
	case stateGoingUp:
		state = 'U'
	default:
		state = 'N'
	}
	if s.ele.d1 {
		d1 = 'X'
	} else {
		d1 = '0'
	}
	if s.ele.d2 {
		d2 = 'X'
	} else {
		d2 = '0'
	}
	if s.ele.d3 {
		d3 = 'X'
	} else {
		d3 = '0'
	}
	action = fmt.Sprintf(action, a...)
	fmt.Printf("%04d\t%c\t%d\t%c\t%c\t%c\t%s\t%s\n", s.time, state, s.ele.floor, d1, d2, d3, step, action)
}

// The heart of the simulation control: It decides which activity is to act
// next (namely, the first element of the WAIT list, which we know is nonempty),
// and jumps to it.
func main() {
	fmt.Println("TIME\tSTATE\tFLOOR\tD1\tD2\tD3\tstep\taction")
	s := newSimulator()
	s.userEnterPrepareForSuccessor()
	for {
		n := s.wait.rlink
		if n == s.wait {
			fmt.Println("ERROR: Wait queue is empty.")
			break
		}
		n.delete()
		w := n.info.(*waitElement)
		s.time = w.nextTime
		if s.time >= maxTime {
			break
		}
		w.nextInst.execute()
	}
}
