package flap

import (
	"sync"
	"math"
	"encoding/binary"
	"bytes"
)

type pcState struct {
	bacSmoothed 		SmoothYs
	balanceAtClearance	Kilometres
	cdSmoothed		SmoothYs
	clearedDistance		Kilometres
	bacPerKm		Kilometres

}

type promisesCorrection struct {
	mux sync.Mutex
	state pcState
}

// To implements db/Serialize
// To implemented as part of db/Serialize
func (self *promisesCorrection) From(buff *bytes.Buffer) error {

	err := self.state.bacSmoothed.From(buff)
	if err != nil {
		return logError(err)
	}

	err = binary.Read(buff,binary.LittleEndian,&self.state.balanceAtClearance)
	if err != nil {
		return logError(err)
	}

	err = self.state.cdSmoothed.From(buff)
	if err != nil {
		return logError(err)
	}

	err = binary.Read(buff,binary.LittleEndian,&self.state.clearedDistance)
	if err != nil {
		return logError(err)
	}

	err = binary.Read(buff,binary.LittleEndian,&self.state.bacPerKm)
	
	return err

}

// From implemented as part of db/Serialize
func (self *promisesCorrection) To(buff *bytes.Buffer) error {

	err := self.state.bacSmoothed.To(buff)
	if err != nil {
		return logError(err)
	}

	err = binary.Write(buff,binary.LittleEndian,&self.state.balanceAtClearance)
	if err != nil {
		return logError(err)
	}

	err = self.state.cdSmoothed.To(buff)
	if err != nil {
		return logError(err)
	}

	err = binary.Write(buff,binary.LittleEndian,&self.state.clearedDistance)
	if err != nil {
		return logError(err)
	}

	err = binary.Write(buff,binary.LittleEndian,&self.state.bacPerKm)
	
	return err

}


// cycle updates all data needed to apply corrections to promise dates and
// resets interim values used to calculate them. To be called daily.
func (self *promisesCorrection) cycle(smoothWindow Days) Kilometres {
	self.mux.Lock()
	defer self.mux.Unlock()
	
	// Update smoothing window with total balance at clearance since last cycle
	if self.state.bacSmoothed.windowSize == 0 {
		self.state.bacSmoothed = SmoothYs{windowSize:int(math.Max(1,float64(smoothWindow))),maxYs:1}
	}
	self.state.bacSmoothed.AddY(float64(self.state.balanceAtClearance))

	// Update smoothing window for total distance in completed trips since last cycle
	if self.state.cdSmoothed.windowSize == 0 {
		self.state.cdSmoothed = SmoothYs{windowSize:int(math.Max(1,float64(smoothWindow))),maxYs:1}
	}
	self.state.cdSmoothed.AddY(float64(self.state.clearedDistance))

	// Recalulate balance at clearance per km travelled
	if self.state.cdSmoothed.ys[0] > 0 {
		self.state.bacPerKm = self.state.balanceAtClearance/Kilometres(self.state.cdSmoothed.ys[0])
	}

	// Reset cyclic values
	self.state.balanceAtClearance = 0
	self.state.clearedDistance = 0
	return Kilometres(self.state.bacSmoothed.ys[0])
}

// change updates interim values  used to calculate promises correction data
func (self *promisesCorrection) change(bac Kilometres,pd Kilometres) {
	self.mux.Lock()
	defer self.mux.Unlock()
	self.state.balanceAtClearance += bac
	self.state.clearedDistance += pd
}

// getBACPerKm returns the average balance at clearance per kilometre travelled
func (self* promisesCorrection) getBACPerKm() Kilometres {
	self.mux.Lock()
	defer self.mux.Unlock()
	return self.state.bacPerKm
}

