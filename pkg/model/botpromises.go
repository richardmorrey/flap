package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"errors"
)

var ENOSPACEFORTRIP = errors.New("No space for trip")
var ETOOMANYDAYSTOCHOOSEFROM = errors.New("Too many days to choose from")

type botPromises struct {
	Weights
	totalDays flap.Days
}

// newBotPromises creates a new botPromises with all
// days within allowed range equally likely to be
// chosen as start day for a trip
func newBotPromises(totalDays flap.Days) *botPromises {
	bp := new(botPromises)
	bp.totalDays = totalDays
	return bp
} 

// buildWeights builds scale of weights covering all days when
// traveller is allowed to plan leaving out days already occupied
// by planned trips. Note: indices for weights are in whole days
// since start of epoch time; backfilling days are still included to
// allow for promises stacking in proposal.
func (self* botPromises) buildWeights(fe *flap.Engine,pp flap.Passport,currentDay flap.Days,length flap.Days) error {

	// Reset state
	self.reset()
	cd := currentDay
	to := cd + (self.totalDays-length)

	// Get traveller record
	t,err := fe.Travellers.GetTraveller(pp)
	if err == nil {

		// Add plan days not already taken by trips in made promises
		it := t.Promises.NewIterator()
		for it.Next() {

			// add days up until the start of the trip in this promise
			sp := flap.Days(it.Value().TripStart/flap.SecondsInDay) - (length-1)
			if sp > cd {
				for ; sp > cd ; cd++ {
					self.addIndexWeight(int(cd),weight(1))
				}
			}

			// Skip days within the trip in this promise if they spread
			// beyond the current day
			ep := flap.Days(it.Value().TripEnd/flap.SecondsInDay) + 1
			if ep > cd {
				cd = ep
			}
		}
	}

	// Add any remaining days to fill whole planning period
	for  ; cd <= to ; cd ++ {
		self.addIndexWeight(int(cd),weight(1))
	}
	return nil
}

// getPromise chooses dates for a future trip that does not overlap with a
// trip for which a promise already exists. It then tries to obtain a promise
// for the new trip and if successful returns start day of  the trip for planning
// and otherwise an error
func (self* botPromises) getPromise(fe *flap.Engine,pp flap.Passport,now flap.EpochTime,length flap.Days,from flap.ICAOCode, to flap.ICAOCode,deterministic bool) (flap.Days,error) {
	

	// Build weights to use to choose trip start day
	nowInDays := flap.Days(now/flap.SecondsInDay)
	err := self.buildWeights(fe,pp,nowInDays,length)
	if err != nil {
		return 0, logError(err)
	}

	// Choose start day. If one cant be found this means there
	// is no gap in the traveller's schedule, regardless of FLAP,
	// where the trip could be taken
	var ts int 
	if deterministic {
		ts,err = self.choosedeterministic()
	} else {
		ts,err = self.choose()
	}
	logDebug("trip start day =",ts)
	if (err != nil) {
		return 0,logError(ENOSPACEFORTRIP)
	}

	// Create airports
	fromAirport,err := fe.Airports.GetAirport(from)
	if (err != nil) {
		return 0,logError(err)
	}
	toAirport,err := fe.Airports.GetAirport(to)
	if (err != nil) {
		return 0,logError(err)
	}

	// Build trip flights. Note flight times do not need to be accurate for promises as long as the
	// start of first flight is earlier than the start of the first flight in the actual trip 
	// and the end of the last flight is later than the end of the last flight in the actual trip.
	var plannedflights [2]flap.Flight
	sds:=flap.EpochTime(ts*flap.SecondsInDay)
	ede:=sds + flap.EpochTime(length*flap.SecondsInDay)
	f,err := flap.NewFlight(fromAirport,sds,toAirport,sds+1)
	if (err != nil) {
		return 0, logError(err)
	}
	plannedflights[0]=*f
	f,err = flap.NewFlight(toAirport,ede-2,fromAirport,ede-1)
	if (err != nil) {
		return 0, logError(err)
	}
	plannedflights[1]=*f
	logDebug("plannedflights:",plannedflights)

	// Obtain promise
	proposal,err := fe.Propose(pp,plannedflights[:],0,now)
	if (err != nil) {
		logDebug("Propose failed with error",err)
		return 0,err
	}
	err = fe.Make(pp,proposal)
	if err == nil {
		logDebug("Made promise for trip on Day ",ts)
	} else {
		logInfo("Failed to make promises ", err)
	}
	return flap.Days(ts)-nowInDays,err
}

