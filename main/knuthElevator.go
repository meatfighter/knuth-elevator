package main

const (
	floors = 5

	floorSubbasement = iota
	floorBasement
	floorFirst
	floorSecond
	floorThird

	floorHome = floorFirst

	stateGoingUp = iota
	stateGoingDown
	stateNeutral
)

type elevator struct {
	callUp   []bool
	callDown []bool
	callCar  []bool
	floor    int
	d1       int
	d2       int
	d3       int
	state    int
}

func newElevator() *elevator {
	return &elevator{
		callUp:   make([]bool, floors),
		callDown: make([]bool, floors),
		callCar:  make([]bool, floors),
		floor:    floorHome,
		state:    stateNeutral,
	}
}

type user struct {
	in         int
	out        int
	giveUpTime int
}

func newUser(in, out, giveUpTime int) *user {
	return &user{
		in:         in,
		out:        out,
		giveUpTime: giveUpTime,
	}
}

func main() {

}
