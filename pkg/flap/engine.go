package flap

import (
	"github.com/richardmorrey/flap/pkg/db"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"bytes"
	"math"
	"errors"
	"sync"
	//"fmt"
)

var EPROMISESNOTENABLED = errors.New("Promises not enabled")
var ETRIPTOOFARAHEAD = errors.New("The Trip is too far ahead")
type Days 			int64
type PromisesAlgo		uint8
const (
	paNone PromisesAlgo = 0x00  
	paLinearBestFit PromisesAlgo = 0x01
	paPolyBestFit PromisesAlgo = 0x02
	pamCorrectBalances PromisesAlgo = 0x10
	pamCorrectDailyTotal PromisesAlgo = 0x20
	pamCorrectPromiseDistance PromisesAlgo = 0x40
	paMask PromisesAlgo = 0x0f
)

type PromisesConfig struct{
	Algo		  	PromisesAlgo
	MaxPoints 	  	uint32
	MaxDays		  	Days
	MaxStackSize	  	StackIndex
	SmoothWindow	  	Days
	CorrectionSmoothWindow	Days
	Degree  	  	uint32
}

type FlapParams struct {
	TripLength		Days
	FlightsInTrip		uint64
	FlightInterval		Days
	DailyTotal		Kilometres
	MinGrounded		uint64
	Promises		PromisesConfig
	TaxiOverhead		Kilometres
	Threads			byte
}

func (self* FlapParams) To(b *bytes.Buffer) error {
	enc := gob.NewEncoder(b) 
	return enc.Encode(self)
}

func (self* FlapParams) From(b *bytes.Buffer) error {
	dec := gob.NewDecoder(b)
	return dec.Decode(self)
}

type backfillState struct {
	totalGrounded uint64
}

// To implements db/Serialize
func (self *backfillState) To(buff *bytes.Buffer) error {
	return binary.Write(buff, binary.LittleEndian,&self.totalGrounded)
}

// From implemments db/Serialize
func (self *backfillState) From(buff *bytes.Buffer) error {
	return binary.Read(buff,binary.LittleEndian,&self.totalGrounded)
}

type Engine struct
{
	Administrator 		*Administrator
	Travellers		*Travellers
	Airports		*Airports
}

// NewEngine creates an instance of an Engine object, which can be used
// to set the FLAP parameters, as well as drive the key FLAP daily processes:
// - Submission of flights taken by carriers
// - Updating of the status of Journeys and Trips for every Traveller.
// - Backfilling of the Daily Total to all grounded Travellers.
func NewEngine(database db.Database, logLevel LogLevel,logFolder string) *Engine {
	NewLogger(logLevel,logFolder)
	engine := new(Engine)
	engine.Travellers = NewTravellers(database)
	engine.Airports   = NewAirports(database)
	engine.Administrator = newAdministrator(database)
	return engine
}

// Reset drops ALL FLAP tables from given database
// holding state related to travellers. If destroy is true
// all tables are dropped
func Reset(database db.Database, destroy bool) error {
	err := dropAdministrator(database)
	if err != nil && err != db.ETABLENOTFOUND {
		return err
	}
	err = dropTravellers(database)
	if err != nil && err != db.ETABLENOTFOUND {
		return err
	}
	if destroy {
		err = DropAirports(database)
		if err != nil && err != db.ETABLENOTFOUND {
			return err
		}
	}
	return nil
}

// getCreateTraveller returns existing traveller record associated 
// with provided passport details if it exists, and otherwise a new
// traveller record with passport details filled in
func (self *Engine) getCreateTraveller(passport Passport, now EpochTime)  *Traveller {
	traveller,err := self.Travellers.GetTraveller(passport)
	if err != nil {
		traveller.passport=passport
		traveller.Created=now
	}
	return &traveller
}

// SubmitFlights submits a list of one or more flights for the traveller
// with the specified passport. It is intended to be invoked by the Carrier
// for each check-in and takes multiple flights to allow for through
// check-ins. If the traveller is not cleared to 
// travel for one or more of the flights the whole submission is rejected
// and the function returned with EGROUNDED. Ths is in effect an instruciton
// to the carrier to refuse the check-in.
// If "debit" is true the distance of all flights is deducted from the travellers
// balance.
func (self *Engine) SubmitFlights(passport Passport, flights []Flight, now EpochTime,debit bool) error {

	// Check args
	if len(flights) == 0 {
		return EINVALIDARGUMENT
	}

	// Retrieve traveller record
	t := self.getCreateTraveller(passport,now)

	
	// Add flights to traveller's flight history
	for _,flight := range flights {

		// Update traveller with the new flight
		bac,pd,err := t.submitFlight(&flight,now,self.Administrator.params.TaxiOverhead,debit)
		if err != nil {
			return err
		}

		// Apply any configured balance adjustment
		self.Administrator.pc.change(bac,pd) 
		if (self.Administrator.params.Promises.Algo & pamCorrectBalances == pamCorrectBalances) &&
				   (bac < 0) {
			t.transact(-bac,now,TTBalanceAdjustment)
		}
	}

	// Store updated traveller
	err := self.Travellers.PutTraveller(*t)
	return err
}

