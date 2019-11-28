package flap 

import (
	//"errors"
	"math"
	//"sort"
	//"time"
)

type epochDays Days
func (self* epochDays) toEpochTime() EpochTime {
	return EpochTime(*self)*SecondsInDay
}
func (self * epochDays) fromEpochTime(et EpochTime) epochDays {
	*self = epochDays(et/SecondsInDay)
	return *self
}

type predictor interface
{
	add(Kilometres) (error)
	predict(Kilometres,epochDays) (epochDays,error)
}

// bestFit predicts dates when a specified distance balance would
// return to credit using a simple linear best fit against a plot
// of distance share against day.
type bestFit struct {
	ys	  	[]Kilometres
	m		float64
	c		float64
	day1		epochDays
}

// newBestFit constructs a new bestFit struct initialized with
// the current epoch time so that predictions can be returned
// in absolute time
func newBestFit(now EpochTime) *bestFit {
	bf := new(bestFit)
	bf.day1.fromEpochTime(now)
	return bf
}

// add adds a datapoint to the plot used for predictions. Must
// be called each day with the distance share credit to each
// account for backfilling that day.
func (self *bestFit) add(share Kilometres) {
	self.ys = append(self.ys,share)
	self.calculateLine()
}

// calculateLine calulates the m (gradient) and c (offset)  values for a y=mx+c
// line best fitting the scatter plot of backfill shares using
// simple linear best fit. For a good expanation of the algorithm see:
// https://www.statisticshowto.datasciencecentral.com/probability-and-statistics/regression-analysis/find-a-linear-regression-equation/
func (self *bestFit) calculateLine() {
	
	// Calculate sums
	var ySum float64
	var xSum float64
	var xxSum float64
	var xySum float64 
	for x,y := range self.ys {
		ySum += float64(y)
		xSum += float64(x)
		xxSum += float64(x)*float64(x)
		xySum += float64(x)*float64(y)
	}
	n := float64(len(self.ys))

	// Calculate gradient and offset
	self.c = ((ySum*xxSum) - (xSum*xySum)) / ((n*xxSum) - (xSum*xSum))
	self.m = ((n*xySum) - (xSum*ySum)) /  ((n*xxSum) - (xSum*xSum))
}

// predict estimates the date when the backfill of the given distance
// will complete with the given start date. Effectively sees the provided
// distance as an area under the y=x+mc graph - and works out and derives
// the unknown variable-- the length of the triangle -- from that. In other
// words solves below for endDay.
// balance = ((endDay-startDay) * ystartDay)/2
// or
// endDay = ((2*balance)/ystartDay) + startDay
func (self *bestFit) predict(balance Kilometres,startDay epochDays) epochDays {
	
	yStartDay := (self.m * float64(startDay)) + self.c
	return epochDays( math.Ceil(((2*float64(balance))/yStartDay)) + float64(startDay)) + self.day1
}

