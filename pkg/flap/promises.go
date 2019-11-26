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

type PromiseProposal struct {
	Promise
	proposalTime EpochTime
}

const MaxPromises=10

type Promises struct {
	entries			[MaxPromises]Promise
}

// Propose returns a proposal for a clearance promise date for a Trip with given
// start and end dates and schedule. This proposal is valid only until line
// of best fit is next calculated
func (self *Promises) Propose(tripStart EpochTime, tripEnd EpochTime, distance Kilometres) (PromiseProposal,error) {
	return PromiseProposal{},ENOTIMPLEMENTED
}

// Make  turns a promise proposal into a committed promise
func (self *Promises) Make(pp PromiseProposal) error {
	return ENOTIMPLEMENTED
}

// Keep asks for a promise matching the providid completed trip to be kept. If a matching
// valid promise is found its clearance data is returned for use by the Traveller. Otherwise
// an error is returned
func (self* Promises) Keep(tripStart EpochTime, tripEnd EpochTime, distance Kilometres) (EpochTime,error) {
	return EpochTime(0),ENOTIMPLEMENTED
}

// Delete a promise with the given start and end date
func (self* Promises) Delete(tripStart EpochTime, tripEnd EpochTime) error {
	return ENOTIMPLEMENTED
}

