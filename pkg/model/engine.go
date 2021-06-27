package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"github.com/richardmorrey/flap/pkg/db"
	"path/filepath"
	"errors"
	"fmt"
	"os"
	"time"
	"math"
	"math/rand"
	"sort"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"encoding/json"
	"encoding/binary"
	"bytes"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/floats"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/palette/brewer"
	"image/color"
)

var EFAILEDTOCREATECOUNTRIESAIRPORTSROUTES = errors.New("Failed to create Countries-Airports-Routes")
var EFAILEDTOCREATETRAVELLERBOTS = errors.New("Failed to create traveller bots")
var EMODELNOTBUILT = errors.New("Failed to find model to load")
var EINVALIDSTARTDAY = errors.New("Invalid start day")
var ENOSUCHTRAVELLER = errors.New("No such traveller")
var EFAILEDTOCREATECOUNTRYWEIGHTS = errors.New("Failed to create country weights")

type Probability float64

type BotSpec struct {
	FlyProbability		Probability
	Weight			weight
	MonthWeights		[]weight
}

type DBType		uint8
const (
	dbLevel DBType = iota
	dbDatastore
)

type DBSpec struct {
	DBType			DBType
	ConnectionString	string
}

type ModelParams struct {
	WorkingFolder		string
	DBSpec			DBSpec
	DataFolder		string
	TrialDays		flap.Days
	DTAlgo			string
	DaysToRun		flap.Days
	TotalTravellers		uint64
	TravellersDailyIncrease	uint64
	BotSpecs		[]BotSpec
	TripLengths		[]flap.Days
	StartDay		time.Time
	DailyTotalFactor	float64
	DailyTotalDelta		float64
	ReportDayDelta		flap.Days
	VerboseReportDayDelta	flap.Days
	ChartWidth		float64
	LargeChartWidth    	float64
	LogLevel		logLevel
	Deterministic		bool
	Threads			uint
	BotFreqFactor		float64
}

type plannedFlight struct {
	from			flap.ICAOCode
	to			flap.ICAOCode
}

type verboseStats struct {
	clearedDistanceDeltas	[]float64
	clearedDaysDeltas	[]float64
}

// reset resets the summary stats structure after a reporting event, avoiding
// unnecessary new memory allocations
func (self *verboseStats) reset() {
	var ssReset verboseStats
	if self.clearedDistanceDeltas == nil {
		ssReset.clearedDistanceDeltas = make([]float64,0,10000)
	} else {
		ssReset.clearedDistanceDeltas = self.clearedDistanceDeltas[:0]
	}
	if self.clearedDaysDeltas == nil {
		ssReset.clearedDaysDeltas = make([]float64,0,10000)
	} else {
		ssReset.clearedDaysDeltas = self.clearedDaysDeltas[:0]
	}
	*self=ssReset
}

type Engine struct {
	FlapParams 			flap.FlapParams
	ModelParams			ModelParams
	plannedFlights			[]plannedFlight
	db				db.Database
	table				db.Table
	verbose				verboseStats
}

type modelState struct {
	totalDayOne float64
	travellersForMinGrounded float64
	totalTravellersCurrent  uint64
	startDate  flap.EpochTime
}

// From implements db/Serialize
func (self *modelState) From(buff *bytes.Buffer) error {
	err := binary.Read(buff,binary.LittleEndian,&self.totalDayOne)
	if err != nil {
		return logError(err)
	}

	err = binary.Read(buff,binary.LittleEndian,&self.startDate)
	if err != nil {
		return logError(err)
	}

	err = binary.Read(buff,binary.LittleEndian,&self.travellersForMinGrounded)
	if err != nil {
		return logError(err)
	}

	err = binary.Read(buff,binary.LittleEndian,&self.totalTravellersCurrent)
	return err
}

// To implemented as part of db/Serialize
func (self *modelState) To(buff *bytes.Buffer) error {

	err := binary.Write(buff,binary.LittleEndian,&self.totalDayOne)
	if err != nil {
		return logError(err)
	}

	err = binary.Write(buff,binary.LittleEndian,&self.startDate)
	if err != nil {
		return logError(err)
	}

	err = binary.Write(buff,binary.LittleEndian,&self.travellersForMinGrounded)
	if err != nil {
		return logError(err)
	}

	err = binary.Write(buff,binary.LittleEndian,&self.totalTravellersCurrent)
	return err
}

const modelstateRecordKey="modelstate"

