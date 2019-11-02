package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"errors"
	"math/rand"
	"math"
)

var  EDAYTOOFARAHEAD = errors.New("Value for days is too far ahead")

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
}

type plannerDay []journey

type journeyPlanner struct{
	days []plannerDay
	tripLengths []flap.Days
}

var ENOJOURNEYSPLANNED = errors.New("No journeys have been planned for today")

// Returns longest and shortest trip lengths from given list
func (self* journeyPlanner) minmaxTripLength() (flap.Days,flap.Days) {
	var maxLength flap.Days
	var minLength = flap.Days(math.MaxInt64)
	for _,length := range(self.tripLengths) {
		if length > maxLength {
			maxLength=length
		}
		if (length < minLength) {
			minLength=length
		}
	}
	return minLength, maxLength
}

// NewJourneyPlanner is factory function for journeyPlanner
func NewJourneyPlanner(tripLengths []flap.Days) (*journeyPlanner,error) {
	if (tripLengths == nil || len(tripLengths)==0) {
		return nil,flap.EINVALIDARGUMENT
	}
	jp := new(journeyPlanner)
	jp.tripLengths=tripLengths
	_, maxLen  := jp.minmaxTripLength() 
	jp.days =make([]plannerDay,maxLen+2,maxLen+2)
	return jp,nil
}

// Adds journey to a day. Day is 0-indexed with 0 meaning "today"
func (self *journeyPlanner) addJourney(j journey, day flap.Days) error {
	if day >= flap.Days(len(self.days)) {
		return glog(EDAYTOOFARAHEAD)
	}
	if self.days[day] == nil {
		self.days[day] = make([]journey,0,100)
	}
	self.days[day] = append(self.days[day],j)
	return nil
}

// Used to add a plan for a trip including the details of the
// traveller. Note outbound journey only is planned at this point.
// Return journey is planned only at point submission of outbound
// journey is accepted by Flight
func (self *journeyPlanner) planTrip(from flap.ICAOCode, to flap.ICAOCode, bot botId) error {
	j:= journey{jt:jtOutbound,flight:journeyFlight{from,to},bot:bot}
	return self.addJourney(j,0)
}

// Plans the inbound journey for given outbound journey. Includes
// random selection of trip length in days. 
func (self *journeyPlanner) planInbound(j * journey,outboundEnd flap.EpochTime,startOfDay flap.EpochTime) error {
	
	// Choose a trip length
	tripLength := self.tripLengths[0]
	if len(self.tripLengths) > 1 {
		tripLength = self.tripLengths[rand.Intn(len(self.tripLengths)-1)]
	}

	// Calculate start day, ensuring the flight starts on a different day to when
	// the outbound flight ends
	inboundDay :=  flap.Days(((outboundEnd - startOfDay) / flap.SecondsInDay)) + tripLength

	// Create journey and add
	j2 := journey{jt:jtInbound,flight:journeyFlight{j.flight.to,j.flight.from},bot:j.bot}
	return self.addJourney(j2,inboundDay)
}

// Builds a flap Flight for a given journey flight and datetime, creating a
// start time randomly within the given day
const airspeed = 0.244 // kms per sec
func (self *journeyPlanner) buildFlight(jf *journeyFlight, startOfDay flap.EpochTime,fe *flap.Engine) (flap.EpochTime,flap.EpochTime,*flap.Flight,flap.Kilometres, error) {

	// Retrieve Airport locations
	fromAirport,err := fe.Airports.GetAirport(jf.from)
	if (err != nil) {
		return 0,0,nil,0,glog(err)
	}
	toAirport,err := fe.Airports.GetAirport(jf.to)
	if (err != nil) {
		return 0,0,nil,0,glog(err)
	}

	// Calc start and end times
	dist,err := fromAirport.Loc.Distance(toAirport.Loc)
	if err != nil {
		return 0,0,nil,0,glog(err)
	}
	start := startOfDay + flap.EpochTime(rand.Intn(flap.SecondsInDay-1))
	end := start + flap.EpochTime(float64(dist)/airspeed)

	// Create flight
	f,e := flap.NewFlight(fromAirport,start,toAirport,end)
	return start,end,f,dist,e
}

// Attempts to submit all flights in all journeys for the current day.
// If the journey is outbound and the submission succeeds then the inbound
// journey is planned
func (self *journeyPlanner) submitFlights(tb *TravellerBots,fe *flap.Engine, startOfDay flap.EpochTime, debit bool) error {
	
	// Check for plan for today
	if self.days[0] == nil {
		return ENOJOURNEYSPLANNED
	}
	
	// Iterate through all journeys for today
	for _, j := range(self.days[0])  {
		
		// Build flight
		start,end,flight,distance,err := self.buildFlight(&(j.flight),startOfDay,fe)
		if (err != nil) {
			return glog(err)
		}

		// Submit flight
		var flights [1]flap.Flight
		flights[0]=*flight
		p,err := tb.getPassport(j.bot)
		if err != nil {
			return glog(err)
		}
		err = fe.SubmitFlights(p,flights[:],start,debit)
		// If successful plan return journey
		if err == nil {
			tb.GetBot(j.bot).stats.Submitted(distance)
			if j.jt==jtOutbound {
				err = self.planInbound(&j,end,startOfDay)
				if err != nil {
					return glog(err)
				}
			}
			//fmt.Printf("%d_%d flying from %s to %s, distance %d\n", j.bot.band,j.bot.index,string(j.flight.from[:]),string(j.flight.to[:]),int(distance))
		} else {
			tb.GetBot(j.bot).stats.Refused()
		// If successful plan return journey
		}
	}

	// Delete the plan for today now it has been submitted
	copy(self.days[:],self.days[1:])
	self.days= self.days[:len(self.days)-1]
	self.days=append(self.days,plannerDay{})
	return nil
}


