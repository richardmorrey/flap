package flap 

import (
	"errors"
	"math"
	//"fmt"
)

var ENOVALIDPREDICTION = errors.New("No valid prediction")
var ENOTENOUGHDATAPOINTS = errors.New("No data points")
var EXORIGINZERO = errors.New("Xorigin can't be zero")
var EMAXPOINTSBELOWTWO = errors.New("maxpoints must be two or more")

type epochDays Days
func (self* epochDays) toEpochTime() EpochTime {
	return EpochTime(*self)*SecondsInDay
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
	xorigin		epochDays
	maxpoints	int
}

// newBestFit constructs a new bestFit struct initialized with
// the current epoch time so that predictions can be returned
// in absolute time
func newBestFit(now EpochTime,maxpoints int) (*bestFit,error) {

	// Enforce an xorigin greater than zero so we dont
	// need to worry about negative integrals
	var xorigin = now.toEpochDays(false)
	if xorigin == 0 {
		return nil,EXORIGINZERO
	}

	// Validate maxpoints
	if maxpoints < 2 {
		return nil, EMAXPOINTSBELOWTWO
	}

	// Create object
	bf := new(bestFit)
	bf.xorigin = xorigin
	bf.maxpoints=maxpoints
	bf.c = -1 // indicates uninitializated state as line cant have -ve values
	return bf,nil
}

// add adds a datapoint to the plot used for predictions. Must
// be called each day with the distance share credit to each
// account for backfilling that day.
func (self *bestFit) add(share Kilometres) {
	if len(self.ys) == self.maxpoints {
		self.ys= self.ys[1:]
	}
	self.ys = append(self.ys,share)
	self.calculateLine()
}

// calculateLine calulates the m (gradient) and c (offset)  values for a y=mx+c
// line best fitting the scatter plot of backfill shares using
// simple linear best fit. For a good expanation of the algorithm see:
// https://www.statisticshowto.datasciencecentral.com/probability-and-statistics/regression-analysis/find-a-linear-regression-equation/
func (self *bestFit) calculateLine() error {

	// Check for data
	if len(self.ys) < 2 {
		return ENOTENOUGHDATAPOINTS
	}
	
	// Calculate sums
	var ySum float64
	var xSum float64
	var xxSum float64
	var xySum float64 
	for x,y := range self.ys {
		realx:= float64(x)+float64(self.xorigin)
		ySum += float64(y)
		xSum += realx
		xxSum += realx*realx
		xySum += realx*float64(y)
	}
	n := float64(len(self.ys))

	// Calculate gradient and offset
	self.c = ((ySum*xxSum) - (xSum*xySum)) / ((n*xxSum) - (xSum*xSum))
	self.m = ((n*xySum) - (xSum*ySum)) /  ((n*xxSum) - (xSum*xSum))
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
		return epochDays(math.MaxInt64),ENOVALIDPREDICTION
	}

	// Calulate integral of start
	is := self.integral(float64(start))

	// Solve quadratic
	//fmt.Printf("a=%f,b=%f,c=%f\n",(self.m/2)*self.c,self.c,-(float64(balance)+is))
	ends,_ := qr(self.m/2,self.c,-(float64(balance)+is))
	//fmt.Printf("ends: %v\n", ends)

	// Choose an answer and return
	choice := math.MaxFloat64
	for _,candidate := range ends {
		//fmt.Printf("considering %f with integral %f\n",candidate, self.integral(candidate))
		if candidate > float64(start) && candidate < choice {
			if self.calcY(candidate) > 0.0 {
				choice=candidate
			}
		}
	}
	if choice == math.MaxFloat64 {
		return epochDays(choice),ENOVALIDPREDICTION
	} else {
		return epochDays(math.Ceil(choice)), nil
	}
}
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