// load loads engine state from given table
func (self *modelState) load(t db.Table) error {
	return t.Get(modelstateRecordKey,self)
}

// save saves engine state to given table
func (self *modelState)  save(t db.Table) error {
	return t.Put(modelstateRecordKey,self)
}

// NewEngine is factory function for Engine
func NewEngine(configFilePath string) (*Engine,error) {

	e:= new(Engine)

	// Load config file
	buff, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil,logError(err)
	}
	err = yaml.Unmarshal(buff, &e)
	if err != nil {
		return nil,logError(err)
	}

	// Set default config values
	if e.ModelParams.ReportDayDelta == 0 {
		e.ModelParams.ReportDayDelta = 1
	}
	if e.ModelParams.Threads == 0 {
		e.ModelParams.Threads = 1
	}
		
	// Validate config
	for _,length := range(e.ModelParams.TripLengths) {
		if length < 2 {
			return nil,logError(flap.EINVALIDARGUMENT)
		}
	}

	// Initialize logger
	NewLogger(e.ModelParams.LogLevel,e.ModelParams.WorkingFolder)

	// Create db
	switch (e.ModelParams.DBSpec.DBType) {
		case dbDatastore:
			e.db = db.NewDatastoreDB(e.ModelParams.DBSpec.ConnectionString)
		default:
			e.db = db.NewLevelDB(e.ModelParams.WorkingFolder)
	}

	// Create model table
	table,err := e.db.OpenTable(modelTableName)
	if  err == db.ETABLENOTFOUND { 
		table,err = e.db.CreateTable(modelTableName)
	}
	if err != nil {
		return e,err
	}
	e.table  = table

	// Seed Random Number Generator
	if e.ModelParams.Deterministic {
		rand.Seed(0) 
	} else {
		rand.Seed(time.Now().UTC().UnixNano())
	}
	
	// Default config for plotter
	plot.DefaultFont="Helvetica"
	return e,nil
}

//Release releases all resources that need to be explicitly released when finished with an Engine
func (self *Engine) Release() {
	self.db.Release()
}

const modelTableName="model"

// Build prepares all persitent data files in order to be able to run the model in the configured data folder. It
// builds the countries-airports-routes and the country weights file that drive flight selection.
func (self *Engine) Build() error {
	
	fmt.Printf("Building...\n")

	//  Reset flap and load airports
	self.Reset(true)
	fe := flap.NewEngine(self.db,flap.LogLevel(self.ModelParams.LogLevel),self.ModelParams.WorkingFolder)
	defer fe.Release()
	err := fe.Administrator.SetParams(self.FlapParams)
	if (err != nil) {
		return logError(err)
	}
	err =fe.Airports.LoadAirports(filepath.Join(self.ModelParams.DataFolder,"airports.dat"))
	if (err != nil) {
		return logError(err)
	}

	// Build countries-airports-flights table from real-world data
	cw := newCountryWeights()
	if  cw == nil {
		return EFAILEDTOCREATECOUNTRYWEIGHTS
	}
	cars := NewCountriesAirportsRoutes(self.db)
	if cars  == nil {
		return EFAILEDTOCREATECOUNTRIESAIRPORTSROUTES
	}
	err = cars.Build(self.ModelParams.DataFolder,cw)
	if (err != nil) {
		return logError(err)
	}
	err = cw.save(self.table)
	if err != nil {
		return logError(err)
	}
	
	fmt.Printf("...Finished\n")
	return nil
}

func max(a flap.Kilometres, b flap.Kilometres) flap.Kilometres {
	if a > b {
		return a
	}
	return b
}

func min(a uint64, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}

// Returns longest and shortest trip lengths
func (self* Engine) minmaxTripLength() (flap.Days,flap.Days) {
	var maxLength flap.Days
	var minLength = flap.Days(math.MaxInt64)
	for _,length := range(self.ModelParams.TripLengths) {
		if length > maxLength {
			maxLength=length
		}
		if (length < minLength) {
			minLength=length
		}
	}
	return minLength, maxLength
}

