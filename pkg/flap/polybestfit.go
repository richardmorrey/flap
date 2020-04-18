package flap 

import (
	"encoding/gob"
	"bytes"
	"gonum.org/v1/gonum/mat"
	"math"
	"reflect"
	"fmt"
)

// polyBestFit predicts dates when a specified distance balance would
// return to credit using a polynomial linear best fit against a plot
// of distance share against day.
type polyBestFit struct {
	smoothYs
	pv		predictVersion
	consts		[]float64
	degree		int
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
	return bf,bf.setWindows(int(cfg.MaxPoints),int(cfg.SmoothWindow))
}

// adds a point to the graph and recalculates polynomial best fit
// using configured degree. Based on https://rosettacode.org/wiki/Polynomial_regression#Go
func (self* polyBestFit) add(today epochDays,y Kilometres) {

    // Add new y value
    self.addY(float64(y))

    if len(self.ys) > self.degree {

	// Build X and Y matrices
	a := self.Vandermonde(float64(today) - float64(epochDays(len(self.ys)-1)))
	b := mat.NewDense(len(self.ys), 1, self.ys)

	// Calculate constants
	c := mat.NewDense(self.degree+1, 1, nil)
	qr := new(mat.QR)
	qr.Factorize(a)
	err := qr.SolveTo(c, false, b)

	// Extract results
	newConsts := make([]float64,0,self.degree+1)
	if err == nil {
		for j:=0; j < self.degree+1; j++ {
			newConsts = append(newConsts,c.At(j,0))
		}
		if !reflect.DeepEqual(newConsts,self.consts) {
			self.pv++
			fmt.Printf("old=%#v,new=%#v\n",self.consts,newConsts)
			self.consts = newConsts
		}
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
func (self* polyBestFit) predictY(x epochDays) float64 {
	var t float64
	if len(self.consts) ==0 && len(self.ys) > 0 {
		t = self.ys[len(self.ys)-1]
	} else {
		for i,v := range(self.consts) {
			t += math.Pow(float64(x),float64(i))*v
		}
	}
	return t
}

// predict performs brute force O(n) predicition of number of days to backfill
// given distance from given day
func (self* polyBestFit) predict(d Kilometres,sd epochDays) (epochDays,error) {
	cd := sd
	for r:=d;  r > 0 ; cd ++ {
		r -= Kilometres(self.predictY(cd))
	}
	return cd, nil
}

// version reports the current version of the polynomial best fit
func (self* polyBestFit) version() predictVersion {
	return self.pv
}

// backfilled performs brute force O(n) prediction of total distance backfilled
// between two given days
func (self* polyBestFit) backfilled(sd epochDays,ed epochDays) (Kilometres,error) {
	var t Kilometres
	for d:= sd; d < ed; d++ {
		t += Kilometres(self.predictY(d))
	}
	return t, nil
}

// To implemented as part of db/Serialize
func (self *polyBestFit) To(buff *bytes.Buffer) error {
	dec := gob.NewDecoder(buff)
	return dec.Decode(self)
}

// From implemented as part of db/Serialize
func (self *polyBestFit) From(buff *bytes.Buffer) error {
	enc := gob.NewEncoder(buff) 
	return enc.Encode(self)
}

