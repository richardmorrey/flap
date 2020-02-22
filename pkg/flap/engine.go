package flap

import (
	"github.com/richardmorrey/flap/pkg/db"
	"encoding/binary"
	"bytes"
	"math"
	"errors"
	"sync"
	//"fmt"
)

var EPROMISESNOTENABLED = errors.New("Promises not enabled")
var ETRIPTOOFARAHEAD = errors.New("The Trip is too far ahead")
type Days 		int64
type PromisesAlgo	uint8

const (
	paNone PromisesAlgo  = iota
	paLinearBestFit
)

type FlapParams struct
{
	TripLength		Days
	FlightsInTrip		uint64
	FlightInterval		Days
	DailyTotal		Kilometres
	MinGrounded		uint64
	PromisesAlgo		PromisesAlgo
	PromisesMaxPoints	uint32
	PromisesMaxDays		Days
	Threads			byte
}

type Administrator struct {
	table db.Table
	params FlapParams
	predictor predictor
}

// newAdministrators creates an instance of Administrator, for
// mangement of Flap parameters - Daily Total, Maximum Flight Interval,
// and Maximum Trip Duration
const adminTableName="administrator"
func newAdministrator(flapdb db.Database) *Administrator {
	
	// Create instance and create/open table 
	administrator := new(Administrator)
	table,err := flapdb.OpenTable(adminTableName)
	if  err == db.ETABLENOTFOUND { 
		table,err = flapdb.CreateTable(adminTableName)
	}
	if err != nil {
		return nil
	}
	administrator.table  = table

	// Read any parameter settings held in table
	blob,err := administrator.table.Get([]byte(adminRecordKey))
	if (err == nil) {
		buf := bytes.NewReader(blob)
		err = binary.Read(buf,binary.LittleEndian,&(administrator.params))
	}

	// Create predictor for promises
	administrator.createPredictor(true)
	
	return administrator
}

// GetParams returns the currently active set of Flap parameters
const adminRecordKey="flapparams" 
func (self *Administrator) GetParams() FlapParams {
	return self.params
}

// SetParams makes a new complete set of Flap parameters active. 
// The values are also written to the db table, replacing
// what was there previously.
func (self *Administrator) SetParams(params FlapParams) error {
	
	// Check for valid table
	if  self.table == nil {
		return ETABLENOTOPEN
	}

	// Check for invalid values
	if (params.FlightsInTrip*2  >  MaxFlights) {
		return EINVALIDFLAPPARAMS
	}
	if (params.FlightInterval*2 >  params.TripLength) {
		return EINVALIDFLAPPARAMS
	}
	if (params.PromisesAlgo != paNone && params.PromisesMaxPoints <=0) {
		return EINVALIDFLAPPARAMS
	}
	var bits int
	for n:=params.Threads; n != 0 ; n=n & (n-1) {
		bits++;
	}
	if bits >  1 {
		return EINVALIDFLAPPARAMS
	}

	// Write record as binary
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian,(&params))
	if (err != nil) {
		return err
	}
	err = self.table.Put([]byte(adminRecordKey),buf.Bytes())

	// Check for promises config change and create new predictor
	if err == nil {
		algoOld := self.params.PromisesAlgo
		self.params=params
		if params.PromisesAlgo != algoOld {
			self.createPredictor(false)
		}
	}
	return err
}

// createPredictor creates predictor of the configured type
func (self* Administrator) createPredictor(load bool) {
	switch self.params.PromisesAlgo { 
		case paLinearBestFit:
			self.predictor,_ = newBestFit(self.params.PromisesMaxPoints)
	}
	if self.validPredictor() && load {
		self.predictor.get(self.table)
	}
}

// validPredictor checks Returns true if promises are enabled  and a predictor exists, and false otherwise.
func (self *Administrator) validPredictor() bool {
	_,exists := self.predictor.(*bestFit)
	return exists
}

// dropAdministrator Adminitrator table from given database
func dropAdministrator(database db.Database) error {
	return database.DropTable(adminTableName)
}