// prepare sets up all data structures needed for a post-build modelling operation
func (self* Engine) prepare() (*CountriesAirportsRoutes, *TravellerBots, *flap.Engine,*journeyPlanner,error) {
	
	// Validate model params
	startDay := flap.EpochTime(self.ModelParams.StartDay.Unix())
	if startDay % flap.SecondsInDay != 0 || startDay == 0 {
		return nil,nil,nil,nil,EINVALIDSTARTDAY
	}

	// Load country-airports-routes model
	cars := NewCountriesAirportsRoutes(self.db)
	if cars == nil {
		return nil,nil,nil,nil,EMODELNOTBUILT
	}

	// Load country weights
	cw := newCountryWeights()
	if  cw == nil {
		return nil,nil,nil,nil,EFAILEDTOCREATECOUNTRYWEIGHTS
	}
	err := cw.load(self.table)
	if err != nil {
		return nil,nil,nil,nil,logError(err)
	}
	
	// load model state
	var ms modelState
	err = ms.load(self.table)
	if err != nil {
		return nil,nil,nil,nil,logError(err)
	}

	// Build flight plans for traveller bots
	travellerBots := NewTravellerBots(cw)
	if travellerBots == nil {
		return nil,nil,nil,nil,EFAILEDTOCREATETRAVELLERBOTS
	}
	err = travellerBots.Build(self.ModelParams,self.FlapParams,ms.totalTravellersCurrent,self.table)
	if (err != nil) {
		return nil,nil,nil,nil,logError(err)
	}
	
	// Create engine
	fe := flap.NewEngine(self.db,flap.LogLevel(self.ModelParams.LogLevel),self.ModelParams.WorkingFolder)

	// Load Journey planner
	jp,err := NewJourneyPlanner(self.db)
	if (err != nil) {
		return nil,nil,nil,nil,logError(err)
	}
	return cars,travellerBots,fe,jp,nil
}

/// modelDay models a single day - planning flights, submitting flights and performing update and backfill
func (self Engine) modelDay(currentDay flap.EpochTime,cars *CountriesAirportsRoutes, tb *TravellerBots, fe *flap.Engine, jp *journeyPlanner, fp *flightPaths) (flap.UpdateBackfillStats, flap.Kilometres,error) {

	// load flap params
	flapParams := fe.Administrator.GetParams()

	// load model state
	var ms modelState
	err := ms.load(self.table)
	if err != nil {
		return flap.UpdateBackfillStats{},0,logError(err)
	}

	// Calculate day of model
	i := 1 + (flap.Days(currentDay/flap.SecondsInDay) - flap.Days(ms.startDate/flap.SecondsInDay))

	// For each travller: Update triphistory and backfill those with distance accounts in
	// deficit.
	fmt.Printf("\rDay %d: Backfilling       ",i)
	us,err :=  fe.UpdateTripsAndBackfill(currentDay)
	if err != nil {
		return flap.UpdateBackfillStats{},0,logError(err)
	}

	// Plan flights for all travellers
	logInfo("DAY ", i ," ", currentDay.ToTime())
	fmt.Printf("\rDay %d: Planning Flights",i)
	err = tb.planTrips(cars,jp,fe,currentDay,self.ModelParams.Deterministic,i,self.ModelParams.Threads)
	if err != nil {
		return flap.UpdateBackfillStats{},0,logError(err)
	}

	// Submit all flights for this day, logging only - i.e. not debiting distance accounts - if
	// we are still pre-loading the journey planner.
	fmt.Printf("\rDay %d: Submitting Flights",i)
	err = jp.submitFlights(tb,fe,currentDay,fp,i>self.ModelParams.TrialDays)
	if err != nil && err != ENOJOURNEYSPLANNED {
		return flap.UpdateBackfillStats{},0,logError(err)
	}

	// If in trial period calculate starting daily total and minimum grounded travellers
	if (i > 0 && i <= self.ModelParams.TrialDays) {
			
		// Calculate DT as average across trial days if specified, defaulting to minimum
		if self.ModelParams.DTAlgo=="average" {
			flapParams.DailyTotal += flap.Kilometres(float64(us.Distance)/float64(self.ModelParams.TrialDays))
		} else {
			flapParams.DailyTotal=max(flapParams.DailyTotal,us.Distance)
		}
		ms.totalDayOne = float64(flapParams.DailyTotal)

		// Set MinGrounded, used by flap to ensure initial backfill share
		// is not too large, to average number of travellers per day over the trial period
		ms.travellersForMinGrounded += float64(us.Travellers)
		flapParams.MinGrounded = uint64(math.Ceil(ms.travellersForMinGrounded/float64(i)))
	} 

	// If next day is beyond trial period then set daily total and min grounded
	if i >= self.ModelParams.TrialDays {

		// Adjust Daily Total for the next day
		flapParams.DailyTotal = flap.Kilometres(float64(flapParams.DailyTotal)*self.ModelParams.DailyTotalFactor)
		flapParams.DailyTotal += flap.Kilometres(self.ModelParams.DailyTotalDelta*ms.totalDayOne/100.0)
		
		// Add new travellers if configured to do so
		if self.ModelParams.TravellersDailyIncrease != 0 {
			err := self.AdjustDailyTotal(self.ModelParams.TravellersDailyIncrease,fe)
			if err != nil {
				return flap.UpdateBackfillStats{},0,logError(err)
			}
			ms.totalTravellersCurrent += self.ModelParams.TravellersDailyIncrease
		}
	}

	// Save any changes to flap params
	err = fe.Administrator.SetParams(flapParams)
	if err != nil {
		return flap.UpdateBackfillStats{},0,logError(err)
	}

	// save engine state
	err = ms.save(self.table)
	if err != nil {
		return flap.UpdateBackfillStats{},0,logError(err)
	}

	// Update summary stats
	var ss summaryStats
	if i>0 {
		ss.update(summaryStatsRow{
			DailyTotal:float64(flapParams.DailyTotal)/float64(self.ModelParams.ReportDayDelta),
			Travellers:float64(us.Travellers)/float64(self.ModelParams.ReportDayDelta),
			Travelled:float64(us.Distance)/float64(self.ModelParams.ReportDayDelta),
			Flights:float64(us.Flights)/float64(self.ModelParams.ReportDayDelta),
			Grounded:float64(us.Grounded)/float64(self.ModelParams.ReportDayDelta), 
			Share: float64(us.Share)/float64(self.ModelParams.ReportDayDelta),
			Date: currentDay},
			self.ModelParams.ReportDayDelta,self.table)
		tb.rotateStats(currentDay,self.ModelParams.ReportDayDelta,self.table)
	}
	if err != nil {
		return flap.UpdateBackfillStats{},0,logError(err)
	}
	return us,flapParams.DailyTotal,nil
}

