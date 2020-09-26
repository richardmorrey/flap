package flap

import (
	"sync"
	"math"
	"encoding/binary"
	"bytes"
)

type pcState struct {
	bacSmoothed 		smoothYs
	balanceAtClearance	Kilometres
	cdSmoothed		smoothYs
	clearedDistance		Kilometres
	bacPerKm		Kilometres

}

type promisesCorrection struct {
	mux sync.Mutex
	state pcState
}

// To implements db/Serialize
func (self *promisesCorrection) To(buff *bytes.Buffer) error {
	return binary.Write(buff, binary.LittleEndian,self.state)
}

// From implemments db/Serialize
func (self *promisesCorrection) From(buff *bytes.Buffer) error {
	return binary.Read(buff,binary.LittleEndian,self.state)
}

// cycle updates all data needed to apply corrections to promise dates and
// resets interim values used to calculate them. To be called daily.
func (self *promisesCorrection) cycle(smoothWindow Days) Kilometres {
	self.mux.Lock()
	defer self.mux.Unlock()
	
	// Update smoothing window with total balance at clearance since last cycle
	if self.state.bacSmoothed.windowSize == 0 {
		self.state.bacSmoothed = smoothYs{windowSize:int(math.Max(1,float64(smoothWindow))),maxYs:1}
	}
	self.state.bacSmoothed.addY(float64(self.state.balanceAtClearance))

	// Update smoothing window for total distance in completed trips since last cycle
	if self.state.cdSmoothed.windowSize == 0 {
		self.state.cdSmoothed = smoothYs{windowSize:int(math.Max(1,float64(smoothWindow))),maxYs:1}
	}
	self.state.cdSmoothed.addY(float64(self.state.clearedDistance))

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

