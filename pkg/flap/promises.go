package flap

import (
	"errors"
	//"math"
	"sort"
	//"time"
)

var EPROMISETOOOLD		 = errors.New("Promise is too old to make")
var ENOROOMFORMOREPROMISES 	 = errors.New("No room for more promises")
var ECLEARANCEDATEHASCHANGED	 = errors.New("Clearance date has changed")
var EOVERLAPSWITHNEXTPROMISE 	 = errors.New("Overlaps with next promise")
var EOVERLAPSWITHPREVPROMISE	 = errors.New("Overlaps with previous promise")
var EPROMISENOTFOUND		 = errors.New("Promise not found")
var EPROMISEDOESNTMATCH		 = errors.New("Promise trip end or distance travelled doesnt match")


// TBD. Once trip reopening is implemented this could be expanded to support "reopen" promises - which
// act if a trip is reopened and the clearance date is a particular value - i.e. a  previous expected promise
// has been made.
type Promise struct {
	TripStart 	EpochTime
	TripEnd	  	EpochTime
	Distance	Kilometres
	Clearance	EpochTime
}

func (self *Promise) older (p Promise) bool {
	return bool(p.TripStart >= self.TripStart)
}

const MaxPromises=10

type Promises struct {
	entries			[MaxPromises]Promise
}

// Propose returns a proposal for a clearance promise date for a Trip with given
// start and end dates and schedule. The promise is not made at this point
func (self *Promises) Propose(tripStart EpochTime, tripEnd EpochTime, distance Kilometres,predictor predictor) (Promise,error) {

	// Check args
	if predictor == nil {
		return Promise{},EINVALIDARGUMENT
	}
	if tripEnd <= tripStart {
		return Promise{},EINVALIDARGUMENT
	}
	if distance <= 0 {
		return Promise{}, EINVALIDARGUMENT
	}

	// Build promise
	var p Promise
	clearance,err := predictor.predict(distance,tripEnd.toEpochDays(true))
	if err == nil {
		p = Promise{tripStart,tripEnd,distance,clearance.toEpochTime()}
	}
	return p,err
}

// Make enforces the given proposed promise by adding it to the list of promises, but
// only if the predicted clearance date hasn't changed.
func (self *Promises) Make(p Promise, now EpochTime, predictor predictor) error {

	// Check that oldest promise can be dropped if we are full
	if self.entries[0].TripStart >= now {
		return ENOROOMFORMOREPROMISES
	}

	// Check clearance date still holds
	clearance,err := predictor.predict(p.Distance,p.TripEnd.toEpochDays(true))
	if err != nil || clearance.toEpochTime() != p.Clearance {
		return ECLEARANCEDATEHASCHANGED
	}

	// Find index to add promise
	i := sort.Search(MaxPromises, func(i int) bool { return self.entries[i].older(p)})
	if  i >= MaxPromises {
		return EPROMISETOOOLD
	}
	
	// Confirm that there is no overlap with promise before or after
	if i < MaxPromises-1 && self.entries[i+1].Clearance >= p.TripStart {
		return EOVERLAPSWITHPREVPROMISE
	}
	if i > 0 && self.entries[i-1].TripStart <= p.Clearance {
		return EOVERLAPSWITHNEXTPROMISE
	}

	// Copy older entries down one - the oldest is dropped - and insert
	copy(self.entries[i+1:], self.entries[i:])
	self.entries[i] = p
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

// Delete a promise with the given start and end date
func (self* Promises) Delete(tripStart EpochTime, tripEnd EpochTime) error {
	return ENOTIMPLEMENTED
}