// Resets state of  model and/or flap engine.
// If destroy is true, all state is destroyed and otherwise only
// state associated with current model run.
func (self *Engine) Reset(destroy bool) error {

	// Delete journey planner state
	dropJourneyPlanner(self.db) 

	// Delete model run state
	self.table.Delete(modelstateRecordKey)
	self.table.Delete(summaryStatsRecordKey)
	self.table.Delete(botStatsRecordKey)

	// Destroy model
	if destroy {
		dropCountriesAirportsRoutes(self.db)
		self.table.Delete(cwFieldName)
	}

	// Delete flap state
	return flap.Reset(self.db,destroy)
}

// Runs the model with configuration as specified in ModelParams, writing results out
// to multiple CSV files in the specified working folder.
func (self *Engine) Run(warmOnly bool, startDay flap.EpochTime) error {
	
	// Override start day if necessary
	var finalStartDay = flap.EpochTime(self.ModelParams.StartDay.Unix())
	if startDay != 0 {
		finalStartDay = startDay
	}

	// Initialize model state and stats
	ms := modelState{startDate:finalStartDay,totalTravellersCurrent:self.ModelParams.TotalTravellers}
	err := ms.save(self.table)
	if err != nil {
		return logError(err)
	}
	var ss summaryStats
	ss.save(self.table)
	var bs botStats
	for i,_ := range self.ModelParams.BotSpecs {
		bs.save(self.table,i)
	}

	// Reset journey planner
	err = self.Reset(false)
	if err != nil {
		return logError(err)
	}

	// Set up data structures
	cars,tb,fe,jp,err := self.prepare()
	if err != nil {
		return logError(err)
	}
	defer fe.Release()
	err = fe.Administrator.SetParams(self.FlapParams)
	if err != nil {
		return logError(err)
	}

	// Calculate number of days needed to warm the model
	_,planDays := self.minmaxTripLength()
	if self.FlapParams.Promises.Algo != 0 {
		planDays += self.FlapParams.Promises.MaxDays
	} 

	// Calculate days to run
	var daysToRun flap.Days
	if !warmOnly {
		daysToRun += self.ModelParams.DaysToRun
	}

	// Model for each of the warm days and the days for the model
	// proper
	currentDay := finalStartDay - flap.EpochTime(uint64(planDays*flap.SecondsInDay))
	flightPaths := newFlightPaths(currentDay)
	logInfo("Running ", planDays, " day prewarm and ", daysToRun," day model")
	for i:=flap.Days(-planDays); i < daysToRun; i++ {
		
		// Run model for one day
		us,dt,err := self.modelDay(currentDay,cars,tb,fe,jp,flightPaths)
		if err != nil {
			return logError(err)
		}

		// Update verbose statistics and report as needed
		self.updateVerboseStats(i,currentDay,dt,us)
		if i % self.ModelParams.VerboseReportDayDelta == 0 {
			flightPaths= reportFlightPaths(flightPaths,currentDay,self.ModelParams.WorkingFolder) 
		}

		// Next day
		currentDay += flap.SecondsInDay
	}

	// Output final charts and finish
	self.reportSummary()
	tb.reportSummary(self.ModelParams)
	fmt.Printf("\nFinished\n")
	return nil
}

