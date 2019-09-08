package main

import (
	"math/rand"
	"time"
)

const (
	floors = 5

	floorSubbasement = 0
	floorBasement    = 1
	floorFirst       = 2
	floorSecond      = 3
	floorThird       = 4

	floorHome = floorFirst

	stateGoingUp = iota
	stateGoingDown
	stateNeutral

	stepWaitForCall          = 1
	stepChangeOfState        = 2
	stepOpenDoors            = 3
	stepLetPeopleOutIn       = 4
	stepCloseDoors           = 5
	stepPrepareToMove        = 6
	stepGoUpAFloor           = 7
	stepGoDownAFloor         = 8
	stepSetInactionIndicator = 9

	minGiveUpTime = 30 * 10
	maxGiveUpTime = 10 * 60 * 10

	minInterTime = 5 * 10
	maxInterTime = 20 * 60 * 10
)

type node struct {
	info  interface{}
	llink *node
	rlink *node
}

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
	nextTime int
	nextInst waitListener
}

func newWaitElement(nextTime int, nextInst waitListener) *waitElement {
	return &waitElement{
		nextTime: nextTime,
		nextInst: nextInst,
	}
}

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
	s.wait.sortIn(newWaitElement(s.time+int(minInterTime+s.random.Int31n(maxInterTime-minInterTime)),
		newWaitFunc(s.userEnterPrepareForSuccessor)))
	s.wait.sortIn(newWaitElement(s.time, newWaitFunc(func() {
		s.userSignalAndWait(newUser(s.userID, in, out, int(minGiveUpTime+s.random.Int31n(maxGiveUpTime-minGiveUpTime))))
	})))
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
		s.scheduleElevatorImmediately(&s.ele.elev1, newWaitFunc(s.executeOpenDoors))
	} else if s.ele.floor == u.in && s.ele.d3 {
		s.ele.d3 = false
		s.ele.d1 = true
		s.scheduleElevatorImmediately(&s.ele.elev1, newWaitFunc(s.executeLetPeopleOutIn))
	} else {
		if u.out > u.in {
			s.ele.callUp[u.in] = true
		} else {
			s.ele.callDown[u.in] = true
		}
		if !s.ele.d2 || s.ele.step == stepWaitForCall {
			s.decision()
		}
	}
	s.wait.sortIn(newWaitElement(s.time, newWaitFunc(func() { s.userEnterQueue(u) })))
}

// U3. [Enter queue.] Insert this user at the rear of QUEUE[IN], which is a linear
// list representing the people waiting on this floor. Now the user waits
// patiently for GIVEUPTIME units of time, unless the elevator arrives first—
// more precisely, unless step E4 of the elevator routine below sends this user
// to U5 and cancels the scheduled activity U4.
func (s *simulator) userEnterQueue(u *user) {
	u.listNode = newNode(u)
	s.ele.queue[u.in].insertLeft(u.listNode)
	u.giveUp = s.wait.sortIn(newWaitElement(s.time+u.giveUpTime, newWaitFunc(func() { s.userGiveUp(u) })))
}

// U4. [Give up.] If FLOOR ̸= IN or D1 = 0, delete this user from QUEUE[IN]
// and from the simulated system. (The user has decided that the elevator is
// too slow, or that a bit of exercise will be better than an elevator ride.) If
// FLOOR = IN and D1 ̸= 0, the user stays and waits (knowing that the wait
// won’t be long).
func (s *simulator) userGiveUp(u *user) {
	if s.ele.floor != u.in || !s.ele.d1 {
		u.listNode.delete()
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
	u.listNode.delete()
	s.ele.stack.insertLeft(u.listNode)
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
	u.listNode.delete()
}

// E1. [Wait for call.] (At this point the elevator is sitting at floor 2 with the doors
// closed, waiting for something to happen.) If someone presses a button, the
// DECISION subroutine will take us to step E3 or E6. Meanwhile, wait.
func (s *simulator) executeWaitForCall() {
	s.ele.step = stepWaitForCall
}

// E2. [Change of state?] If STATE = GOINGUP and CALLUP[j] = CALLDOWN[j] =
// CALLCAR[j] = 0 for all j > FLOOR, then set STATE ← NEUTRAL or STATE ←
// GOINGDOWN, according as CALLCAR[j] = 0 for all j < FLOOR or not, and set
// all CALL variables for the current floor to zero. If STATE = GOINGDOWN, do
// similar actions with directions reversed.
func (s *simulator) executeChangeOfState() {
	s.ele.step = stepChangeOfState
	if s.ele.state == stateGoingUp {
		for j := s.ele.floor + 1; j < floors; j++ {
			if s.ele.callUp[j] || s.ele.callDown[j] || s.ele.callCar[j] {
				goto done2
			}
		}
		for j := s.ele.floor - 1; j >= 0; j-- {
			if s.ele.callCar[j] {
				s.ele.state = stateGoingDown
				goto done1
			}
		}
		s.ele.state = stateNeutral
	} else if s.ele.state == stateGoingDown {
		for j := s.ele.floor - 1; j >= 0; j-- {
			if s.ele.callUp[j] || s.ele.callDown[j] || s.ele.callCar[j] {
				goto done2
			}
		}
		for j := s.ele.floor + 1; j < floors; j++ {
			if s.ele.callCar[j] {
				s.ele.state = stateGoingUp
				goto done1
			}
		}
		s.ele.state = stateNeutral
	}
done1:
	s.ele.callUp[s.ele.floor] = false
	s.ele.callDown[s.ele.floor] = false
	s.ele.callCar[s.ele.floor] = false
done2:
	s.scheduleElevator(&s.ele.elev1, 0, newWaitFunc(s.executeOpenDoors))
}

// E3. [Open doors.] Set D1 and D2 to any nonzero values. Set elevator activity
// E9 to start up independently after 300 units of time. (This activity may be
// canceled in step E6 below before it occurs. If it has already been scheduled
// and not canceled, we cancel it and reschedule it.) Also set elevator activity
// E5 to start up independently after 76 units of time. Then wait 20 units of
// time (to simulate opening of the doors) and go to E4.
func (s *simulator) executeOpenDoors() {
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

}

func (s *simulator) executeCloseDoors() {

}

func (s *simulator) executePrepareToMove() {
	s.ele.step = stepPrepareToMove
}

func (s *simulator) executeSetInactionIndicator() {
	s.ele.step = stepSetInactionIndicator
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

func main() {
	//s := newSimulator()
}
