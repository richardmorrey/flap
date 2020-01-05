package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"errors"
)

var ENOSPACEFORTRIP = errors.New("No space for trip")

type botPromises struct {
	Weights
	totalDays int
}

// newBotPromises creates a new botPromises with all
// days within allowed range equally likely to be
// chosen as start day for a trip
func newBotPromises(totalDays flap.Days) *botPromises {
	bp := new(botPromises)
	bp.totalDays = int(totalDays)
	return bp
} 

// getPromise chooses dates for a future trip that does not overlap with a
// trip for which a promise already exists. It then tries to obtain a promise
// for the new trip and if successful returns start day of  the trip for planning
// and otherwise an error
func (self* botPromises) getPromise(fe *flap.Engine,pp flap.Passport,currentDay flap.EpochTime,length flap.Days,from flap.ICAOCode, to flap.ICAOCode) (flap.Days,error) {
	
	// Retrieve traveller object
	t,err := fe.Travellers.GetTraveller(pp)
	if err != nil {
		return 0, glog(err)
	}

	// Build weights to cover all possible days for start of the trip
	// making sure that any day that is not suitable (is part of a planned trip 
	// or is too close to start of a planned trip) has zero weight
	self.reset()
	daysOffset:=flap.Days(currentDay/flap.SecondsInDay)
	it := t.Promises.NewIterator()
	var sd,ed flap.Days
	for it.Next() {
		sd = flap.Days(it.Value().TripStart/flap.SecondsInDay) - daysOffset - length
		ed = flap.Days(it.Value().TripEnd/flap.SecondsInDay) - daysOffset 
		self.addMultiple(1,int(sd))
		self.addMultiple(0,int(ed+1-sd))
	}
	self.addMultiple(1,self.totalDays-len(self.Scale))

	// Choose start day. If one cant be found this means there
	// is no gap in the traveller's schedule, regardless of FLAP,
	// where the trip could be taken
	ts,err := self.choose()
	if (err != nil) {
		return 0,ENOSPACEFORTRIP
	}

	// Create airports
	fromAirport,err := fe.Airports.GetAirport(from)
	if (err != nil) {
		return 0,glog(err)
	}
	toAirport,err := fe.Airports.GetAirport(to)
	if (err != nil) {
		return 0,glog(err)
	}

	// Build trip flights. Note flight times do not need to be accurate for promises as long as the
	// start of first flight is earlier than the start of the first flight in the actual trip 
	// and the end of the last flight is later than the end of the last flight in the actual trip.
	var plannedflights [2]flap.Flight
	sds:=currentDay + flap.EpochTime(ts*flap.SecondsInDay)
	ede:=sds + flap.EpochTime(length*flap.SecondsInDay)
	f,err := flap.NewFlight(fromAirport,sds,toAirport,sds+1)
	if (err != nil) {
		return 0, glog(err)
	}
	plannedflights[0]=*f
	f,err = flap.NewFlight(fromAirport,ede-2,toAirport,ede-1)
	if (err != nil) {
		return 0, glog(err)
	}
	plannedflights[1]=*f

	// Obtain promise
	proposal,err := fe.Propose(pp,plannedflights[:],0,currentDay)
	if (err != nil) {
		return 0,glog(err)
	}
	return flap.Days(ts),fe.Make(pp,proposal)
}

