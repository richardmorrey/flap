package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"os"
	"strings"
	"sync"
	"gonum.org/v1/plot/plotter"
	"strconv"
)

var ECOULDNTFINDWEIGHT= errors.New("Couldnt find country weight for bot")
type bandIndex uint8
type botIndex uint32

type travellerBot struct {
	countryStep    float64
	numInstances   botIndex
	planner	       botPlanner
	stats	       botStats
	cancelledPts   plotter.XYs
	travelledPts   plotter.XYs
}

type botId struct {
	band  bandIndex
	index botIndex
}

// fromPassport derives botid from passport
func (self *botId) fromPassport(p flap.Passport) error {
	n, err := strconv.ParseUint(string(p.Number[:2]), 10, 8)
	if err != nil {
		return err
	}
	self.band = bandIndex(n)
	n, err = strconv.ParseUint(string(p.Number[2:]), 10, 32)
	if err != nil {
		return err
	}
	self.index = botIndex(n)
	return nil
}

// getPassport calculates passport, including 
// passport number, deterministically, from the bot
// band and index within band. This saves holding
// full passport details in memory for all planned
// journeys
func (self *TravellerBots) getPassport(bot botId) (flap.Passport,error) {
	numberStr:= fmt.Sprintf("%02d%07d",bot.band,bot.index)
	var p flap.Passport
	copy(p.Number[:],numberStr)
	weightIn := weight(self.bots[bot.band].countryStep*float64(bot.index))
	w,err := self.countryWeights.find(weightIn)
	if err != nil {
		return p,err
	}
	copy(p.Issuer[:],self.countryWeights.Countries[w])
	return p,nil
}

type botStats struct {
	distance flap.Kilometres
	flightsTaken	 uint64
	flightsRefused   uint64
	tripsCancelled   uint64
	tripsPlanned	 uint64
	mux		 sync.Mutex
}

// Submitted updates stats to reflect fact a journey
// has been successfully submitted
func (self *botStats) Submitted(dist flap.Kilometres) {
	self.mux.Lock()
	defer self.mux.Unlock()
	self.flightsTaken++
	self.distance += dist
}

// Refused updates stats to relect fact a journey 
// submission has been refused by flap
func (self  *botStats) Refused() {
	self.mux.Lock()
	defer self.mux.Unlock()
	self.flightsRefused++
}

// Canclled updates stats to relect fact a promise
// request has been refused by flap
func (self  *botStats) Cancelled() {
	self.mux.Lock()
	defer self.mux.Unlock()
	self.tripsCancelled++
}

// Planned updates stats to relect fact a promise
// request has been refused by flap
func (self  *botStats) Planned() {
	self.mux.Lock()
	defer self.mux.Unlock()
	self.tripsPlanned++
}

// Report writes a single CSV row with stats
func (self *travellerBot) ReportDay(day flap.Days) string {
	
	// One thread at a time
	self.stats.mux.Lock()
	defer self.stats.mux.Unlock()

	// Format stats
	var cancelled float64
	if (self.stats.tripsPlanned+self.stats.tripsCancelled) > 0 {
		cancelled = (float64(self.stats.tripsCancelled)/float64(self.stats.tripsPlanned+self.stats.tripsCancelled))*100
	}
	distance := float64(self.stats.distance)/float64(flap.Kilometres(self.numInstances))
	line := fmt.Sprintf("%f,%f,%f,",
		(float64(self.stats.flightsRefused)/float64(self.stats.flightsTaken+self.stats.flightsRefused))*100,
		cancelled,distance)

	// Add points for summary graphs
	if self.travelledPts == nil {
		self.travelledPts = make(plotter.XYs, 0)
	}
	self.travelledPts = append(self.travelledPts,plotter.XY{X:float64(day),Y:distance})
	if self.cancelledPts == nil {
		self.cancelledPts = make(plotter.XYs, 0)
	}
	self.cancelledPts = append(self.cancelledPts,plotter.XY{X:float64(day),Y:cancelled})

	// Reset counters
	self.stats.flightsTaken = 0
	self.stats.flightsRefused = 0
	self.stats.distance = 0
	self.stats.tripsCancelled = 0
	self.stats.tripsPlanned = 0
	return line
}

type TravellerBots struct {
	bots	[]travellerBot
	countryWeights *countryWeights
	fh		*os.File
	statsFolder    string
	tripLengths	[]flap.Days
}

func NewTravellerBots(cw *countryWeights) *TravellerBots {
	tbs := new(TravellerBots)
	tbs.bots = make([]travellerBot,0,10)
	tbs.countryWeights=cw
	return tbs
}

func (self *TravellerBots) GetBot(id botId) *travellerBot {
	return &(self.bots[id.band])
}

