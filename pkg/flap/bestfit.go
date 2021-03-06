package flap 

import (
	"errors"
	"math"
	"github.com/richardmorrey/flap/pkg/db"
	"encoding/binary"
	"bytes"
)

var ENOVALIDPREDICTION = errors.New("No valid prediction")
var ENOTENOUGHDATAPOINTS = errors.New("No data points")
var EXORIGINZERO = errors.New("Xorigin can't be zero")

type epochDays Days
func (self epochDays) toEpochTime() EpochTime {
	return EpochTime(self)*SecondsInDay
}

type predictVersion uint64
type predictor interface
{
	add(epochDays,Kilometres)
	predict(Kilometres,epochDays) (epochDays,error)
	version() predictVersion
	backfilled(epochDays,epochDays) (Kilometres,error)
	state() ([]float64,[]float64,error)
	db.Serialize
}

// bestFit predicts dates when a specified distance balance would
// return to credit using a simple linear best fit against a plot
// of distance share against day.
type bestFit struct {
	SmoothYs
	m		float64
	c		float64
	pv		predictVersion
}

// newBestFit constructs a new bestFit struct initialized with
// the current epoch time so that predictions can be returned
// in absolute time
func newBestFit(cfg PromisesConfig) (*bestFit,error) {

	// Create object
	bf := new(bestFit)
	bf.c = -1 // indicates uninitializated state as line cant have -ve values

	// Initialize smoothing window
	err := bf.SetWindows(int(cfg.MaxPoints),int(cfg.SmoothWindow))
	if err != nil {
		return nil,logError(err)
	}
	return bf,nil
}

// To implemented as part of db/Serialize
func (self *bestFit) To(buff *bytes.Buffer) error {

	err := self.SmoothYs.To(buff)
	if (err != nil) {
		return err
	}

	err = binary.Write(buff,binary.LittleEndian,&self.m)
	if err != nil {
		return logError(err)
	}

	err = binary.Write(buff,binary.LittleEndian,&self.c)
	if err != nil {
		return logError(err)
	}

	return binary.Write(buff,binary.LittleEndian,&self.pv)
}

// From implemented as part of db/Serialize
func (self *bestFit) From(buff *bytes.Buffer) error {

	err := self.SmoothYs.From(buff)
	if (err != nil) {
		return err
	}

	err = binary.Read(buff,binary.LittleEndian,&self.m)
	if err != nil {
		return logError(err)
	}

	err = binary.Read(buff,binary.LittleEndian,&self.c)
	if err != nil {
		return logError(err)
	}

	return binary.Read(buff,binary.LittleEndian,&self.pv)

}

// state returns set of points lasted used for regression and the regression results
// as a list of consts in ascending degree
func (self* bestFit) state() ([]float64,[]float64,error) {
	if len(self.ys) < 2 {
		return nil,nil,ENOTENOUGHDATAPOINTS
	}
	return self.ys,[]float64{self.c,self.m},nil
}

// version returns number indicating current version of the best fit line.
// number is incremented each time value of m or c changes.
func (self *bestFit) version() predictVersion {
	return self.pv
}

// add adds a datapoint to the plot used for predictions. Must
// be called each day with the distance share credit to each
// account for backfilling that day.
func (self *bestFit) add(x epochDays, y Kilometres) {
	self.AddY(float64(y))
	self.calculateLine(x)
}

// calculateLine calulates the m (gradient) and c (offset)  values for a y=mx+c
// line best fitting the scatter plot of backfill shares using
// simple linear best fit. For a good expanation of the algorithm see:
// https://www.statisticshowto.datasciencecentral.com/probability-and-statistics/regression-analysis/find-a-linear-regression-equation/
func (self *bestFit) calculateLine(xmax epochDays) error {

	// Check for data
	if len(self.ys) < 2 {
		return ENOTENOUGHDATAPOINTS
	}
	
	// Calculate sums
	var ySum float64
	var xSum float64
	var xxSum float64
	var xySum float64 
	xorigin := xmax - epochDays(len(self.ys)-1) 
	for x,y := range self.ys {
		realx:= float64(xorigin)+float64(x)
		ySum += y
		xSum += realx
		xxSum += realx*realx
		xySum += realx*y
	}
	n := float64(len(self.ys))

	// Calculate gradient and offset
	c := ((ySum*xxSum) - (xSum*xySum)) / ((n*xxSum) - (xSum*xSum))
	m := ((n*xySum) - (xSum*ySum)) /  ((n*xxSum) - (xSum*xSum))
	if (self.c != c || self.m != m) {
		self.pv++
		self.c=c
		self.m=m
	}
	logInfo("m=",self.m,"c=",self.c)
	return nil
}

