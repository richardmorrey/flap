package flap 

import (
	"gonum.org/v1/gonum/stat"
	"errors"
	"encoding/binary"
	"bytes"
)

var EMAXPOINTSBELOWTWO = errors.New("maxpoints must be two or more")

type smoothYs struct {
	windowSize int
	maxYs int
	ys [] float64
	window [] float64
}

// From implements db/Serialize
func (self *smoothYs) From(buff *bytes.Buffer) error {

	var fixedSize int32
	err := binary.Read(buff, binary.LittleEndian,&fixedSize)
	if err != nil {
		return err
	}
	self.windowSize = int(fixedSize)

	err = binary.Read(buff, binary.LittleEndian,&fixedSize)
	if err != nil {
		return err
	}
	self.maxYs=int(fixedSize)

	var n  int32
	err = binary.Read(buff, binary.LittleEndian,&n)
	if err != nil {
		return err
	}
	for i:=int32(0); i < n; i++ {
		var v float64
		err = binary.Read(buff, binary.LittleEndian,&v)
		if err != nil {
			return err
		}
		self.ys= append(self.ys,v)
	}

	n = int32(len(self.window))
	err = binary.Read(buff, binary.LittleEndian,&n)
	if err != nil {
		return err
	}
	for i:=int32(0); i < n; i++ {
		var v float64
		err = binary.Read(buff, binary.LittleEndian,&v)
		if err != nil {
			return err
		}
		self.window =append(self.window,v)
	}

	return err
}

// To implements db/Serialize
func (self *smoothYs) To(buff *bytes.Buffer) error {

	fixedSize := int32(self.windowSize)
	err := binary.Write(buff, binary.LittleEndian,&fixedSize)
	if err != nil {
		return err
	}

	fixedSize = int32(self.maxYs)
	err = binary.Write(buff, binary.LittleEndian,&fixedSize)
	if err != nil {
		return err
	}

	n := int32(len(self.ys))
	err = binary.Write(buff, binary.LittleEndian,&n)
	if err != nil {
		return err
	}
	for i:=int32(0); i < n; i++ {
		err = binary.Write(buff, binary.LittleEndian,&self.ys[i])
		if err != nil {
			return err
		}
	}

	n = int32(len(self.window))
	err = binary.Write(buff, binary.LittleEndian,&n)
	if err != nil {
		return err
	}
	for i:=int32(0); i < n; i++ {
		err = binary.Write(buff, binary.LittleEndian,&self.window[i])
		if err != nil {
			return err
		}
	}

	return err
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

