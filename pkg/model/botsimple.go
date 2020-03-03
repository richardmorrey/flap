package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"math/rand"
)

type botPlanner interface
{
	build(*BotSpec) error
	areWePlanning(*flap.Engine,flap.Passport,flap.EpochTime,flap.Days) bool
	canWePlan(*flap.Engine,flap.Passport,flap.EpochTime,flap.ICAOCode,flap.ICAOCode,flap.Days) (flap.Days,error)
}

type botSimple struct {
	probs *yearProbs
}

// build initializes planner
func (self botSimple) build(bs *BotSpec) error {

	// build year of daily probabilities
	var err error
	self.probs,err = newYearProbs(bs)
	if (err != nil) {
		return logError(err)
	}
	return nil
} 

// areWePlanning returns true if we want to plan a trip today
func (self botSimple) areWePlanning(fe *flap.Engine,pp flap.Passport, currentDay flap.EpochTime, tripLength flap.Days) bool {

	// Confirm not mid-trip
	t,err := fe.Travellers.GetTraveller(pp)
	if  (err == nil) && t.MidTrip() {
		return false
	}

	// Decide whether to fly
	dice:=Probability(rand.Float64())
	if dice <= self.probs.getDayProb(currentDay) {
		return true
	}
	return false
}

// canWePlan confirms whether traveller can plan the proposed trip on the given day
func (self botSimple) canWePlan(fe *flap.Engine,pp flap.Passport,now flap.EpochTime,from flap.ICAOCode,to flap.ICAOCode,tripLength flap.Days) (flap.Days,error) {
	return 0,nil
}
