package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"errors"
	"math"
	"math/rand"
	"fmt"
)

var ENOSPACEFORTRIP = errors.New("No space for trip")
var ETOOMANYDAYSTOCHOOSEFROM = errors.New("Too many days to choose from")
var ENOTPLANNINGTODAY = errors.New("Not planning today")

type promisesPlanner struct {
	simplePlanner
	Weights
	totalDays  	 flap.Days
	chosenWeight	 weight
	planProbability  Probability
	planDay		 flap.EpochTime
}

// makes a clone of given planner (not a deep copy)
// but a functionally compatible version that can
// be used simultaneously
func (self *promisesPlanner) clone() botPlanner {
	clone := new(promisesPlanner)
	clone.probs = self.probs
	clone.totalDays =  self.totalDays
	return clone
}

// build stores the maximum number of days ahead for promises to use
// when choosing whether and when to choose to fly
func (self *promisesPlanner) build(bs BotSpec, fp flap.FlapParams, mp ModelParams) error {
	self.totalDays = fp.Promises.MaxDays
	return self.simplePlanner.build(bs,fp,mp)
} 

const TENPOWERNINE=Probability(1000*1000*1000)
const NOTPLANNING=math.MaxInt32
// calcDayProb derives the probability of a flight starting on given day
// by dividing the probability of flying on a given calendar year by
// the number of days to choose from for the current planning decision
// Note a factor is applied to make the prob a valid (integer) weight
func (self* promisesPlanner) addDayProb(cd flap.Days,totalDays flap.Days,dayOfModel flap.Days) {
	self.addIndexWeight(int(cd),
		weight(self.probs.getDayProb(flap.EpochTime(cd*flap.SecondsInDay),dayOfModel)*TENPOWERNINE/Probability(totalDays)))
}

// prepareWeights builds scale of weights covering all days when
// traveller is allowed to plan leaving out days already occupied
// by planned trips. Note: indices for weights are in whole days
// since start of epoch time; backfilling days are still included to
// allow for promises stacking in proposal.
func (self* promisesPlanner) prepareWeights(fe *flap.Engine,pp flap.Passport,currentDay flap.Days,length flap.Days,dayOfModel flap.Days) error {

	// Reset state
	self.reset()
	cd := currentDay
	availableDays := self.totalDays-length
	to := cd + availableDays

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
					self.addDayProb(cd,availableDays,dayOfModel)
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
		self.addDayProb(cd,availableDays,dayOfModel)
	}
	
	// Add a final weight for "not planning" to make the total probability up to 1
	tw,err := self.topWeight()
	if err != nil {
		tw = 0
	}
	self.addIndexWeight(NOTPLANNING, weight(TENPOWERNINE)-tw)
	//logDebug("flyprob=",tw,"notflyprob=",weight(TENPOWERNINE)-tw)
	return nil
}

// areWePlanning makes a decision on whether we are planning today based on the calendar day probabilities
// for the planning period. It is only an approximation as, for speed reasons, the current plans of the
// traveller are not considered. Note in case of a decsion to (probably) fly the dice roll is cached
// to use when actually planning.
func (self* promisesPlanner) areWePlanning(fe *flap.Engine,pp flap.Passport,now flap.EpochTime,length flap.Days, dayOfModel flap.Days) bool {

	// Calculate probability of planning today if we need to
	if self.planDay != now {
		self.planProbability = 0
		self.planDay = now
		availableDays := self.totalDays-length
		lastDay := now + (flap.EpochTime(availableDays)*flap.SecondsInDay)
		for cd := now; cd <  lastDay; cd+=flap.SecondsInDay {
			self.planProbability += self.probs.getDayProb(cd,dayOfModel)/Probability(availableDays)
		}
	}

	// Roll the dice and store it as a weight if we are planning
	dice:=Probability(rand.Float64())
	if dice <= self.planProbability {
		self.chosenWeight= weight(dice*TENPOWERNINE)
		return true
	}
	self.chosenWeight=0
	return false
}

// whenWillWeFly tries to obtain a promise
// for the new trip and if successful returns start of the day of the trip for planning
// and otherwise an error
func (self *promisesPlanner) whenWillWeFly(fe *flap.Engine,pp flap.Passport,now flap.EpochTime,from flap.ICAOCode,to flap.ICAOCode,length flap.Days,dayOfModel flap.Days) (flap.EpochTime,error) {

	// Make sure we have chosen to plan
	if self.chosenWeight==0 {
		return 0, logError(ENOTPLANNINGTODAY)
	}

	// Build weights to use to choose trip start day
	nowInDays := flap.Days(now/flap.SecondsInDay)
	err := self.prepareWeights(fe,pp,nowInDays,length,dayOfModel)
	if err != nil {
		return 0,logError(err)
	}

	// Attempt to choose start day.
	ts,err := self.find(self.chosenWeight)
	if err != nil {
		return 0, logError(err)
	}

	// If the top weight (indicated we are not planning) has
	// been chosen then return
	if (ts == NOTPLANNING) {
		logDebug("Not planning after all")
		return 0,ENOTPLANNINGTODAY
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
	epochStartDay := flap.Days(ts)
	sds:=flap.EpochTime(epochStartDay*flap.SecondsInDay)
	ede:=sds + flap.EpochTime(length*flap.SecondsInDay)
	f,err := flap.NewFlight(fromAirport,sds,toAirport,sds+1)
	if (err != nil) {
		return 0,logError(err)
	}
	plannedflights[0]=*f
	f,err = flap.NewFlight(toAirport,ede+(flap.SecondsInDay-2),fromAirport,(ede+flap.SecondsInDay-1))
	if (err != nil) {
		return 0,logError(err)
	}
	plannedflights[1]=*f
	logDebug("plannedflights:",plannedflights)

	// Obtain promise
	proposal,err := fe.Propose(pp,plannedflights[:],0,now)
	if (err != nil) {
		logError(err)
		return 0,ENOSPACEFORTRIP
	}

	// Make promise
	err = fe.Make(pp,proposal,now)
	if err == nil {
		logline := fmt.Sprintf("Made promise for %s for trip from %s to %s leaving %s returning %s", pp.ToString(),fromAirport.Code.ToString(), toAirport.Code.ToString(), sds.ToTime(), ede.ToTime())
		logDebug(logline)
	} else {
		return 0, logError(err)
	}
	return sds,err
}