// AccountForMoreTravellers adjusts daily total upwards to
// account for new travellers joining FLAP.
func (self *Engine) AdjustDailyTotal(newTravellers uint64, fe *flap.Engine) error {
	
	// Retreive amount to adjust by for each new traveller
	var ss summaryStats
	ss.load(self.table)
	distPerTraveller,err := ss.calculateMeanDaily(self.ModelParams.ReportDayDelta)
	if err != nil {
		return err
	}

	// Adjust daily total
	flapParams := fe.Administrator.GetParams()
	flapParams.DailyTotal += distPerTraveller*flap.Kilometres(newTravellers)
	err = fe.Administrator.SetParams(flapParams)
	return err
}
 
// Report generates summary reports
func (self *Engine) Report() error {
	
	_,tb,_,_,err := self.prepare()
	if err != nil {
		return logError(err)
	}

	// Output final charts and finish
	self.reportSummary()
	tb.reportSummary(self.ModelParams)
	fmt.Printf("\nFinished\n")
	return nil
}

// SummaryStats returns current summary stats as a JSON array (containing one object for each day)
func (self  *Engine) SummaryStats() (string,error) {
	var ss summaryStats
	err := ss.load(self.table)
	if err != nil {
		return "",err
	}
	return ss.asJSON(),nil
}

// RunOneDay runs the model for the one specified day
func (self *Engine) RunOneDay(startOfDay flap.EpochTime) error {

	cars,tb,fe,jp,err := self.prepare()
	if err != nil {
		return logError(err)
	}
	defer fe.Release()

	_,_,err = self.modelDay(startOfDay,cars,tb,fe,jp,nil)
	return err
}

// bandToPassport maps a specfied band and bot number to a passport definition,
// so thast the corresponding traveller record for the bot can be retreived
func (self *Engine) bandToPassport(band uint64,bot uint64) (flap.Passport,error) {

	// Load country weights (need to establish issuing country of passport)
	var p flap.Passport
	cw := newCountryWeights()
	err := cw.load(self.table)
	if err != nil {
		return p,logError(err)
	}

	// load model state
	var ms modelState
	err = ms.load(self.table)
	if err != nil {
		return p,logError(err)
	}

	// Create travellerbots struct
	travellerBots := NewTravellerBots(cw)
	if travellerBots == nil {
		return p,logError(EFAILEDTOCREATETRAVELLERBOTS)
	}
	err = travellerBots.Build(self.ModelParams,self.FlapParams,ms.totalTravellersCurrent,self.table)
	if (err != nil) {
		return p,logError(err)
	}

	// Valid args
	if band >= uint64(len(travellerBots.bots)) {
		return p,logError(ENOSUCHTRAVELLER)
	}
	if bot >= uint64(travellerBots.bots[band].numInstances) {
		return p,logError(ENOSUCHTRAVELLER)
	}

	// Resolve given spec to a passport and look up in the travellers db
	p,err = travellerBots.getPassport(botId{bandIndex(band),botIndex(bot)})
	if err != nil {
		return  p,logError(err)
	}
	return p,nil
}

// TripHistoryAsJSON reports the trip history for the specificied traveller both in JSON and KML format
func (self *Engine) TripHistoryAsJSON(band uint64,bot uint64) (flap.Passport,string,error){

	// Get traveller's passport
	p,err := self.bandToPassport(band,bot)
	if (err != nil) {
		return p,"",err
	}

	//  Initialize flap
	fe := flap.NewEngine(self.db,flap.LogLevel(self.ModelParams.LogLevel),self.ModelParams.WorkingFolder)
	defer fe.Release()
	
	// Resolve passport to traveller
	t,err := fe.Travellers.GetTraveller(p)
	if err != nil {
		return p,"",logError(err)
	}

	// Return the traveller as JSON
	return p,t.AsJSON(),nil
}

