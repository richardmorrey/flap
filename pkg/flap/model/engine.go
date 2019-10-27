package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"github.com/richardmorrey/flap/pkg/flap/db"
	"path/filepath"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
	"math/rand"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var EFAILEDTOCREATECOUNTRIESAIRPORTSROUTES = errors.New("Failed to create Countries-Airports-Routes")
var EFAILEDTOCREATETRAVELLERBOTS = errors.New("Failed to create traveller bots")
var EMODELNOTBUILT = errors.New("Failed to find model to load")
var ESTARTDAYNOTWHOLEDAYS = errors.New("Start day must be EpochTime representing start of UTC day")
var ENOSUCHTRAVELLER = errors.New("No such traveller")

type Probability float64

type BotSpec struct {
	PlanProbability		Probability
	//planEarliest		flap.Days
	//planPeriod		flap.Days
	Weight			weight
}

type ModelParams struct {
	WorkingFolder		string
	DataFolder		string
	TrialDays		flap.Days
	DTAlgo			string
	DaysToRun		flap.Days
	TotalTravellers		uint64
	BotSpecs		[]BotSpec
	TripLengths		[]flap.Days
	StartDay		flap.EpochTime
	DailyTotalFactor	float64
}

type plannedFlight struct {
	from			flap.ICAOCode
	to			flap.ICAOCode
}

type Engine struct {
	FlapParams 			flap.FlapParams
	ModelParams			ModelParams
	plannedFlights			[]plannedFlight
	db				*db.LevelDB
	fh				*os.File
}

var glogger *log.Logger
func glog(e error) error {
	if glogger != nil {
		glogger.Output(2,e.Error())
	}
	return e
}

// NewEngine is factory function for Engine
func NewEngine(configFilePath string) (*Engine,error) {
	
	e:= new(Engine)

	// Set a simple config
/* e.FlapParams = flap.FlapParams{TripLength:365,FlightsInTrip:50,FlightInterval:1,DailyTotal:0,MinGrounded:0}
	e.ModelParams = ModelParams{WorkingFolder:"/home/spencerthehalfwit/working",DataFolder:"/home/spencerthehalfwit/flapdata", DaysToRun:1000, TotalTravellers:1000,TrialDays:100,DailyTotalFactor:0.999}
	e.ModelParams.BotSpecs = []BotSpec{{0.02,1},{0.002,9}}
	e.ModelParams.TripLengths = []flap.Days{1,2,3,5,7,7,7,7,14,14,14,14,28}
	buff,err:=yaml.Marshal(e)
    	err = ioutil.WriteFile("config.yaml", buff, 0644)
*/
	// Load config file
	buff, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil,glog(err)
	}
	err = yaml.Unmarshal(buff, &e)
	if err != nil {
		return nil,glog(err)
	}

	// Initialize logger
	logpath := filepath.Join(e.ModelParams.WorkingFolder,"model.log")
	f, _ := os.OpenFile(logpath,os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	glogger = log.New(f, "model ", log.LstdFlags | log.Lshortfile)

	// Create db
	e.db = db.NewLevelDB(e.ModelParams.WorkingFolder)

	// Seed Random Number Generator
	rand.Seed(time.Now().UTC().UnixNano())
	return e,nil
}

//Release releases all resources that need to be explicitly released when finished with an Engine
func (self *Engine) Release() {
	self.db.Release()
}

