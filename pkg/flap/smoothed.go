package flap 

import (
	"gonum.org/v1/gonum/stat"
	"errors"
)

var EMAXPOINTSBELOWTWO = errors.New("maxpoints must be two or more")

type smoothYs struct {
	windowSize int
	maxYs int
	ys [] float64
	window [] float64
}

// Adds a value to the smoothing window, and calculates and adds the new value
func (self *smoothYs) addY(v float64) error {

	// Add to smoothing window
	if len(self.window) == self.windowSize {
		self.window= self.window[1:]
	}
	self.window = append(self.window, v)

	// Add average as new point
	if len(self.ys) == self.maxYs {
		self.ys= self.ys[1:]
	}
	self.ys = append(self.ys,stat.Mean(self.window,nil))
	return nil
}

// Set window sizes
func (self *smoothYs) setWindows(maxYs int, windowSize int) error {

	// Validate maxpoints
	if maxYs < 2 {
		return EMAXPOINTSBELOWTWO
	}
	
	// Set values
	self.maxYs=maxYs
	if windowSize < 1 {
		self.windowSize=1
	} else {
		self.windowSize = windowSize
	}
	return nil
}

