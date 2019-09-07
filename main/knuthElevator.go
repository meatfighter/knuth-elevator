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
	stepLetPeopleInOut       = 4
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

func (x *node) deleteElement(info interface{}) {
	p := x
	for {
		p = p.llink
		if p == x {
			return
		} else if p.info == info {
			p.delete()
			return
		}
	}
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

type elevator struct {
	callUp   []bool
	callDown []bool
	callCar  []bool
	floor    int
	d1       bool
	d2       bool
	d3       bool
	state    int
	step     int
	pending  *node
}

func newElevator() *elevator {
	return &elevator{
		callUp:   make([]bool, floors),
		callDown: make([]bool, floors),
		callCar:  make([]bool, floors),
		floor:    floorHome,
		state:    stateNeutral,
		step:     stepWaitForCall,
	}
}

type simulator struct {
	time   int
	userID int
	wait   *node
	random *rand.Rand
	ele    *elevator
}

func newSimulator() *simulator {
	return &simulator{
		wait:   newWaitQueue(),
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
		ele:    newElevator(),
	}
}

func (s *simulator) scheduleElevator(delay int, listener waitListener) {
	s.ele.pending.delete()
	s.ele.pending = s.wait.sortIn(newWaitElement(s.time+delay, listener))
}

func (s *simulator) u() {
	/*
		// u1
		s.userID++
		id := s.userID
		in := int(s.random.Int31n(floors))
		out := int(s.random.Int31n(floors - 1))
		if out >= in {
			out++
		}
		giveUpTime := int(minGiveUpTime + s.random.Int31n(maxGiveUpTime-minGiveUpTime))
		interTime := int(minInterTime + s.random.Int31n(maxInterTime-minInterTime))
		s.wait.sortIn(newWaitElement(s.time+interTime, newWaitFunc(s.u)))

		// u2
	*/
}

// E1. [Wait for call.] (At this point the elevator is sitting at floor 2 with the doors
// closed, waiting for something to happen.) If someone presses a button, the
// DECISION subroutine will take us to step E3 or E6. Meanwhile, wait.
func (s *simulator) executeWaitForCall() {
	s.ele.step = stepWaitForCall
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
	s.scheduleElevator(300, newWaitFunc(s.executeSetInactionIndicator))
	// TODO ...
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
		s.scheduleElevator(20, newWaitFunc(s.executeOpenDoors))
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
		s.scheduleElevator(20, newWaitFunc(s.executePrepareToMove))
	}
}

func main() {
	//s := newSimulator()
}