type Engine struct
{
	Administrator 		*Administrator
	Travellers		*Travellers
	Airports		*Airports
	totalGrounded		uint64
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
// Call with care.
func Reset(database db.Database) error {
	err := dropAdministrator(database)
	if err != nil && err != db.ETABLENOTFOUND {
		return err
	}
	err = dropAirports(database)
	if err != nil && err != db.ETABLENOTFOUND {
		return err
	}
	err = dropTravellers(database)
	if err != nil && err != db.ETABLENOTFOUND {
		return err
	}
	return nil
}

// getCreateTraveller returns existing traveller record associated 
// with provided passport details if it exists, and otherwise a new
// traveller record with passport details filled in
func (self *Engine) getCreateTraveller(passport Passport)  *Traveller {
	traveller,err := self.Travellers.GetTraveller(passport)
	if err != nil {
		traveller.passport=passport
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
// If "debit" is false the distance of all flights is deducted from the travellers
// balance.
func (self *Engine) SubmitFlights(passport Passport, flights []Flight, now EpochTime,debit bool) error {

	// Check args
	if len(flights) == 0 {
		return EINVALIDARGUMENT
	}

	// Retrieve traveller record
	t := self.getCreateTraveller(passport)
	
	// Add flights to traveller's flight history
	for _,flight := range flights {
		err := t.submitFlight(&flight,now,debit)
		if err != nil {
			return err

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
func (self *Engine) UpdateTripsAndBackfill(now EpochTime) (uint64,Kilometres,uint64,error) {
	
	// Check we are at start of day
	if now % SecondsInDay != 0 {
		return 0,0,0,EINVALIDARGUMENT
	}

	// Calculate backfill share
	var share Kilometres
	backfillers := 	Kilometres(math.Max(float64(self.Administrator.params.MinGrounded),float64(self.totalGrounded)))
	if backfillers > 0 {
		share = self.Administrator.params.DailyTotal / backfillers

		// Add calculated share to predictor algorithm
		if self.Administrator.validPredictor() {
			self.Administrator.predictor.add(now.toEpochDays(false),share)
			self.Administrator.predictor.put(self.Administrator.table)
			logInfo("Added predictior data point:",now.toEpochDays(false),share)
		}
	}

	// Update all travellers
	threads := uint(self.Administrator.params.Threads)
	if threads == 0 {
		threads = 1
	}
	logInfo("Backfilling with", threads,"threads")
	stats := make(chan updateStats, threads)
	var wg sync.WaitGroup
	delta := (math.MaxUint8+1)/threads
	for i := uint(0); i < math.MaxUint8; i+=delta {
		wg.Add(1)
		t :=  func(s byte,e byte) {stats <- self.updateSomeTravellers(s,e,share,now);wg.Done()}
		go t(byte(i),byte(i+delta-1))
	}
	wg.Wait()

	// Add up the stats
	var ut updateStats 
	close(stats)
	for elem := range stats {
		ut.grounded += elem.grounded
		ut.travellers += elem.travellers
		ut.distance += elem.distance
		if (elem.err != nil) {
			ut.err = elem.err
		}
	}

	// Update total grounded and return
	self.totalGrounded=ut.grounded
	return ut.travellers,ut.distance,self.totalGrounded,ut.err
}

type updateStats struct {
	grounded uint64
	travellers uint64
	distance  Kilometres
	err	error
}

func (self *Engine) updateSomeTravellers(prefixStart byte, prefixEnd byte, share Kilometres,now EpochTime) updateStats {

	var us updateStats
	var prefix [1]byte

	// Iterate through all keys with a first byte in the given
	// range
	logInfo("Backfilling from",prefixStart,"to",prefixEnd)
	for pc:=int(prefixStart); pc <= int(prefixEnd); pc++ {

		// Iterate over current start byte
		prefix[0]=byte(pc)
		it,err := self.Travellers.NewIterator(prefix[:])
		if err != nil {
			us.err=err
			return us
		}
		//fmt.Printf("Iterating over %d\n",prefix)
		for it.Next() {

			// Retrieve traveller
			changed:=false
			traveller := it.Value()
			// Update trip history
			distanceYesterday,err := traveller.tripHistory.Update(&self.Administrator.params,now) 
			if err == nil {
				if distanceYesterday > 0 {
					us.distance += distanceYesterday
					us.travellers ++
				}
				changed = true
			}

			// Check for a promise to keep
			kept := traveller.keep()
			changed = changed || kept

			// Backfill if not travelling and balance is negative
			if !traveller.MidTrip() && traveller.balance < 0 {
				traveller.balance += share
				us.grounded++
				changed = true
			}

			// Save changes if necessary
			if changed {
				self.Travellers.PutTraveller(traveller)
			}
		}

		// Release interface
		us.err = it.Error()
		it.Release()
		if us.err != nil {
			return us
		}
	}
	logInfo("Finished backfilling from",prefixStart,"to",prefixEnd)
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

	// Determine trip start, end, and distance
	var td Kilometres
	ts := MaxEpochTime
	te := tripEnd
	for i:=0; i < len(flights); i++ {
		td += flights[i].distance
		if flights[i].start < ts {
			ts=flights[i].start
		}
		if flights[i].end  > te {
			te=flights[i].end
		}
	}

	// Check proposed trip is not too far in the future
	if ts.toEpochDays(true) - now.toEpochDays(false) > epochDays(self.Administrator.params.PromisesMaxDays) {
		return nil,ETRIPTOOFARAHEAD
	}

	// Ask for proposal and return the result
	return self.getCreateTraveller(passport).Promises.propose(ts,te,td,now,self.Administrator.predictor)
}

// Make attempts to apply a proposal for changes to a traveller's set of clearance promises.
func (self *Engine) Make(passport Passport, proposal *Proposal) error {

	// Validate arguments
	if proposal == nil {
		return EINVALIDARGUMENT
	}

	// Check promises are active
	if !self.Administrator.validPredictor() {
		return EPROMISESNOTENABLED
	}

	// Make promise
	t := self.getCreateTraveller(passport)
	err := t.Promises.make(proposal,self.Administrator.predictor)
	if (err == nil) {
		self.Travellers.PutTraveller(*t)
	}
	return err
}