// calcY returns y value of current best fit line for given x value
func (self *bestFit) calcY(x float64) float64 {
	return x*self.m + self.c
}

// predict estimates the date when the backfill of the given distance
// will complete with the given start date. Effectively sees the provided
// distance as an area under the y=mx+c graph, and works out the unknown
// variable - the end day - using simple calculus and the quadratic formula.
// In detail:
// 1) The integral of mx+c is (m/2)(x^2) + cx
// 2) The balance to be backfilled is integral(end) - integral(start)
// 3) Start is known (day after end of trip) so integral(start) can be calculated
// 4) We can then derive the following quadratric equation that can be solved for 
//    end using the quadratic formuka.
//    (m/2)(end^2) +  (c*end) - (balance+integral(start))
// 5) The quadratic formula provides two values. We choose the lowest one that
//    is greater then start and has a positive y value as the prediction. If
//    neither fit those criteria we return error that prediction cannot be made
func (self *bestFit) predict(balance Kilometres,start epochDays) (epochDays,error) {

	// Check for valid state
	if self.c < 0 {
		return epochDays(math.MaxInt64),ENOTENOUGHDATAPOINTS
	}

	// Calulate integral of start
	is := self.integral(float64(start))

	// Solve quadratic
	ends,_ := qr(self.m/2,self.c,-(float64(balance)+is))

	// Choose an answer 
	choice := math.MaxFloat64
	for _,candidate := range ends {
		if candidate > float64(start) && candidate < choice {
			if self.calcY(candidate) > 0.0 {
				choice=candidate
			}
		}
	}

	// Calculate an estimate based on 0 gradient and last point
	// if no valid choice from the formula
	if choice == math.MaxFloat64 {
		l := len(self.ys)
		if  l >  0 {
			choice = float64(start) + (float64(balance)/self.ys[l-1])
		}
	}
	
	// Return choice if we have made one or otherwise return
	// an answer assuming horizontal line.
	if choice == math.MaxFloat64 {
		return epochDays(choice), ENOVALIDPREDICTION
	} else {
		return epochDays(math.Ceil(choice)), nil
	}
}

// backfilled predicts distance that would be backfilled for a single traveller between
// the two given days
func (self* bestFit) backfilled(start epochDays,end epochDays) (Kilometres,error) {

	// Check for valid state
	if self.c < 0 {
		return 0,ENOTENOUGHDATAPOINTS
	}

	// Assumme a horizontal line if it dips under zero
	// during the given period ...
	d1 := float64(start)
	d2 := float64(end)
	if self.calcY(d1) < 0  || self.calcY(d2) <0 {
		logDebug("using horizontal line")
		return Kilometres(end-start) * Kilometres(self.ys[len(self.ys)-1]),nil
	}

	// ... otherwise return difference between the integrals
	// for the given dates
	return Kilometres(self.integral(d2)-self.integral(d1)),nil
}

// integral gives the integral for a given day
func (self *bestFit) integral(d float64) float64 {
	return ((self.m/2)*d*d) + self.c*d
}

// From http://www.rosettacode.org/wiki/Roots_of_a_quadratic_function#Go
func qr(a, b, c float64) ([]float64, []complex128) {
	d := b*b-4*a*c
   switch {
   case d == 0:
       // single root
       return []float64{-b/(2*a)}, nil
   case d > 0:
       // two real roots
       if b < 0 {
           d = math.Sqrt(d)-b
       } else {
           d = -math.Sqrt(d)-b
       }
       return []float64{d/(2*a), (2*c)/d}, nil
   case d < 0:
       // two complex roots
       den := 1/(2*a)
       t1 := complex(-b*den, 0)
       t2 := complex(0, math.Sqrt(-d)*den)
       return nil, []complex128{t1+t2, t1-t2}
   }
   // otherwise d overflowed or a coefficient was NAN
   return nil, nil
}
