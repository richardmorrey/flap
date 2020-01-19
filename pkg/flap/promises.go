package flap

import (
	"errors"
	"sort"
	//"fmt"
)

var EINTERNAL			 = errors.New("Reached internal state that shouldn't be possible")
var ENOROOMFORMOREPROMISES 	 = errors.New("No room for more promises")
var EOVERLAPSWITHNEXTPROMISE 	 = errors.New("Overlaps with next promise")
var EOVERLAPSWITHPREVPROMISE	 = errors.New("Overlaps with previous promise")
var EPROMISENOTFOUND		 = errors.New("Promise not found")
var EPROMISEDOESNTMATCH		 = errors.New("Promise trip end or distance travelled doesnt match")
var EEXCEEDEDMAXSTACKSIZE	 = errors.New("Exceeded max promise stack size")
var EPROPOSALEXPIRED		 = errors.New("Proposal has expired")

type StackIndex int8
type Promise struct {
	TripStart 	EpochTime
	TripEnd	  	EpochTime
	Distance	Kilometres
	Clearance	EpochTime
	StackIndex	StackIndex
	CarriedOver	Kilometres
}

func (self *Promise) older (p Promise) bool {
	return bool(p.TripStart >= self.TripStart)
}

func (self *Promise) tobackfill() Kilometres {
	return self.Distance + self.CarriedOver
}

const MaxPromises=10

type Promises struct {
	entries			[MaxPromises]Promise
}

type Proposal struct {
	Promises
	version predictVersion
}

// Propose returns a proposal for a clearance promise date for a Trip with given
// start and end dates and schedule. The promise is not made at this point
func (self *Promises) propose(tripStart EpochTime,tripEnd EpochTime,distance Kilometres, now EpochTime, predictor predictor) (*Proposal,error) {

	// Check args
	if predictor == nil {
		return nil,logError(EINVALIDARGUMENT)
	}
	if tripEnd <= tripStart {
		return nil,logError(EINVALIDARGUMENT)
	}
	if distance <= 0 {
		return nil,logError(EINVALIDARGUMENT)
	}
	if tripStart <= now {
		return nil,logError(EINVALIDARGUMENT)
	}
	if tripStart == 0 {
		return nil,logError(EINVALIDARGUMENT)
	}

	// Check that oldest promise can be dropped if we are full
	if self.entries[MaxPromises-1].TripStart >= now && self.entries[MaxPromises-1].TripStart > 0 {
		return nil,ENOROOMFORMOREPROMISES
	}

	// Create a copy of the current promises to work on
	var pp Proposal
	pp.entries = self.entries

	// Calculate clearance date, defaulting to the next day if predictor
	// is not ready yet
	var p Promise
	clearance,err := predictor.predict(distance,tripEnd.toEpochDays(true))
	if err !=nil {
		clearance = tripEnd.toEpochDays(false)+1
		logDebug("predict failed: ",err)
	}
	p = Promise{TripStart:tripStart,TripEnd:tripEnd,Distance:distance,Clearance:clearance.toEpochTime()}
	
	// Find index to add promise
	i := sort.Search(MaxPromises, func(i int) bool { return self.entries[i].older(p)})
	if  i >= MaxPromises {
		return nil,logError(EINTERNAL)
	}
	
	// Confirm that there is no trip overlap with promise before or after
	if (i < MaxPromises) && (pp.entries[i].TripEnd) >= p.TripStart {
		return nil,logError(EOVERLAPSWITHPREVPROMISE)
	}
	if i > 0 && pp.entries[i-1].TripStart <= p.TripEnd {
		return nil,logError(EOVERLAPSWITHNEXTPROMISE)
	}

	// Copy older entries down one - the oldest is dropped - and insert
	copy(pp.entries[i+1:], pp.entries[i:])
	pp.entries[i] = p
	pp.version = predictor.version()

	// Stack promises to ensure no overlap
	err = pp.restack(i,predictor)
	if err == nil {
		return &pp,nil
	} else {
		return nil,logError(err)
	}
}

