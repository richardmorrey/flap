package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	//"time"
	//"errors"
)

type yearProbs struct {
	days [366]Probability
}

// newYearProbs creates a fly probability for every day in the calendar
// year for given band of traveller bots
func newYearProbs(bs *BotSpec) *yearProbs {
	yp := new(yearProbs)
	for  i:=0; i < 366; i++ {
		yp.days[i]=bs.PlanProbability
	}
	return yp
}

// Returns the fly probability for calenday of given date
func (self* yearProbs) getDayProb(date flap.EpochTime) Probability {

	// Calculate day of year, adjusting for non-leap years
	t := date.ToTime()

	// Return probability of flying on that day
	return self.days[t.YearDay()-1]
}

