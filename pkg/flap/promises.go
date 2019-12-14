package flap

import (
	"errors"
	"sort"
	//"fmt"
)

var EPROMISETOOOLD		 = errors.New("Promise is too old to make")
var ENOROOMFORMOREPROMISES 	 = errors.New("No room for more promises")
var EOVERLAPSWITHNEXTPROMISE 	 = errors.New("Overlaps with next promise")
var EOVERLAPSWITHPREVPROMISE	 = errors.New("Overlaps with previous promise")
var EPROMISENOTFOUND		 = errors.New("Promise not found")
var EPROMISEDOESNTMATCH		 = errors.New("Promise trip end or distance travelled doesnt match")
var EEXCEEDEDMAXSTACKSIZE	 = errors.New("Exceeded max promise stack size")
var EPROPOSALEXPIRED		 = errors.New("Proposal has expired")

type Promise struct {
	TripStart 	EpochTime
	TripEnd	  	EpochTime
	Distance	Kilometres
	Clearance	EpochTime
	stackIndex	int8
}

func (self *Promise) older (p Promise) bool {
	return bool(p.TripStart >= self.TripStart)
}

const MaxPromises=10

type Promises struct {
	entries			[MaxPromises]Promise
	version			predictVersion
}

// Propose returns a proposal for a clearance promise date for a Trip with given
// start and end dates and schedule. The promise is not made at this point
func (self *Promises) Propose(tripStart EpochTime,tripEnd EpochTime,distance Kilometres, now EpochTime, predictor predictor) (*Promises,error) {

	// Check args
	if predictor == nil {
		return nil,EINVALIDARGUMENT
	}
	if tripEnd <= tripStart {
		return nil,EINVALIDARGUMENT
	}
	if distance <= 0 {
		return nil,EINVALIDARGUMENT
	}

	// Check that oldest promise can be dropped if we are full
	if self.entries[MaxPromises-1].TripStart >= now {
		return nil,ENOROOMFORMOREPROMISES
	}

	// Create a copy of the current promise to work on
	var pp Promises
	pp.entries = self.entries

	// Calculate clearance date
	var p Promise
	clearance,err := predictor.predict(distance,tripEnd.toEpochDays(true))
	if err == nil {
		p = Promise{tripStart,tripEnd,distance,clearance.toEpochTime(),0}
	}

	// Find index to add promise
	i := sort.Search(MaxPromises, func(i int) bool { return self.entries[i].older(p)})
	if  i >= MaxPromises {
		return nil,EPROMISETOOOLD
	}
	
	// Confirm that there is no trip overlap with promise before or after
	if i < MaxPromises-1 && pp.entries[i+1].TripEnd >= p.TripStart {
		return nil,EOVERLAPSWITHPREVPROMISE
	}
	if i > 0 && pp.entries[i-1].TripStart <= p.TripEnd {
		return nil,EOVERLAPSWITHNEXTPROMISE
	}

	// Copy older entries down one - the oldest is dropped - and insert
	copy(pp.entries[i+1:], pp.entries[i:])
	pp.entries[i] = p

	// Stack promises to ensure no overlap
	err = pp.restack(i,predictor)
	if err == nil {
		return &pp,nil
	} else {
		return nil,err
	}
}

// Make enforces the given promise proposal by overwriting the current list of promises
// with it, but only if the predictor is the same version uses to make the proposal.
func (self *Promises) Make(pp *Promises, predictor predictor) error {
	if predictor.version() != pp.version {
		return EPROPOSALEXPIRED
	}
	self.entries=pp.entries
	return nil
}

// keep asks for a promise applying to completed trip with given details to be kept. If a matching
// valid promise is found its clearance date is returned for use by the Traveller. Otherwise
// an error is returned
func (self* Promises) keep(tripStart EpochTime, tripEnd EpochTime, distance Kilometres) (EpochTime,error) {

	// Look from oldest to newest promise looking for first entry that:
	// 1) Has a start time greater than or equal to that given
	// 2) Has an end time less than or equal to that given
	// 3) has a distance equal to that given
	for i:=MaxPromises-1; i>=0 && self.entries[i].TripEnd <= tripEnd; i-- {
		if self.entries[i].TripStart >= tripStart && self.entries[i].Distance == distance {
			return self.entries[i].Clearance,nil
		} 
	}
	return EpochTime(0), EPROMISEDOESNTMATCH
}

// updateStackEntry updates stack entry i clearance date to allow the trip after to proceed and 
// updates the clearance date of the trip after to account for the early clearance of stack entry
// i
func (self* Promises) updateStackEntry(i int, predictor predictor) error {
	
	// Validate args
	if i==0 || i > MaxPromises-1 {
		return EINVALIDARGUMENT
	}
	if predictor == nil {
		return EINVALIDARGUMENT
	}

	// Set clearance date to start of day before next trip
	cd := self.entries[i-1].TripStart.toEpochDays(false)
	self.entries[i].Clearance = cd.toEpochTime() 
	if self.entries[i].stackIndex == 0 {
		self.entries[i].stackIndex = 1
	}

	// Update stack entry
	if i < MaxPromises-1 && self.entries[i+1].stackIndex == MaxStackSize {
		return EEXCEEDEDMAXSTACKSIZE
	}
	self.entries[i].stackIndex = self.entries[i+1].stackIndex+1

	// Calculate clearance date for next promise, taking account of distance
	// not cleared from the stacked promise
	leftOvers,err := predictor.backfilled(self.entries[i].TripEnd.toEpochDays(true),self.entries[i].Clearance.toEpochDays(false))
	if err != nil {
		return err
	}
	clearance,err := predictor.predict(self.entries[i-1].Distance+leftOvers,self.entries[i-1].TripEnd.toEpochDays(true))
	if err != nil {
		return err
	}
	self.entries[i-1].Clearance=clearance.toEpochTime()
	return nil
}

// restack
const MaxStackSize = 3
func (self* Promises) restack(i int, predictor predictor) error {
	
	// Check previous promise and extend stack if clearance date overlaps
	if  i < MaxPromises -1 && self.entries[i+1].Clearance >= self.entries[i].TripStart {
		err := self.updateStackEntry(i-1,predictor)
		if err != nil {
			return err
		}
	}

	// Check following promises, extending stack as needed to allow
	// trips to happen
	for j:=i; j>0 && self.entries[j+1].Clearance >= self.entries[j].TripStart; j-- {
		
		// Update
		err := self.updateStackEntry(j,predictor)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete deletes a promise with the given trip start and end date. If the promise
// is stacked then deleteStack can be used to force deletion of entire stack.
func (self* Promises) Delete(tripStart EpochTime, tripEnd EpochTime,deleteStack bool) error {
	return ENOTIMPLEMENTED
}

