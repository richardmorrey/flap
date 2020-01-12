package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"errors"
)

var ENOSPACEFORTRIP = errors.New("No space for trip")
var ETOOMANYDAYSTOCHOOSEFROM = errors.New("Too many days to choose from")

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
func (self* botPromises) getPromise(fe *flap.Engine,pp flap.Passport,currentDay flap.EpochTime,length flap.Days,from flap.ICAOCode, to flap.ICAOCode,deterministic bool) (flap.Days,error) {
	
	// Build weights to cover all possible days for start of the trip
	// making sure that any day that is not suitable (is part of a planned trip 
	// or is too close to start of a planned trip) has zero weight
	self.reset()
	t,err := fe.Travellers.GetTraveller(pp)
	daysToChooseFrom := (self.totalDays-int(length-1))
	uptoDay:=flap.Days(currentDay/flap.SecondsInDay)
	lastDay:=uptoDay + flap.Days(daysToChooseFrom)
	if err == nil {
		it := t.Promises.NewIterator()
		var sd,ed flap.Days
		for it.Next() {

			// Add days when trip can start - up until the start of the current 
			// promise trip, but allowing for the trip length
			sd = flap.Days(it.Value().TripStart/flap.SecondsInDay) - (length-1)
			if sd > uptoDay {
				self.addMultiple(1,int(sd-uptoDay))
				uptoDay=sd
			}
			
			// Add days when trip cant start - up until the end of the promise trip
			ed = flap.Days(it.Value().TripEnd/flap.SecondsInDay) + 1
			if ed >=uptoDay {
				self.addMultiple(0,int(ed-uptoDay))
				uptoDay=ed
			}
		}
	}

	// Add remaining days
	self.addMultiple(1,int(lastDay-uptoDay))
	if len(self.Scale) !=  daysToChooseFrom {
		return 0, logError(ETOOMANYDAYSTOCHOOSEFROM)
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
	logDebug("ts=",ts,"weights=",self.Scale)
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
	sds:=currentDay + flap.EpochTime(ts*flap.SecondsInDay)
	ede:=sds + flap.EpochTime(length*flap.SecondsInDay)
	f,err := flap.NewFlight(fromAirport,sds+1,toAirport,sds+2)
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
	proposal,err := fe.Propose(pp,plannedflights[:],0,currentDay)
	if (err != nil) {
		return 0,logError(err)
	}
	err = logError(fe.Make(pp,proposal))
	if err == nil {
		logDebug("Made promise for trip in ",ts," days")
	}
	return flap.Days(ts),err
}