// UpdateTripsAndBackfill iterates through all Traveller records, carrying out
// two key FLAP processes for each traveller:
// (1) Update the trip history, applying FLAP parameters and the provided date time to end journeys and trips
// (2) Backfilling with a share of the DailyTotal if the traveller is grounded.
// Note it counts and stores the total number of grounded travellers over the course of the iteration to use
// for calculation of the backfill share for the next invocation.
// It must be invoked once a day with a datetime that is the start of that UTC day.
func (self *Engine) UpdateTripsAndBackfill(now EpochTime) (UpdateBackfillStats,error) {
	
	// Check we are at start of day
	ut := *NewUpdateBackfillStats()
	if now % SecondsInDay != 0 {
		return ut,EINVALIDARGUMENT
	}

	// Retrieve and cycle promises correction if enabled
	var pc Kilometres
	if self.Administrator.params.Promises.Algo & pamCorrectDailyTotal == pamCorrectDailyTotal {
		pc = self.Administrator.pc.cycle(self.Administrator.params.Promises.CorrectionSmoothWindow)
		logDebug("DailyTotal=",self.Administrator.params.DailyTotal,"PromisesCorrection=",pc)
	}

	// Calculate backfill share
	backfillers := 	Kilometres(math.Max(float64(self.Administrator.params.MinGrounded),float64(self.Administrator.bs.totalGrounded)))
	if backfillers > 0 {
		ut.Share = (self.Administrator.params.DailyTotal+pc) / backfillers

		// Add calculated share to predictor algorithm
		if self.Administrator.validPredictor() {
			self.Administrator.predictor.add(now.toEpochDays(false),ut.Share)
			ut.BestFitPoints,ut.BestFitConsts,_ = self.Administrator.predictor.state()
			logInfo("Added predictior data point:",now.toEpochDays(false),ut.Share)
		}
	}

	// Create snapshot for faster multithreaded reads
	ss,err := self.Travellers.TakeSnapshot()
	if err != nil {
		return UpdateBackfillStats{},logError(err)
	}
	defer ss.Release()

	// Update all travellers
	threads := uint(self.Administrator.params.Threads)
	if threads == 0 {
		threads = 1
	}
	logDebug("Backfilling with ", threads," threads")
	stats := make(chan UpdateBackfillStats, threads)
	var wg sync.WaitGroup
	delta := 16/threads
	for i := uint(0); i < 16; i+=delta {
		wg.Add(1)
		t :=  func(s byte,e byte) {stats <- self.updateSomeTravellers(s,e,ut.Share,now,ss);wg.Done()}
		go t(byte(i),byte(i+delta-1))
	}
	wg.Wait()

	// Add up the stats
	close(stats)
	for elem := range stats {
		ut.Grounded += elem.Grounded
		ut.Travellers += elem.Travellers
		ut.Distance += elem.Distance
		ut.Flights += elem.Flights
		ut.ClearedDistanceDeltas = append(ut.ClearedDistanceDeltas,elem.ClearedDistanceDeltas...)
		ut.ClearedDaysDeltas = append(ut.ClearedDaysDeltas,elem.ClearedDaysDeltas...)
		if (elem.Err != nil) {
			ut.Err = elem.Err
		}
	}

	// Update total grounded and return
	self.Administrator.bs.totalGrounded=ut.Grounded
	return ut,ut.Err
}

type UpdateBackfillStats struct {
	Grounded 		uint64
	Travellers 		uint64
	Distance  		Kilometres
	Flights			uint64
	Share			Kilometres
	ClearedDistanceDeltas	[]Kilometres
	ClearedDaysDeltas	[]Days
	BestFitPoints		[]float64
	BestFitConsts		[]float64
	Err			error
}

func NewUpdateBackfillStats() *UpdateBackfillStats {
	ubs := new(UpdateBackfillStats)
	ubs.ClearedDistanceDeltas = make([]Kilometres,0,1000)
	ubs.ClearedDaysDeltas = make([]Days,0,1000)
	return ubs
}

