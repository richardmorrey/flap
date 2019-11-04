package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"errors"
	"fmt"
	"math/rand"
	"path/filepath"
	"os"
	"strings"
)

var ECOULDNTFINDWEIGHT= errors.New("Couldnt find country weight for bot")
type bandIndex uint8
type botIndex uint32

type travellerBot struct {
	countryStep    float64
	numInstances   botIndex
	flyProb	       Probability
	stats	       botStats
}

type botId struct {
	band  bandIndex
	index botIndex
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
	taken	 uint64
	refused  uint64
}

// Submitted updates stats to reflect fact a journey
// has been successfully submitted
func (self *botStats) Submitted(dist flap.Kilometres) {
	self.taken++
	self.distance += dist
}

// Refused updates stats to relect fact a journey 
// submission has been refused by flap
func (self  *botStats) Refused() {
	self.refused++
}

// Report writes a single CSV row with stats
func (self *travellerBot) ReportDay(rdd flap.Days) string {
	
	// Format stats
	line := fmt.Sprintf("%d,%d,%d,",
		self.stats.taken/uint64(self.numInstances),
		self.stats.refused/uint64(self.numInstances),
		self.stats.distance/flap.Kilometres(self.numInstances))

	// Reset counters
	self.stats.taken = 0
	self.stats.refused = 0
	self.stats.distance = 0

	return line
}

type TravellerBots struct {
	bots	[]travellerBot
	countryWeights *CountryWeights
	fh		*os.File
	statsFolder    string
}

func NewTravellerBots(cw *CountryWeights) *TravellerBots {
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
					line += fmt.Sprintf(",tpp_%d,rpp_%d,dpp_%d",bb,bb,bb) 
				}
				line +="\n"
				self.fh.WriteString(line)
			}
		}

		// Write line
		line := ""
		for bb := bandIndex(0); bb < bandIndex(len(self.bots)); bb++ {
			line += self.bots[bb].ReportDay(rdd)
		}
		line=strings.TrimRight(line,",")
		line+="\n"
		self.fh.WriteString(line)
	}
}

// Build constructs bot configurations for each band from provided model params
func (self *TravellerBots) Build(modelParams *ModelParams) error {

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
		return glog(err)
	}
	for _, botspec := range modelParams.BotSpecs {
		var bot travellerBot
		bot.numInstances= botIndex((float64(botspec.Weight)/float64(weightTotal))*float64(modelParams.TotalTravellers))
		if (bot.numInstances > 0) {
			bot.countryStep= float64(topWeight)/float64(bot.numInstances)
		}
		bot.flyProb = botspec.PlanProbability
		self.bots  = append(self.bots,bot)
	}
	return nil
}

// planTrips "throws dice" for every traveller bot in every band according to probability
// of travellers in the band travelling on any one day. If the dice comes up and the
// travellerbot is not in the middle of a trip, a new trip is planned using weighted 
// country-airpots-routes model
func (self *TravellerBots) planTrips(cars *CountriesAirportsRoutes, jp* journeyPlanner, travellers *flap.Travellers) error {

	// Iterate through each bot in each band
	for i:=bandIndex(0); i < bandIndex(len(self.bots)); i++ {
		for j:=botIndex(0); j < self.bots[i].numInstances; j++ {
			dice:=Probability(rand.Float64())
			if dice <= self.bots[i].flyProb {

				// Retrieve passport
				p,err := self.getPassport(botId{i,j})
				if err != nil {
					return glog(err)
				}

				// Check bot is not already travelling
				t,err := travellers.GetTraveller(p)
				if (err != nil) || !t.MidTrip() {

					// Retrieve source country record for traveller
					car,err := cars.getCountry(p.Issuer)
					if err != nil {
						return glog(err)
					}

					// Choose source airport
					ap,err := car.choose()
					if err != nil {
						return glog(err)
					}
					airport := car.Airports[ap]

					// Choose route (destination airport)
					route,err := airport.choose()
					if err != nil {
						return glog(err)
					}

					// plan trip
					err = jp.planTrip(airport.Code, airport.Routes[route].To, botId{i,j})
					if err != nil {
						return glog(err)
					}
				}
			}
		}
	}
	return nil
} 