type jsonPromise struct {
	TripStart time.Time
	TripEnd time.Time
	Clearance time.Time
	Distance flap.Kilometres
	Stacked flap.StackIndex
	CarriedOver flap.Kilometres
}

// Write out promises for the given traveller as JSON string
func (self *Engine) PromisesAsJSON(band uint64,bot uint64) (string,error) {

	// Get traveller's passport
	p,err := self.bandToPassport(band,bot)
	if (err != nil) {
		return "",err
	}

	//  Initialize flap
	fe := flap.NewEngine(self.db,flap.LogLevel(self.ModelParams.LogLevel),self.ModelParams.WorkingFolder)
	defer fe.Release()
	
	// Resolve passport to traveller
	t,err := fe.Travellers.GetTraveller(p)
	if err != nil {
		return "",logError(err)
	}

	// Render promises as JSON
	promises := make([]jsonPromise,0)
	it := t.Promises.NewIterator()
	for it.Next() {
		p := it.Value()
		promises = append(promises,jsonPromise{TripStart:p.TripStart.ToTime(),TripEnd:p.TripEnd.ToTime(),Clearance:p.Clearance.ToTime(),Distance:p.Distance,Stacked:p.StackIndex,CarriedOver:p.CarriedOver})
	}
	jsonData, _ := json.MarshalIndent(promises, "", "    ")
	return string(jsonData),nil
}

type jsonTransaction struct {
	Date 	time.Time
	Distance  flap.Kilometres
	Type 	flap.TransactionType
}

// Write out promises for the given traveller as JSON string
func (self *Engine) TransactionsAsJSON(band uint64,bot uint64) (string,error) {

	// Get traveller's passport
	p,err := self.bandToPassport(band,bot)
	if (err != nil) {
		return "",err
	}

	//  Initialize flap
	fe := flap.NewEngine(self.db,flap.LogLevel(self.ModelParams.LogLevel),self.ModelParams.WorkingFolder)
	defer fe.Release()
	
	// Resolve passport to traveller
	t,err := fe.Travellers.GetTraveller(p)
	if err != nil {
		return "",logError(err)
	}

	// Render transactions as JSON
	transactions := make([]jsonTransaction,0)
	it := t.Transactions.NewIterator()
	for it.Next() {
		t := it.Value()
		transactions = append(transactions,jsonTransaction{Date:t.Date.ToTime(),Distance: t.Distance, Type:t.TT})
	}
	jsonData, _ := json.MarshalIndent(transactions, "", "    ")
	return string(jsonData),nil
}


type jsonAccount struct {
	Balance  flap.Kilometres
	Cleared  flap.ClearanceReason
	ClearanceDate time.Time 
}

// Write out promises for the given traveller as JSON string
func (self *Engine) AccountAsJSON(band uint64,bot uint64, now flap.EpochTime) (string,error) {

	// Get traveller's passport
	p,err := self.bandToPassport(band,bot)
	if (err != nil) {
		return "",err
	}

	//  Initialize flap
	fe := flap.NewEngine(self.db,flap.LogLevel(self.ModelParams.LogLevel),self.ModelParams.WorkingFolder)
	defer fe.Release()
	
	// Resolve passport to traveller
	t,err := fe.Travellers.GetTraveller(p)
	if err != nil {
		return "",logError(err)
	}

	// Render account state as JSON
	var account jsonAccount
	account.Balance = t.Balance
	account.Cleared = t.Cleared(now)
	account.ClearanceDate =  t.Kept.Clearance.ToTime()
	jsonData, _ := json.MarshalIndent(account, "", "    ")
	return string(jsonData),nil
}

