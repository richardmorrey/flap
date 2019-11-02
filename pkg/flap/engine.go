package flap

import (
	"github.com/richardmorrey/flap/pkg/db"
	"encoding/binary"
	"bytes"
	"math"
)

type Days int64

type FlapParams struct
{
	TripLength	Days
	FlightsInTrip   uint64
	FlightInterval  Days
	DailyTotal      Kilometres
	MinGrounded	uint64
}

type Administrator struct {
	table db.Table
	params FlapParams
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
	return  administrator
}

// GetParams returns the currently active set of Flap parameters
const adminRecordKey="flapparams" 
func (self *Administrator) GetParams() FlapParams {
	return self.params
}

// SetParams makes a new complete set of Flap parameters active. 
// The values are also written to the db table, replacing
// what was there previously
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

	// Write record as binary
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian,(&params))
	if (err != nil) {
		return err
	}
	err = self.table.Put([]byte(adminRecordKey),buf.Bytes())
	if err == nil {
		self.params=params
	}
	return err
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
func NewEngine(database db.Database) *Engine {
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

	// Retrieve or create traveller
	traveller,err := self.Travellers.GetTraveller(passport)
	if err != nil {
		traveller.passport=passport
		traveller.cleared = math.MaxInt64
	}
	
	// Add flights to traveller's flight history
	for _,flight := range flights {
		err = traveller.submitFlight(&flight,now,debit)
		if err != nil {
			return err

		}
	}
	
	// Store updated traveller
	err = self.Travellers.PutTraveller(traveller)
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
	}

	// Iterate through travellers
	var newGrounded uint64
	it,err := self.Travellers.NewIterator(nil)
	if err != nil {
		return 0,0,0,err
	}
	defer it.Release()
	var totalTravellersYesterday uint64
	var totalDistanceYesterday Kilometres
	for it.Next() {

		// Retrieve traveller
		changed:=false
		traveller := it.Value()

		// Update trip history
		distanceYesterday,err := traveller.tripHistory.Update(&self.Administrator.params,now) 
		if err == nil {
			if distanceYesterday > 0 {
				totalDistanceYesterday += distanceYesterday
				totalTravellersYesterday++
			}
			changed = true
		}

		// Backfill if grounded
		if !traveller.Cleared(now) {
			traveller.balance += share
			newGrounded++
			changed = true
		}

		// Save changes if necessary
		if changed {
			self.Travellers.PutTraveller(traveller)
		}
	}

	// Update total grounded
	self.totalGrounded=newGrounded
	return totalTravellersYesterday, totalDistanceYesterday,self.totalGrounded, it.Error()
}

