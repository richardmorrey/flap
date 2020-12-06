package flap

import (
	"github.com/richardmorrey/flap/pkg/db"
)

type Administrator struct {
	table db.Table
	params FlapParams
	predictor predictor
	pc promisesCorrection
	bs backfillState
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

	// Load any existing state
	administrator.Load()

	return administrator
}

const paramsRecordKey="params" 
const predictorRecordKey="predictor"
const correctionRecordKey="correction"
const backfillRecordKey="backfill"

// Load loads all administrative state from the database
func (self *Administrator)  Load() {

	// FLAP Parameters
	logError(self.table.Get(paramsRecordKey,&self.params))

	// Promises predictor
	self.createPredictor()
	if self.validPredictor() {
		logError(self.table.Get(predictorRecordKey, self.predictor))
	}

	// Promises correction
	logError(self.table.Get(correctionRecordKey, &self.pc))

	// Backfill state
	logError(self.table.Get(backfillRecordKey, &self.bs))

}

// Save saves all administrative state back to the database
func (self* Administrator) Save() error {

	// FLAP parameters
	err := self.table.Put(paramsRecordKey,&(self.params))
	if err != nil {
		return logError(err)
	}

	// Promises predictor
	if self.validPredictor() {
		err = self.table.Put(predictorRecordKey, self.predictor)
		if err != nil {
			return logError(err)
		}
	}

	// Promises correction
	err = self.table.Put(correctionRecordKey, &self.pc)
	if err != nil {
		return logError(err)
	}

	// Backfill state
	err = self.table.Put(backfillRecordKey, &self.bs)
	if err != nil {
		return logError(err)
	}

	return nil
}

// GetParams returns the currently active set of Flap parameters
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
	if (params.Promises.Algo != paNone && params.Promises.MaxPoints <=0) {
		return EINVALIDFLAPPARAMS
	}
	var bits int
	for n:=params.Threads; n != 0 ; n=n & (n-1) {
		bits++;
	}
	if bits >  1 || params.Threads > 16 {
		return EINVALIDFLAPPARAMS
	}

	// Check for promises config change and create new predictor
	algoOld := self.params.Promises.Algo
	self.params=params
	if params.Promises.Algo != algoOld {
		self.createPredictor()
	}
	return nil
}

// createPredictor creates predictor of the configured type
func (self* Administrator) createPredictor() {
	switch self.params.Promises.Algo & paMask { 
		case paLinearBestFit:
			self.predictor,_ = newBestFit(self.params.Promises)
		case paPolyBestFit:
			self.predictor,_ = newPolyBestFit(self.params.Promises)
	}
	if self.validPredictor() {
		logInfo("Running with promises config",self.params.Promises)
	} else {
		logInfo("Running without promises")
	}
}

// validPredictor checks Returns true if promises are enabled  and a predictor exists, and false otherwise.
func (self *Administrator) validPredictor() bool {
	_,exists := self.predictor.(*bestFit)
	if (!exists) {
		_,exists = self.predictor.(*polyBestFit)
	}
	return exists
}

// dropAdministrator Adminitrator table from given database
func dropAdministrator(database db.Database) error {
	return database.DropTable(adminTableName)
}