// updateVerboseStats updates the data for verbose stats to reflect current day and outputs if it is time to do so
func (self *Engine) updateVerboseStats(day flap.Days, currentDay flap.EpochTime, dt flap.Kilometres, us flap.UpdateBackfillStats) {
	
	// Update verbose stats
	for _,v := range us.ClearedDistanceDeltas {
		self.verbose.clearedDistanceDeltas = append(self.verbose.clearedDistanceDeltas, float64(v))
	}
	for _,u := range us.ClearedDaysDeltas {
		self.verbose.clearedDaysDeltas = append(self.verbose.clearedDaysDeltas, float64(u))
	}


	// Output verbose data if requests
	if self.ModelParams.VerboseReportDayDelta > 0 && day % self.ModelParams.VerboseReportDayDelta == 0 {

		// Clearance Days Deltas Distribution
		self.reportDistribution(self.verbose.clearedDaysDeltas,10,24, "cleareddaysdistro",currentDay,day)

		// Clearance Distance Deltas Distribution
		self.reportDistribution(self.verbose.clearedDistanceDeltas,500,24,"cleareddistancedistro",currentDay,day)

		// Regression accuracy
		self.reportRegression(us.BestFitPoints,us.BestFitConsts,"regression",currentDay,day)

		// Wipe stats
		self.verbose.reset()
	}
}

// reportSummary generates summary line chart comparing daily allowance with distance travelled
// over time
func (self* Engine) reportSummary() {

	// Retreive stats
	var ss summaryStats
	ss.load(self.table)
	travelledPts,allowancePts := ss.compile(self.ModelParams.WorkingFolder,self.ModelParams.ReportDayDelta)
	
	// Set axis labels
	p, err := plot.New()
	if err != nil {
		return
	}
	p.X.Label.Text = "Day"
	p.Y.Label.Text = "Distance (km)"

	// Add the source points for each band
	line, err := plotter.NewLine(allowancePts)
	line.LineStyle.Width = 1
 	line.LineStyle.Dashes = []vg.Length{10}
     	line.LineStyle.DashOffs =  vg.Length(2)
	line.LineStyle.Color = color.Black
	p.Add(line)
	_,_,_,ymax := plotter.XYRange(allowancePts)
	if ymax > p.Y.Max {
		p.Y.Max = ymax
	}
	line, err = plotter.NewLine(travelledPts)
	p.Add(line)
	line.LineStyle.Width = 1
	palette,err := brewer.GetPalette(brewer.TypeSequential,"Reds",3)
	if err != nil {
		return 
	}
	line.LineStyle.Color = palette.Colors()[2]
	p.Add(line)
	_,_,_,ymax = plotter.XYRange(travelledPts)
	if ymax > p.Y.Max {
		p.Y.Max = ymax
	}

	// Set the axis ranges
	p.X.Max= float64(self.ModelParams.DaysToRun)
	p.X.Min = 1
	p.Y.Min = 0

	// Save the plot to a PNG file.
	fp := filepath.Join(self.ModelParams.WorkingFolder,"summary.png")
	w:= vg.Length(self.ModelParams.LargeChartWidth)*vg.Centimeter
	p.Save(w, w/2, fp); 
}

// reportRegression reports the match between given set of points and result
// of linear regression against those points as a png raster image.
func (self* Engine) reportRegression(y []float64, consts []float64, foldername string,currentDay flap.EpochTime, day flap.Days) {

	// Check for something to plot
	if len(consts) == 0 {
		return
	}

	// Set labels and titles
	p, err := plot.New()
	if err != nil {
		return
	}
	p.X.Label.Text = "Days since 1970-01-01"
	p.Y.Label.Text = "Daily Share (km)"
	t := currentDay.ToTime()
	p.Title.Text = fmt.Sprintf("Day %d",day)

	// Define function based on consts
	f := plotter.NewFunction(
		func(x float64) float64 {
			var y float64
			for i,v := range consts {
				y += math.Pow(x,float64(i))*v
			}
			return y
		})
	f.LineStyle.Width = 2
 	f.LineStyle.Dashes = []vg.Length{10}
     	f.LineStyle.DashOffs =  vg.Length(2)
	f.LineStyle.Color = color.Black
	
	// Add the source points
	pts := make(plotter.XYs, len(y))
	firstDay := float64(currentDay/flap.SecondsInDay) - float64(len(y))
	lastDay := firstDay
	for i := range pts {
		pts[i].X = lastDay 
		pts[i].Y = y[i]
		lastDay++
	}
	actual, err := plotter.NewLine(pts)
	if err == nil {
		p.Add(actual)
	}
	palette,err := brewer.GetPalette(brewer.TypeSequential,"Reds",3)
	if err != nil {
		return 
	}
	actual.LineStyle.Width=2
	actual.LineStyle.Color = palette.Colors()[2]

	// Add the calculated regression line
	p.Add(f)

	// Set the axis ranges
	topY := floats.Max(y)
	p.X.Max= firstDay
	p.X.Min = lastDay
	p.Y.Min = -0.25*topY
	p.Y.Max = topY + 0.25*topY

	// Make sure the folder exists
	folder := filepath.Join(self.ModelParams.WorkingFolder,foldername)
	os.MkdirAll(folder, os.ModePerm)

	// Save the plot to a PNG file.
	fn := t.Format("2006-01-02") + ".png"
	fp := filepath.Join(folder,fn)
	w:= vg.Length(self.ModelParams.ChartWidth)*vg.Centimeter
	p.Save(w, w/2, fp); 
}

