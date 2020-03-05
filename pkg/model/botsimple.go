package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"math/rand"
)

type botPlanner interface
{
	clone() botPlanner
	build(*BotSpec) error
	areWePlanning(*flap.Engine,flap.Passport,flap.EpochTime,flap.Days) bool
	canWePlan(*flap.Engine,flap.Passport,flap.EpochTime,flap.ICAOCode,flap.ICAOCode,flap.Days) (flap.Days,error)
}

type botSimple struct {
	probs *yearProbs
}

// makes a clone of given planner (not a deep copy)
// but a functionally compatible version that can
// be used simultaneously
func (self *botSimple) clone() botPlanner {
	clone := new(botSimple)
	clone.probs = self.probs
	return clone
}

// build initializes planner
func (self *botSimple) build(bs *BotSpec) error {

	// build year of daily probabilities
	var err error
	self.probs,err = newYearProbs(bs)
	if (err != nil) {
		return logError(err)
	}
	return nil
} 

// areWePlanning returns true if we want to plan a trip today
func (self *botSimple) areWePlanning(fe *flap.Engine,pp flap.Passport, currentDay flap.EpochTime, tripLength flap.Days) bool {

	// Confirm not mid-trip
	t,err := fe.Travellers.GetTraveller(pp)
	if  (err == nil) && t.MidTrip() {
		return false
	}

	// Decide whether to fly
	dice:=Probability(rand.Float64())
	if dice <= self.probs.getDayProb(currentDay) {
		logDebug("Travelling today", pp.ToString())
		return true
	}
	return false
}

// canWePlan confirms whether traveller can plan the proposed trip on the given day
func (self *botSimple) canWePlan(fe *flap.Engine,pp flap.Passport,now flap.EpochTime,from flap.ICAOCode,to flap.ICAOCode,tripLength flap.Days) (flap.Days,error) {
	return 0,nil
}
