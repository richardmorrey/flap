package flap

import (
	//"errors"
	//"math"
	//"sort"
	//"time"
)

type Promise struct {
	TripStart 	EpochTime
	TripEnd	  	EpochTime
	ClearanceDate	EpochTime
}

const MaxPromises=10

type Promises struct {
	entries			[MaxPromises]Promise
	predictor		predictor
}

// Propose returns a proposal for a clearance promise date for a Trip with given
// start and end dates and schedule. The promise is not made at this point
func (self *Promises) Propose(tripStart EpochTime, tripEnd EpochTime, distance Kilometres) (Promise,error) {
	return Promise{},ENOTIMPLEMENTED
}

// Make enforces the given proposed promise by adding it to the list of promises, but
// only if the predicted end data hasnt changed.
func (self *Promises) Make(pp Promise) error {
	return ENOTIMPLEMENTED
}

// Keep asks for a promise matching the providid completed trip to be kept. If a matching
// valid promise is found its clearance date is returned for use by the Traveller. Otherwise
// an error is returned
func (self* Promises) Keep(tripStart EpochTime, tripEnd EpochTime, distance Kilometres) (EpochTime,error) {
	return EpochTime(0),ENOTIMPLEMENTED
}

// Delete a promise with the given start and end date
func (self* Promises) Delete(tripStart EpochTime, tripEnd EpochTime) error {
	return ENOTIMPLEMENTED
}

