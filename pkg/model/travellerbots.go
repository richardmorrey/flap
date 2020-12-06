package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"github.com/richardmorrey/flap/pkg/db"
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"os"
	"strings"
	"sync"
	"strconv"
	"encoding/gob"
	"bytes"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/palette/brewer"
)

var ECOULDNTFINDWEIGHT= errors.New("Couldnt find country weight for bot")
type bandIndex uint8
type botIndex uint32

type travellerBot struct {
	countryStep    float64
	numInstances   botIndex
	planner	       botPlanner
	stats	       botStats
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

type botStatsRow struct {
	Distance flap.Kilometres
	FlightsTaken	 uint64
	FlightsRefused   uint64
	TripsCancelled   uint64
	TripsPlanned	 uint64
	Date		flap.EpochTime
	Entries		int
}

type botStats struct {
	Rows []botStatsRow
	mux  sync.Mutex
}

// newRow starts a new record for counting stats
func (self *botStats) newRow() {
	self.Rows = append(self.Rows,botStatsRow{})
}

// Submitted updates stats to reflect fact a journey
// has been successfully submitted
func (self *botStats) Submitted(dist flap.Kilometres) {
	self.mux.Lock()
	defer self.mux.Unlock()
	self.Rows[len(self.Rows)-1].FlightsTaken++
	self.Rows[len(self.Rows)-1].Distance += dist
}

// Refused updates stats to relect fact a journey 
// submission has been refused by flap
func (self  *botStats) Refused() {
	self.mux.Lock()
	defer self.mux.Unlock()
	self.Rows[len(self.Rows)-1].FlightsRefused++
}

// Canclled updates stats to relect fact a promise
// request has been refused by flap
func (self  *botStats) Cancelled() {
	self.mux.Lock()
	defer self.mux.Unlock()
	self.Rows[len(self.Rows)-1].TripsCancelled++
}

// Planned updates stats to relect fact a promise
// request has been refused by flap
func (self  *botStats) Planned() {
	self.mux.Lock()
	defer self.mux.Unlock()
	self.Rows[len(self.Rows)-1].TripsPlanned++
}

// rotate rotates row if needed and saves stats 
func (self *botStats) rotate(date flap.EpochTime, rdd flap.Days, t db.Table) {
	i:= len(self.Rows) -1
	self.Rows[i].Date = date
	self.Rows[i].Entries +=1
	if self.Rows[i].Entries == int(rdd) {
		self.newRow()
	}
	self.save(t,i)
}
func (self* botStats) To(b *bytes.Buffer) error {
	enc := gob.NewEncoder(b) 
	return enc.Encode(self)
}

func (self* botStats) From(b *bytes.Buffer) error {
	dec := gob.NewDecoder(b)
	return dec.Decode(self)
}

const botStatsRecordKey="botstats"

// load loads bot stats from given table
func (self *botStats) load(t db.Table, index int) error {
	key := fmt.Sprintf("%s-%2d",botStatsRecordKey,index)
	err :=  t.Get(key,self)
	if len(self.Rows) == 0 {
		self.newRow()
	}
	return err
}

// save saves bot stats to given table
func (self *botStats)  save(t db.Table, index int) error {
	key := fmt.Sprintf("%s-%2d",botStatsRecordKey,index)
	return t.Put(key,self)
}

type botStatsCompiled struct {
	lines []string
	travelled plotter.XYs
	cancelled plotter.XYs
}

// compile collates stats across all days for reporting
func (self *botStats) compile(numInstances botIndex, rdd flap.Days) botStatsCompiled {
	
	// One thread at a time
	self.mux.Lock()
	defer self.mux.Unlock()

	// Create items to return
	travelledPts := make(plotter.XYs, 0)
	cancelledPts := make(plotter.XYs, 0)
	var lines []string
	for i,row := range self.Rows { 

		// Skip incomplete rows
		if row.Entries != int(rdd) {
			continue
		}

		// Format stats
		var day = float64(i*int(rdd))
		var cancelled float64
		if (row.TripsPlanned+row.TripsCancelled) > 0 {
			cancelled = (float64(row.TripsCancelled)/float64(row.TripsPlanned+row.TripsCancelled))*100
		}
		distance := float64(row.Distance)/float64(flap.Kilometres(numInstances))
		line := fmt.Sprintf("%f,%f,%f,",
			(float64(row.FlightsRefused)/float64(row.FlightsTaken+row.FlightsRefused))*100,
			cancelled,distance)
		lines = append(lines,line)
		travelledPts = append(travelledPts,plotter.XY{X:day,Y:distance})
		cancelledPts = append(cancelledPts,plotter.XY{X:day,Y:cancelled})
	}
	return botStatsCompiled{lines, travelledPts, cancelledPts}
}

type TravellerBots struct {
	bots	[]travellerBot
	countryWeights *countryWeights
	fh		*os.File
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

// rotateStats adds new bot stats records if necessary and persists
func (self *TravellerBots) rotateStats(date flap.EpochTime,rdd flap.Days,t db.Table) {
	for i:= 0 ; i < len(self.bots); i++ {
		self.bots[i].stats.rotate(date,rdd,t)
	}
}

// Build constructs bot configurations for each band from provided model params
func (self *TravellerBots) Build(modelParams ModelParams,flapParams flap.FlapParams, t db.Table) error {

	// Check arguments
	if (len(modelParams.BotSpecs) == 0) {
		return flap.EINVALIDARGUMENT
	}

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
	for i, botspec := range modelParams.BotSpecs {
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
		bot.stats.load(t,i)
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


// RepotBandsSummary outputs csv and charts summarising band stats over the course of the whoel model
func (self *TravellerBots) reportSummary(mp ModelParams) {

	// Compile results
	var compiledBands []botStatsCompiled
	for _,bot := range(self.bots) {
		compiledBands = append(compiledBands, bot.stats.compile(bot.numInstances, mp.ReportDayDelta))
	}
	
	// Output CSV
	fn := filepath.Join(mp.WorkingFolder,"bands.csv")
	fh,_ := os.Create(fn)
	if fh == nil {
		return
	}
	line:="Day"
	for i,_ := range(compiledBands) {
		line += fmt.Sprintf(",refusedpercent_%d,cancelledpercent_%d,distance_%d",i,i,i) 
	}
	line +="\n"
	fh.WriteString(line)
	for i:=0; i < len(compiledBands[0].lines); i++ {
		lineOut := fmt.Sprintf("%d,",flap.Days(i+1) *mp.ReportDayDelta)
		for _,compiled := range(compiledBands) {
			lineOut += compiled.lines[i]
		}
		lineOut=strings.TrimRight(lineOut,",")
		lineOut+="\n"
		fh.WriteString(lineOut)
	}

	// Output graphs
	reportBandsDistance(compiledBands,mp)
	reportBandsCancelled(compiledBands,mp)
}

// reportBandsDistance generates filled line charts covering the covered bands for distance travelled
// over time.
func reportBandsDistance(compiled []botStatsCompiled, mp ModelParams) {

	// Set axis labels
	p, err := plot.New()
	if err != nil {
		return
	}
	p.X.Label.Text = "Day"
	p.Y.Label.Text = "Distance (km)"

	// Add the source points for each band
	palette,err := brewer.GetPalette(brewer.TypeSequential,"Reds",len(compiled))
	if err != nil {
		return 
	}
	for i,compiledBand := range(compiled) {
		line, err := plotter.NewLine(compiledBand.travelled)
		if err != nil {
			return
		}
		line.FillColor = palette.Colors()[i]
		p.Add(line)

		// Update y ranges
		_,_,_,ymax := plotter.XYRange(compiledBand.travelled)
		if ymax > p.Y.Max {
			p.Y.Max = ymax
		}
	}

	// Set the axis ranges
	p.X.Max= float64(mp.DaysToRun)
	p.X.Min = 1
	p.Y.Min = 0

	// Save the plot to a PNG file.
	fp := filepath.Join(mp.WorkingFolder,"distancebyband.png")
	w:= vg.Length(mp.LargeChartWidth)*vg.Centimeter
	p.Save(w, w/2, fp); 
}

// reportBandsCancelled generates filled line charts covering the covered bands for distance travelled
// and % trips cancelled over time.
func reportBandsCancelled(compiled []botStatsCompiled,mp ModelParams) {

	// Set axis labels
	p, err := plot.New()
	if err != nil {
		return
	}
	p.X.Label.Text = "Day"
	p.Y.Label.Text = "Trips Cancelled (%)"

	// Add the source points for each band
	palette,err := brewer.GetPalette(brewer.TypeSequential,"Reds",len(compiled))
	if err != nil {
		return 
	}
	for i,compiledBand := range(compiled) {
		line, err := plotter.NewLine(compiledBand.cancelled)
		if err != nil {
			return
		}
		line.FillColor = palette.Colors()[i]
		p.Add(line)
	}

	// Set the axis ranges
	p.X.Max= float64(mp.DaysToRun)
	p.X.Min = 1
	p.Y.Min = 0
	p.Y.Max = 100

	// Save the plot to a PNG file.
	fp := filepath.Join(mp.WorkingFolder,"cancelledbyband.png")
	w:= vg.Length(mp.LargeChartWidth)*vg.Centimeter
	p.Save(w, w/2, fp); 
}