// Build prepares all persitent data files in order to be able to run the model in the configured data folder. It
// (a) Builds the countries-airports-routes and the country weights file that drive flight selection.
// (c) Ensures the Flap library Travellers Table is empty
func (self *Engine) Build() error {
	
	fmt.Printf("Building...\n")

	//  Reset flap and load airports
	err := flap.Reset(self.db)
	fe := flap.NewEngine(self.db)
	err = fe.Administrator.SetParams(self.FlapParams)
	if (err != nil) {
		return glog(err)
	}
	err =fe.Airports.LoadAirports(filepath.Join(self.ModelParams.DataFolder,"airports.dat"))
	if (err != nil) {
		return glog(err)
	}

	// Build countries-airports-flights table from real-world data
	cars := NewCountriesAirportsRoutes(self.db)
	if cars  == nil {
		return EFAILEDTOCREATECOUNTRIESAIRPORTSROUTES
	}
	err = cars.Build(self.ModelParams.DataFolder,self.ModelParams.WorkingFolder)
	if (err != nil) {
		return glog(err)
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

// Runs the model with configuration as specified in ModelParams, writing results out
// to multiple CSV files in the specified working folder.
func (self *Engine) Run() error {

	// Validate model params
	if self.ModelParams.StartDay % flap.SecondsInDay != 0 {
		return ESTARTDAYNOTWHOLEDAYS
	}

	// Load country-airports-routes model
	cars := NewCountriesAirportsRoutes(self.db)
	if cars == nil {
		return EMODELNOTBUILT
	}

	// Load country weights
	var countryWeights CountryWeights
	err := countryWeights.load(self.ModelParams.WorkingFolder)
	if err != nil {
		return glog(err)
	}
	
	// Build flight plans for traveller bots
	travellerBots := NewTravellerBots(&countryWeights)
	if travellerBots == nil {
		return EFAILEDTOCREATETRAVELLERBOTS
	}
	err = travellerBots.Build(&(self.ModelParams))
	if (err != nil) {
		return glog(err)
	}
	
	// Reset flap and load airports
	err = flap.Reset(self.db)
	fe := flap.NewEngine(self.db)
	err = fe.Administrator.SetParams(self.FlapParams)
	if (err != nil) {
		return glog(err)
	}
	err =fe.Airports.LoadAirports(filepath.Join(self.ModelParams.DataFolder,"airports.dat"))
	if (err != nil) {
		return glog(err)
	}

	// Model each day
	jp,err := NewJourneyPlanner(self.ModelParams.TripLengths)
	if (err != nil) {
		return glog(err)
	}
	maxTripLength,_ := jp.minmaxTripLength()
	currentDay := self.ModelParams.StartDay
	for i:=flap.Days(1); i <= self.ModelParams.DaysToRun; i++ {
		
		fmt.Printf("\rDay %d: Planning Flights",i)
		err = travellerBots.planTrips(cars,jp,fe.Travellers)
		if err != nil {
			return glog(err)
		}

		fmt.Printf("\rDay %d: Submitting Flights",i)
		err = jp.submitFlights(travellerBots,fe,currentDay,i>self.ModelParams.TrialDays)
		if err != nil && err != ENOJOURNEYSPLANNED {
			return glog(err)
		}

		fmt.Printf("\rDay %d: Backfilling       ",i)
		currentDay += flap.SecondsInDay
		t,d,g,err :=  fe.UpdateTripsAndBackfill(currentDay)
		if err != nil {
			return glog(err)
		}
		
		// Report daily stats
		travellerBots.ReportDay(i)
		self.reportDay(i,self.FlapParams.DailyTotal,t,d,g)

		// Calculate/set new daily total
		if (i<= self.ModelParams.TrialDays) {
			
			// Calculate DT as average across trial days if specified, defaulting to minimum
			if self.ModelParams.DTAlgo=="average" {
				self.FlapParams.DailyTotal += flap.Kilometres(float64(d)/float64(self.ModelParams.TrialDays))
			} else {
				self.FlapParams.DailyTotal=max(self.FlapParams.DailyTotal,d)
			}

			// Update MinGrounded, skipping until maxTripLength has been reached
			// so we get full coverage of return journeys
			if i > maxTripLength {
				self.FlapParams.MinGrounded = min(self.FlapParams.MinGrounded,t)
			}
		} else  {

			// Adjust Daily Total for the next day
			self.FlapParams.DailyTotal = flap.Kilometres(float64(self.FlapParams.DailyTotal)*self.ModelParams.DailyTotalFactor)
			err = fe.Administrator.SetParams(self.FlapParams)
			if err != nil {
				return glog(err)
			}
		}
	}
	fmt.Printf("\nFinished\n")
	return nil
}

// ShowTraveller reports the trip history for the specificied traveller bot in JSON format
func (self *Engine) ShowTraveller(band uint64,bot uint64) (string,error){

	// Load country weights (need to establish issuing country of passport)
	var countryWeights CountryWeights
	err := countryWeights.load(self.ModelParams.WorkingFolder)
	if err != nil {
		return "",glog(err)
	}

	// Create travellerbots struct
	travellerBots := NewTravellerBots(&countryWeights)
	if travellerBots == nil {
		return "",glog(EFAILEDTOCREATETRAVELLERBOTS)
	}
	err = travellerBots.Build(&(self.ModelParams))
	if (err != nil) {
		return "",glog(err)
	}

	// Valid args
	if band >= uint64(len(travellerBots.bots)) {
		return "",glog(ENOSUCHTRAVELLER)
	}
	if bot >= uint64(travellerBots.bots[band].numInstances) {
		return "",glog(ENOSUCHTRAVELLER)
	}

	//  Initialize flap
	fe := flap.NewEngine(self.db)
	err = fe.Administrator.SetParams(self.FlapParams)
	if (err != nil) {
		return "",glog(err)
	}

	// Resolve given spec to a passport and look up in the travellers db
	p,err := travellerBots.getPassport(botId{bandIndex(band),botIndex(bot)})
	if err != nil {
		return  "",glog(err)
	}
	fmt.Printf("\nSearching for Passport Number:%s, Issuer:%s\n", string(p.Number[:]),string(p.Issuer[:]))
	t,err := fe.Travellers.GetTraveller(p)
	if err != nil {
		return "", glog(err)
	}

	// Return the traveller as JSON
	return t.AsJSON(), nil
}

// reportDay reports daily total set for the day as well as total distance
// travelled and total travellers travelling
func (self *Engine) reportDay(day flap.Days, dt flap.Kilometres, t uint64,d flap.Kilometres,g uint64) {

	// Open file
	if self.fh == nil{
		fn := filepath.Join(self.ModelParams.WorkingFolder,"summary.csv")
		self.fh,_ = os.Create(fn)
		if self.fh != nil {
			self.fh.WriteString("Day,DailyTotal,Travelled,Travellers,Grounded\n")
		}
	}

	// Write line
	if day <= self.ModelParams.TrialDays {
		dt =0 
	}
	if self.fh != nil {
		line := fmt.Sprintf("%d,%d,%d,%d,%d\n",day,dt,d,t,g)
		self.fh.WriteString(line)
	}
}