// ReportStats reports daily stats for each bot band into
// an appropriately named csv file in the working folder
func (self *TravellerBots) ReportDay(day flap.Days, rdd flap.Days) {

	if day %  rdd == 0 {

		// Open file if not open
		if self.fh == nil{
			fn := filepath.Join(self.statsFolder,"bands.csv")
			self.fh,_ = os.Create(fn)
			if self.fh != nil {
				line:="Day"
				for bb := bandIndex(0); bb < bandIndex(len(self.bots)); bb++ {
					line += fmt.Sprintf(",refusedpercent_%d,cancelledpercent_%d,distance_%d",bb,bb,bb) 
				}
				line +="\n"
				self.fh.WriteString(line)
			}
		}

		// Write line
		line := fmt.Sprintf("%d,",day)
		for bb := bandIndex(0); bb < bandIndex(len(self.bots)); bb++ {
			line += self.bots[bb].ReportDay(day)
		}
		line=strings.TrimRight(line,",")
		line+="\n"
		self.fh.WriteString(line)
	}
}

// Build constructs bot configurations for each band from provided model params
func (self *TravellerBots) Build(modelParams ModelParams,flapParams flap.FlapParams) error {

	// Check arguments
	if (len(modelParams.BotSpecs) == 0) {
		return flap.EINVALIDARGUMENT
	}
	self.statsFolder=modelParams.WorkingFolder

	// Calculate total bot weight
	var weightTotal  weight
	for _, botspec := range modelParams.BotSpecs {
		weightTotal += botspec.Weight
	}

	// Create bots
	topWeight, err := self.countryWeights.topWeight()
	if err != nil {
		return logError(err)
	}
	for _, botspec := range modelParams.BotSpecs {
		var bot travellerBot
		bot.numInstances= botIndex((float64(botspec.Weight)/float64(weightTotal))*float64(modelParams.TotalTravellers))
		if (bot.numInstances > 0) {
			bot.countryStep= float64(topWeight)/float64(bot.numInstances)
		}
		if flapParams.Promises.Algo &^ flap.PromisesAlgo(0xf0) != 0 {
			bot.planner = new(promisesPlanner)
		} else {
			bot.planner = new(simplePlanner)
		}
		bot.planner.build(botspec,flapParams)
		if (err != nil) {
			return logError(err)
		}
		self.bots  = append(self.bots,bot)
	}

	// Store trip lengths
	self.tripLengths =  modelParams.TripLengths
	return nil
}

// planTrips "throws dice" for every traveller bot in every band according to probability
// of travellers in the band travelling on any one day. If the dice comes up and the
// travellerbot is not in the middle of a trip, a new trip is planned using weighted 
// country-airports-routes model
func (self *TravellerBots) planTrips(cars *CountriesAirportsRoutes, jp* journeyPlanner, fe *flap.Engine,currentDay flap.EpochTime,deterministic bool, threads uint) error {
	
	// Create configured number of threads to plan trips and wait for them to finish
	perrs := make(chan error, threads)
	var wg sync.WaitGroup
	for i := uint(0); i < threads; i++ {
		wg.Add(1)
		t :=  func (step uint,offset uint) {perrs <- self.doPlanTrips(cars,jp,fe,currentDay,deterministic,step,offset);wg.Done()}
		go t(threads,i)
	}
	wg.Wait()

	// Return first error reported
	close(perrs)
	for elem := range perrs {
		if elem != nil {
			return elem
		}
	}
	return nil
}

func (self *TravellerBots) doPlanTrips(cars *CountriesAirportsRoutes, jp* journeyPlanner, fe *flap.Engine,currentDay flap.EpochTime,deterministic bool,threads uint, offset uint) error {

	// Iterate through each bot in each band
	for i:=bandIndex(0); i < bandIndex(len(self.bots)); i++ {
		logDebug("Started planning  band ",i," start ",offset," step ",threads)
		planner := self.bots[i].planner.clone()
		for j:=botIndex(offset); j < self.bots[i].numInstances; j+=botIndex(threads) {

			// Retrieve passport
			p,err := self.getPassport(botId{i,j})
			if err != nil {
				return logError(err)
			}
			
			// Choose trip length
			tripLength := self.tripLengths[0]
			if len(self.tripLengths) > 1 {
				tripLength = self.tripLengths[rand.Intn(len(self.tripLengths)-1)]
			}

			// Decide whether to plan a trip
			if planner.areWePlanning(fe,p,currentDay,tripLength) {

				// Choose trip
				from,to,err := cars.chooseTrip(p)
				if err != nil {
					return logError(err)
				}

				// Decide if the trip is allowed ...
				ts,err := planner.whenWillWeFly(fe,p,currentDay,from,to,tripLength)
				switch (err) {
					case nil: 
						err = jp.planTrip(from,to,tripLength,p,ts,fe)
						if err != nil {
							return logError(err)
						} else {
							self.bots[i].stats.Planned()
						}
					case ENOSPACEFORTRIP:
						self.bots[i].stats.Cancelled() 
					case ENOTPLANNINGTODAY:
						break
					default:
						return logError(err)
				}
			}
		}
		logDebug("Finished planning  band ",i," start ",offset," step ",threads)
	}
	return nil
} 

