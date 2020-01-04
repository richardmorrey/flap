package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"errors"
)

type botPromises struct {
	weights
	totalDays int
}

// newBotPromises creates a new botPromises with all
// days within allowed range equally likely to be
// chosen as start day for a trip
func newBotPromises(totalDays flap.Days) *botPromises {
	bp = new(botPromises)
	self.totalDays = totalDays
	return bp
} 

// getPromise chooses dates for a future trip that does not overlap with a
// trip for which a promise already exists. It then tries to obtain a promise
// for the new trip and if successful returns start day of  the trip for planning
// and otherwise an error
func (self* botPromises) getPromise(fe *flap.Engine,pp flap.Passport,currentDay flap.EpochTime,length flap.Days,from flap.ICAOCode, to flap.ICAOCode) flap.Day,error {
	
	// Retrieve traveller object
	t,err := fe.Travellers.GetTraveller(pp)
	if err != nil {
		return 0, glog(err)
	}

	// Build weights to cover all possible days for start of the trip
	// making sure that any day that is not suitable (is part of a planned trip 
	// or is too close to start of a planned trip) has zero weight
	self.reset()
	daysOffset:=currentDay.toEpochDays()
	it := t.Promises.NewIterator()
	var sd,ed flap.Days
	for it.Next() {
		sd = it.Value().TripStart.toEpochDays(false) - daysOffset - length
		ed = it.Value().TripEnd.toEpochDays(true) - daysOffset
		self.addMultiple(1,sd)
		self.addMultiple(0,ed+1-sd)
	}
	self.addMultiple(self.totalDays-len(self.weights),1)

	// Choose start day. If one cant be found this means there
	// is no gap in the traveller's schedule, regardless of FLAP,
	// where the trip could be taken
	ts,err = self.choose()
	if (err != nil) {
		return 0,ENOSPACEFORTRIP
	}

	// Build trip flights. Note flight times
	// do not need to be accurate for promises as long as the
	// start of first flight is earlier than the start of the first flight in the actual trip 
	// and the end of the last flight is later than the end of the last flight in the actual trip.
	var plannedflights [2]Flight
	start=currentDay + flap.EpochTime(ts*flap.SecondsInDay)
	end:=start + flap.EpochTime((length*flap.SecondsInDay)
	f,err := flap.NewFlight(from,start,start+1)
	if (err != nil) {
		return 0, glog(err)
	}
	plannedFlights[0]=*f
	end=
	f,err := flap.NewFlight(from,end-2,end-1)
	if (err != nil) {
		return 0, glog(err)
	}
	plannedFlights[1]=*f

	// Obtain promise
	proposal,err := engine.Propose(pp,plannedflights,0,SecondsInDay)
	if (err != nil) {
		return 0,glog(err)
	}

	// Return result
}

