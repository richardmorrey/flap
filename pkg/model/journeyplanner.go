package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"errors"
	"math/rand"
	"sync"
)

var  EDAYTOOFARAHEAD = errors.New("Value for days is too far ahead")
var  EZEROPLANNINGDAYS = errors.New("Zero planning days specified")

type journeyFlight struct {
	from flap.ICAOCode
	to flap.ICAOCode
}

type journeyType uint8

const (
	jtOutbound journeyType  = iota
	jtInbound
)

type journey struct {
	jt		journeyType
	flight		journeyFlight
	bot		botId
	length		flap.Days
}

type plannerDay []journey

type journeyPlanner struct{
	days []plannerDay
	mux  sync.Mutex
}

var ENOJOURNEYSPLANNED = errors.New("No journeys have been planned for today")

// NewJourneyPlanner is factory function for journeyPlanner
func NewJourneyPlanner(planningDays flap.Days) (*journeyPlanner,error) {
	if planningDays < 1 {
		return nil, EZEROPLANNINGDAYS
	}
	jp := new(journeyPlanner)
	jp.days =make([]plannerDay,planningDays+2,planningDays+2)
	return jp,nil
}

// Adds journey to a day. Day is 0-indexed with 0 meaning "today"
func (self *journeyPlanner) addJourney(j journey, day flap.Days) error {

	// Prevent mult-threaded write to self.days
	self.mux.Lock()
	defer self.mux.Unlock()

	// Check day isnt too far ahead
	if day >= flap.Days(len(self.days)) {
		return logError(EDAYTOOFARAHEAD)
	}

	// Make slice if this is the first journey on specified day
	if self.days[day] == nil {
		self.days[day] = make([]journey,0,100)
	}

	// Add the journey
	self.days[day] = append(self.days[day],j)
	logDebug("Added journey. Day +",day," now has ", len(self.days[day]), " journeys.")
	return nil
}

// Used to add a plan for a trip including the details of the
// traveller. Note outbound journey only is planned at this point.
// Return journey is planned only at point submission of outbound
// journey is accepted by Flight
func (self *journeyPlanner) planTrip(from flap.ICAOCode, to flap.ICAOCode, length flap.Days, bot botId, day flap.Days) error {
	j:= journey{jt:jtOutbound,flight:journeyFlight{from,to},length:length,bot:bot}
	return self.addJourney(j,day)
}

// Plans the inbound journey for given outbound journey 
func (self *journeyPlanner) planInbound(j * journey) error {
	
	// Create journey for last day of tripd
	j2 := journey{jt:jtInbound,flight:journeyFlight{j.flight.to,j.flight.from},bot:j.bot}
	return self.addJourney(j2,j.length-1)
}

// flightLength calulates distance and duration of flight between given two airports
const airspeed = 0.244 // kms per se
func (self *journeyPlanner) flightLength(from flap.Airport, to flap.Airport) (flap.Kilometres,flap.EpochTime,error) {

	dist,err := from.Loc.Distance(to.Loc)
	if err != nil {
		return 0,0,logError(err)
	}
	return dist,flap.EpochTime(float64(dist)/airspeed),nil
}

// Builds a flap Flight for a given journey flight and datetime, creating a
// start time randomly within the given day
func (self *journeyPlanner) buildFlight(jf *journeyFlight, startOfDay flap.EpochTime,fe *flap.Engine) (flap.EpochTime,flap.EpochTime,*flap.Flight,flap.Kilometres, error) {

	// Retrieve airport records
	fromAirport,err := fe.Airports.GetAirport(jf.from)
	if (err != nil) {
		return 0,0,nil,0,logError(err)
	}
	toAirport,err := fe.Airports.GetAirport(jf.to)
	if (err != nil) {
		return 0,0,nil,0,logError(err)
	}

	// Calculate flight length
	dist,duration,err := self.flightLength(fromAirport,toAirport)
	if (err != nil) {
		return 0,0,nil,0,logError(err)
	}
	
	// Set start and end time, ensuring flight ends by end of first day to avoid overlap
	// with return journey
	start := startOfDay + flap.EpochTime(rand.Intn(int(flap.SecondsInDay-duration-1)))
	end := start + duration

	// Create flight
	f,e := flap.NewFlight(fromAirport,start,toAirport,end)
	return start,end,f,dist,e
}

// Attempts to submit all flights in all journeys for the current day.
// If the journey is outbound and the submission succeeds then the inbound
// journey is planned
func (self *journeyPlanner) submitFlights(tb *TravellerBots,fe *flap.Engine, startOfDay flap.EpochTime, fp *flightPaths, debit bool) error {

	// Prevent mult-threaded write to self.days
	self.mux.Lock()
	defer self.mux.Unlock()
	logInfo("submitFlights processing ", len(self.days[0])," journeys")

	// Iterate through all journeys for today
	for _, j := range(self.days[0])  {
		
		// Build flight
		start,end,flight,distance,err := self.buildFlight(&(j.flight),startOfDay,fe)
		if (err != nil) {
			return logError(err)
		}

		// Submit flight
		var flights [1]flap.Flight
		flights[0]=*flight
		p,err := tb.getPassport(j.bot)
		if err != nil {
			return logError(err)
		}
		err = fe.SubmitFlights(p,flights[:],start,debit)
		// If successful  ...
		if err == nil {
			// ... plan journey ...
			tb.GetBot(j.bot).stats.Submitted(distance)
			if j.jt==jtOutbound {
				err = self.planInbound(&j)
				if err != nil {
					return logError(err)
				}
			}
			// ... and report
			fp.addFlight(j.flight.from,j.flight.to,start,end,fe.Airports,j.bot.band)
			logDebug("Submitted flight for",j.bot)
		} else {
			logDebug("Flight submission refused for",j.bot,":",start.ToTime())
			tb.GetBot(j.bot).stats.Refused()
		}
	}

	// Delete the plan for today now it has been submitted
	copy(self.days[:],self.days[1:])
	self.days[len(self.days)-1]=plannerDay{}
	return nil
}