// make enforces the given promise proposal by overwriting the current list of promises
// with it, but only if the predictor is the same version uses to make the proposal.
func (self *Promises) make(pp *Proposal, predictor predictor) error {
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

	// Check for valid trip details
	if distance == 0 {
		return EpochTime(0),EINVALIDARGUMENT
	}

	// Look from oldest to newest promise looking for first entry where:
	// 1) Start time given is greater than or equal to entry start time
	// 2) End time given is less than or equal to entry end time
	// 3) Distance given is equal to entry distance
	for i:=MaxPromises-1; i>=0 && tripStart >= self.entries[i].TripStart; i-- {
		if tripEnd <= self.entries[i].TripEnd && self.entries[i].Distance == distance {
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

	// Set clearance date to start of day of next trip
	cd := self.entries[i-1].TripStart.toEpochDays(false)
	self.entries[i].Clearance = cd.toEpochTime() 

	// Set stack index to one more than that of previous
	// promise if there is one
	var lastIndex StackIndex
	if i < MaxPromises-1 {
		if self.entries[i+1].StackIndex >= MaxStackSize {
			return EEXCEEDEDMAXSTACKSIZE
		}
		lastIndex = self.entries[i+1].StackIndex
	}
	self.entries[i].StackIndex = lastIndex+1

	// Calculate clearance date for next promise, taking account of distance
	// not cleared from promise i
	distdone,err := predictor.backfilled(self.entries[i].TripEnd.toEpochDays(true)+1,self.entries[i].Clearance.toEpochDays(false))
	if err != nil {
		distdone = 0
		logDebug("backfill failed: ",err)
	}
	self.entries[i-1].CarriedOver = self.entries[i].tobackfill() - distdone
	clearance,err := predictor.predict(self.entries[i-1].tobackfill(),self.entries[i-1].TripEnd.toEpochDays(true)+1)
	if err != nil {
		clearance = self.entries[i-1].TripEnd.toEpochDays(false)+1
		logDebug("predict failed: ",err)
	}
	self.entries[i-1].Clearance=clearance.toEpochTime()
	return nil
}

// restack updates the clearance date/stack status for the promise immediately preceding and those following
// the index of the given promise that has been inserted in order to retain the following position
// - Clearance date of each trip is soon enough to allow following trip to start
// - No sequence of more than 3 stacked promises
// If this is not possible then an error is returned. Note this function does not change the TripStart, TripEnd
// or Distance fields of any entry.
const MaxStackSize = 3
func (self* Promises) restack(i int, predictor predictor) error {
	
	// Check previous promise and extend stack if clearance date overlaps
	if  i < MaxPromises -1 && self.entries[i+1].Clearance >= self.entries[i].TripStart {
		err := self.updateStackEntry(i+1,predictor)
		if err != nil {
			return logError(err)
		}
	}

	// Check following promises, extending stack as needed to allow
	// trips to happen
	for j:=i; j > 0 && self.entries[j].Clearance >= self.entries[j-1].TripStart; j-- {
		
		// Update
		err := self.updateStackEntry(j,predictor)
		if err != nil {
			return logError(err)
		}
	}
	return nil
}

// Delete deletes a promise with the given trip start and end date. If the promise
// is stacked then deleteStack can be used to force deletion of entire stack.
func (self* Promises) delete(tripStart EpochTime, tripEnd EpochTime,deleteStack bool) error {
	return ENOTIMPLEMENTED
}

type PromisesIterator struct {
	index int
	promises *Promises
}

func (self *PromisesIterator) Next() (bool) {
	if self.index > 0 {
		self.index--
		return true
	}
	return false
}

func (self *PromisesIterator) Value() Promise {
	return self.promises.entries[self.index]
}

func (self *PromisesIterator) Error() error {
	return nil
}

// NewIterator provides iterator for iterating over all promises from oldest to newest
// by trip start time.
func (self *Promises) NewIterator() *PromisesIterator {
	iter := new(PromisesIterator)
	iter.index = sort.Search(MaxPromises,  func(i int) bool {return self.entries[i].TripStart==0}) 
	iter.promises=self
	return iter
}

