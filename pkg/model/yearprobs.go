package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"errors"
)

type dayProbs float64
type yearProbs struct {
	days [366]dayProb
}

// newYearProbs creates a fly probability for every day in the calendar
// year for given band of traveller bots
func newYearProbs(bs *BotSpec) {
	for  i=0; i < 366; i++ {
		days[i]=bs.PlanProbability
	}
}

// Returns the fly probability for calenday of given date
func (self* yearProbs) getDayProb(date flap.EpochTime) {
	return self.days[time.Unix(date).YearDay()-1]
}