func (self *Engine) updateSomeTravellers(prefixStart byte, prefixEnd byte, share Kilometres,now EpochTime, ss *TravellersSnapshot) UpdateBackfillStats {

	us := *NewUpdateBackfillStats()
	var prefix [1]byte

	// Iterate through all keys with a first byte in the given
	// range
	logDebug("Backfilling from ",prefixStart," to ",prefixEnd)
	bw,err := self.Travellers.MakeBatch(10000)
	if err != nil {
		us.Err = logError(err)
		return us
	}
	defer bw.Release();

	for pc:=int(prefixStart); pc <= int(prefixEnd); pc++ {

		// Iterate over current start byte
		prefix[0]=byte(pc)
		prefixstr := hex.EncodeToString(prefix[:])
		it,err := ss.NewIterator(prefixstr[1:])
		if err != nil {
			us.Err= logError(err)
			return us
		}
		for it.Next() {

			// Retrieve traveller
			changed:=false
			traveller := it.Value()

			// Update trip history
			distanceYesterday,flightsYesterday,err := traveller.tripHistory.Update(&self.Administrator.params,now) 
			if err == nil {
				if distanceYesterday > 0 {
					us.Distance += distanceYesterday
					us.Travellers ++
					us.Flights += flightsYesterday
				}
				changed = true
			}
			
			// Report any clearance deltas if appropriate
			if (traveller.Kept.Clearance > 0 && traveller.Kept.StackIndex==0) {
				nowDays := Days(now.toEpochDays(false))
				clearDays := Days(traveller.Kept.Clearance.toEpochDays(false))
				if (nowDays == clearDays) {
					us.ClearedDistanceDeltas = append(us.ClearedDistanceDeltas,traveller.Balance)
				}
				if (traveller.Balance + share >= 0) {
					us.ClearedDaysDeltas = append(us.ClearedDaysDeltas,nowDays-clearDays)
				}
			}

			// Backfill if not travelling and balance is negative
			if !traveller.MidTrip() && traveller.Balance < 0 {
				traveller.transact(share,now,TTDailyShare)
				us.Grounded++
				changed = true
			}

			// Check for a promise to keep
			kept := traveller.keep()
			if kept {
				changed = true
			}

			// Save changes if necessary
			if changed {
				bw.Put(traveller)
			}

		}

		// Release interface
		us.Err = it.Error()
		it.Release()
		if us.Err != nil {
			return us
		}
	}
	logDebug("Finished backfilling from ",prefixStart," to ",prefixEnd)
	return us
}

// Propose returns a proposal for change to the given traveller's set of clearance promises to
// accomodate proposed ordered set of flights whilst keeping all existing promises.
// The set of flights are treated as a single trip with start time the start time of the first
// flight and the end of the trip being endTrip if endTrip isnt 0 and the end of the last flight
// otherwise. Returns error if no proposal can be made.
func (self *Engine) Propose(passport Passport,flights [] Flight, tripEnd EpochTime, now EpochTime) (*Proposal,error) {

	// Validate args
	if len(flights)==0 {
		return nil,EINVALIDARGUMENT
	}
	
	// Check promises are active
	if !self.Administrator.validPredictor() {
		return nil,EPROMISESNOTENABLED
	}

	// Determine trip start, end, distance to backfill and distance travelled
	var distance Kilometres
	var travelled Kilometres
	ts := MaxEpochTime
	te := tripEnd
	for i:=0; i < len(flights); i++ {
		travelled += flights[i].Distance
		distance += flights[i].Distance + self.Administrator.params.TaxiOverhead
		if flights[i].Start < ts {
			ts=flights[i].Start
		}
		if flights[i].End  > te {
			te=flights[i].End
		}
	}

	// Check proposed trip is not too far in the future
	if ts.toEpochDays(true) - now.toEpochDays(false) > epochDays(self.Administrator.params.Promises.MaxDays) {
		return nil,ETRIPTOOFARAHEAD
	}

	// Adjust distance for mean balance at clearance if so configured
	if self.Administrator.params.Promises.Algo & pamCorrectPromiseDistance == pamCorrectPromiseDistance{
		distance -= self.Administrator.pc.getBACPerKm()*distance
	}

	// Ask for proposal and return the result
	return self.getCreateTraveller(passport,now).Promises.propose(ts,te,distance,travelled,now,self.Administrator.predictor,
								  self.Administrator.params.Promises.MaxStackSize)
}

// Make attempts to apply a proposal for changes to a traveller's set of clearance promises.
func (self *Engine) Make(passport Passport, proposal *Proposal, now EpochTime) error {

	// Validate arguments
	if proposal == nil {
		return EINVALIDARGUMENT
	}

	// Check promises are active
	if !self.Administrator.validPredictor() {
		return EPROMISESNOTENABLED
	}

	// Make promise
	t := self.getCreateTraveller(passport,now)
	err := t.Promises.make(proposal,self.Administrator.predictor)
	if (err == nil) {
		self.Travellers.PutTraveller(*t)
	}
	return err
}

// Release saves state and clears up resources when instance is finished with
func (self *Engine) Release() {
	self.Administrator.Save()
}

