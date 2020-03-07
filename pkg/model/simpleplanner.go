package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"math/rand"
)

type botPlanner interface
{
	clone() botPlanner
	build(BotSpec,flap.FlapParams) error
	areWePlanning(*flap.Engine,flap.Passport,flap.EpochTime,flap.Days) int
	canWePlan(*flap.Engine,flap.Passport,flap.EpochTime,flap.ICAOCode,flap.ICAOCode,flap.Days,flap.Days) error
}

type simplePlanner struct {
	probs *yearProbs
}

// makes a clone of given planner (not a deep copy)
// but a functionally compatible version that can
// be used simultaneously
func (self *simplePlanner) clone() botPlanner {
	clone := new(simplePlanner)
	clone.probs = self.probs
	return clone
}

// build initializes planner
func (self *simplePlanner) build(bs BotSpec,fp flap.FlapParams) error {

	// build year of daily probabilities
	var err error
	self.probs,err = newYearProbs(&bs)
	if (err != nil) {
		return logError(err)
	}
	return nil
} 

// areWePlanning returns 0-indexed day offset from today to plan to fly or
// -1 if we are not planning today
func (self *simplePlanner) areWePlanning(fe *flap.Engine,pp flap.Passport, currentDay flap.EpochTime, tripLength flap.Days) int {

	// Confirm not mid-trip
	t,err := fe.Travellers.GetTraveller(pp)
	if  (err == nil) && t.MidTrip() {
		return -1
	}

	// Decide whether to fly
	dice:=Probability(rand.Float64())
	if dice <= self.probs.getDayProb(currentDay) {
		logDebug("Travelling today", pp.ToString())
		return 0
	}
	return -1
}

// canWePlan confirms whether traveller can plan the proposed trip on the given day
func (self *simplePlanner) canWePlan(fe *flap.Engine,pp flap.Passport,now flap.EpochTime,from flap.ICAOCode,to flap.ICAOCode,tripLength flap.Days, ts flap.Days) (error) {
	return nil
}
