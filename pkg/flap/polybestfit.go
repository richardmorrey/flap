package flap 

import (
	"encoding/binary"
	"bytes"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/floats"
	"math"
	"reflect"
	//"fmt"
)

// polyBestFit predicts dates when a specified distance balance would
// return to credit using a polynomial linear best fit against a plot
// of distance share against day.
type polyBestFit struct {
	SmoothYs
	pv		predictVersion
	consts		[]float64
	degree		int
}

// To implemented as part of db/Serialize
func (self *polyBestFit) To(buff *bytes.Buffer) error {

	err := self.SmoothYs.To(buff)
	if (err != nil) {
		return err
	}

	err = binary.Write(buff,binary.LittleEndian,&self.pv)
	if err != nil {
		return logError(err)
	}

	n := int32(len(self.consts))
	err = binary.Write(buff, binary.LittleEndian,&n)
	if err != nil {
		return err
	}
	for i:=int32(0); i < n; i++ {
		err = binary.Write(buff, binary.LittleEndian,&self.consts[i])
		if err != nil {
			return err
		}
	}

	fixedLength := uint32(self.degree)
	return binary.Write(buff,binary.LittleEndian,&fixedLength)
}

// From implemented as part of db/Serialize
func (self *polyBestFit) From(buff *bytes.Buffer) error {

	err := self.SmoothYs.From(buff)
	if (err != nil) {
		return err
	}

	err = binary.Read(buff,binary.LittleEndian,&self.pv)
	if err != nil {
		return logError(err)
	}

	var n int32
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
		self.consts= append(self.consts,v)
	}

	var fixedLength uint32
	err = binary.Read(buff,binary.LittleEndian,&fixedLength)
	if (err == nil) {
		self.degree = int(fixedLength)
	}
	return err
}

// newBestFit constructs a new bestFit struct initialized with
// the current epoch time so that predictions can be returned
// in absolute time
func newPolyBestFit(cfg PromisesConfig) (*polyBestFit,error) {

	// Create object
	bf := new(polyBestFit)
	bf.degree = int(cfg.Degree)
	bf.consts = make([]float64,0,bf.degree+1)
	
	// Initialize smoothing window
	return bf,bf.SetWindows(int(cfg.MaxPoints),int(cfg.SmoothWindow))
}

// returns current input of to regression and results of regression as
// list of consts in ascending degree
func (self* polyBestFit) state() ([]float64,[]float64,error) {
	if len(self.consts) == 0 {
		return nil,nil,ENOTENOUGHDATAPOINTS
	}
	return self.ys,self.consts,nil
}

// adds a point to the graph and recalculates polynomial best fit
// using configured degree. Based on https://rosettacode.org/wiki/Polynomial_regression#Go
func (self* polyBestFit) add(today epochDays,y Kilometres) {

    // Add new y value
    self.AddY(float64(y))

    if len(self.ys) > self.degree {

	// Build X and Y matrices
	a := self.Vandermonde(float64(today) - float64(epochDays(len(self.ys)-1)))
	b := mat.NewDense(len(self.ys), 1, self.ys)

	// Calculate constants
	c := mat.NewDense(self.degree+1, 1, nil)
	qr := new(mat.QR)
	qr.Factorize(a)
	err := qr.SolveTo(c, false, b)
	if err != nil {
		logError(err)
	}

	// Extract results
	newConsts := make([]float64,0,self.degree+1)
	for j:=0; j < self.degree+1; j++ {
		newConsts = append(newConsts,floats.Round(c.At(j,0),10))
	}
	if !reflect.DeepEqual(newConsts,self.consts) {
		self.pv++
		self.consts = newConsts
	}
    }
}
func (self* polyBestFit) Vandermonde(xorigin float64) *mat.Dense {
	x := mat.NewDense(len(self.ys), self.degree+1, nil)
	for i,_ := range self.ys {
		for j, p := 0, 1.; j <= self.degree; j, p = j+1, p*(float64(i)+xorigin) {
			x.Set(i, j, p)
		}
	}
	return x
}

// predictY predicts y value for given x, using calculated constants if available
// and the last given y value otherwise
func (self* polyBestFit) predictY(x epochDays) (float64,error) {
	if len(self.consts) ==0 {
		return 0,ENOVALIDPREDICTION
	}

	var t float64
	for i,v := range(self.consts) {
		t += math.Pow(float64(x),float64(i))*v
	}

	if t < 0 {
		return 0,ENOVALIDPREDICTION
	} 
	return t, nil
}

// predict performs brute force O(n) predicition of number of days to backfill
// given distance from given day
func (self* polyBestFit) predict(d Kilometres,sd epochDays) (epochDays,error) {

	// Check for some data to derive prediciton from
	lys := len(self.ys)
	if lys == 0 {
		return 0,ENOTENOUGHDATAPOINTS
	}

	// Subtract from target prediciton for one day after another
	// until we get to zero
	cd := sd
	for r:=d;  r > 0 ; cd ++ {

		// Get prediction for current day ...
		dd,err := self.predictY(cd)
		if err == nil {
			r -= Kilometres(dd)
		} else {
		   // ... switch to just using the last share reported
		   // on error
			return (sd + epochDays(d/Kilometres(self.ys[lys-1]))), nil
		} 
	}
	return cd, nil
}

// version reports the current version of the polynomial best fit
func (self* polyBestFit) version() predictVersion {
	return self.pv
}

// backfilled performs brute force O(n) prediction of total distance backfilled
// between end of two given days
func (self* polyBestFit) backfilled(sd epochDays,ed epochDays) (Kilometres,error) {

	// Check for some data to derive prediction from
	lys := len(self.ys)
	if lys == 0 {
		return 0,ENOTENOUGHDATAPOINTS
	}

	// Add up predicted share for each day in range given
	var t Kilometres
	for d:= sd+1; d <= ed; d++ {
		dd, err := self.predictY(d)
		if (err == nil) {
			t += Kilometres(dd)
		} else {
			// revert to simple calc using last share
			// on error
			return  Kilometres(ed-sd)*Kilometres(self.ys[lys-1]),nil
		}
	}
	return t, nil
}