// reportDistribution reports a distribution of the given points with given bin size
// both as a csv and png raster image.
func (self* Engine) reportDistribution(x []float64, binSize float64, maxBars int, foldername string, currentDay flap.EpochTime,day flap.Days) {

	// Check for values to build histogram from
	if len(x) == 0 {
		return
	}

	// Create initial buckets
	sort.Float64s(x)	
	min := float64(int64(floats.Min(x))/int64(binSize))*binSize - binSize 
	max := float64(int64(floats.Max(x))/int64(binSize))*binSize + binSize*2 
	dividers := make([]float64, int((max-min)/binSize+1))
	if len(dividers) < 2 {
		return
	}
	floats.Span(dividers, min, max)

	// Merge the lowest buckets if we are above max bars
	for ; len(dividers) > maxBars +1 ;  {
		dividers = dividers[1:]
		dividers[0] = -math.MaxFloat64 
	}

	// Generate distribution
	hist := stat.Histogram(nil, dividers, x, nil)

	// Make sure there is a folder to write to
	folder := filepath.Join(self.ModelParams.WorkingFolder,foldername)
	os.MkdirAll(folder, os.ModePerm)

	// Open csv file
	t := currentDay.ToTime()
	fn := t.Format("2006-01-02") + ".csv"
	fp := filepath.Join(folder,fn)
	fh,_ := os.Create(fp)

	// Write out histogram
	if fh != nil {
		for _,val := range hist  {
			line := fmt.Sprintf("%d\n",int(val))
			fh.WriteString(line)
		}
		defer fh.Close()
	}

	// Build the bars
	imagewidth:= vg.Length(self.ModelParams.ChartWidth)*vg.Centimeter
	bars, err := plotter.NewBarChart(plotter.Values(hist),imagewidth/vg.Length(maxBars+2))
	if err == nil {
		palette,err := brewer.GetPalette(brewer.TypeSequential,"Reds",3)
		if err != nil {
			return 
		}
		bars.LineStyle.Width = vg.Length(1)
		bars.Color = palette.Colors()[2]

		// Build labels
		labels := make([]string,0,len(dividers))
		for _,d := range dividers {
			if d == -math.MaxFloat64 {
				labels = append(labels,"...")
			} else {
				labels = append(labels,fmt.Sprintf("%d",int(d)))
			}
		}

		// Add to a chart
		p, err := plot.New()
		if err == nil {
			p.Title.Text = fmt.Sprintf("Day %d - Day %d", (day+1) - self.ModelParams.VerboseReportDayDelta,day)
			p.Y.Label.Text = "Completed Trips"
			p.X.Label.Text = "Balance At Clearance (km)"
			p.Add(bars)
			p.NominalX(labels...)
			p.X.Tick.Label.Rotation = math.Pi/2 
			p.X.Tick.Label.YAlign = draw.YCenter
			p.X.Tick.Label.XAlign = draw.XRight
			
			// Write to a file
			t := currentDay.ToTime()
			fn := t.Format("2006-01-02") + ".png"
			fp := filepath.Join(folder,fn)
			p.Save(imagewidth, imagewidth/2, fp); 
		}
	}
}

// Calculates RSquared - measurment of fit - between given points and consts for results of regression
// to fit those points
func (self* Engine) calcRSquared(values []float64, consts []float64,currentDay flap.EpochTime) float64 {

	// Check for something to plot
	if len(consts) == 0 {
		return 0
	}

	// Calculate estimated values for given points
	estimates := make([]float64,0)
	xStart := float64(currentDay/flap.SecondsInDay) 
	xEnd := xStart + float64(len(values))
	for x := xStart; x < xEnd; x++ {
		var t float64
		for i,v := range(consts) {
			t += math.Pow(float64(x),float64(i))*v
		}
		estimates = append(estimates,t)
	}

	// Calculate RSquared
	return stat.RSquaredFrom(estimates, values, nil)

}
