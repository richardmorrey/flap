package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"time"
	"errors"
)

var ENOVALIDPROBABILITYINBOTSPEC = errors.New("Invalid BotSpec configuration")

type yearProbs struct {
	days [366]Probability
}

// newYearProbs creates a fly probability for every day in the calendar
// year for given band of traveller bots
func newYearProbs(bs *BotSpec) (*yearProbs,error) {
	
	// Create instance
	yp := new(yearProbs)
	
	// Validate config
	if bs.FlyProbability == 0.0 || bs.FlyProbability > 1 {
		return nil,logError(ENOVALIDPROBABILITYINBOTSPEC,bs.FlyProbability)
	}
	if bs.MonthWeights != nil && len(bs.MonthWeights) != 12 {
		return nil,logError(ENOVALIDPROBABILITYINBOTSPEC,len(bs.MonthWeights))
	}

	// If there is a valid list of month probabilities use that
	if len(bs.MonthWeights) == 12 {
		
		// Calculate totals for the year
		var yearWeight Probability
		for yr := time.Date(2020, time.January, 1, 1, 0, 0, 0, time.UTC);yr.Year()==2020; yr=yr.Add(time.Hour*24) {
			yearWeight +=  Probability(bs.MonthWeights[yr.Month()-1])
		}

		// Set day probs for each day in a leap year
		for yr := time.Date(2020, time.January, 1, 1, 0, 0, 0, time.UTC);yr.Year()==2020; yr=yr.Add(time.Hour*24) {
			yp.days[yr.YearDay()-1]=(Probability(bs.MonthWeights[yr.Month()-1])/yearWeight)*bs.FlyProbability*366
		}
		logInfo("Set daily probabilities:",yp.days)
		return yp,nil
	}

	// Default to simple probability 
	for  i:=0; i < 366; i++ {
		yp.days[i]=bs.FlyProbability
	}
	return yp,nil
}

// isLeap returns true if the given time is in a leap year
func (self* yearProbs) isLeap(t time.Time) bool {
	yr:=t.Year()
	return yr%400 == 0 || yr%4 == 0 && yr%100 != 0
}
    
// Returns the fly probability for calendar day of given date
func (self* yearProbs) getDayProb(date flap.EpochTime) Probability {

	// Check for a leap year
	t := date.ToTime()

	// Calculate day of year, adjusting for non-leap years
	yd := t.YearDay()
	if self.isLeap(t) && yd > (31+28) {
		yd++
	}

	// Return probability of flying on that day
	return self.days[t.YearDay()-1]
}

