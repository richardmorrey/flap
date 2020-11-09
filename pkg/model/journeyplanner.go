package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"github.com/richardmorrey/flap/pkg/db"
	"errors"
	"math/rand"
	"sync"
	"strings"
	"bytes"
	"encoding/binary"
	"fmt"
)

var  EDAYTOOFARAHEAD = errors.New("Value for days is too far ahead")
var  EZEROPLANNINGDAYS = errors.New("Zero planning days specified")

type journeyType uint8

const (
	jtOutbound journeyType  = iota
	jtInbound
	jtPromise
)

type journey struct {
	jt		journeyType
	flight		flap.Flight
	length		flap.Days
}

// From implements db/Serialize
func (self *journey) From(buff *bytes.Buffer) error {
	
	err := self.flight.From(buff)
	if err != nil {
		return err
	}

	err = binary.Read(buff,binary.LittleEndian,&self.jt)
	if err != nil {
		return logError(err)
	}

	err = binary.Read(buff,binary.LittleEndian,&self.length)
	return err
}

// To implemented as part of db/Serialize
func (self *journey) To(buff *bytes.Buffer) error {

	err := self.flight.To(buff)
	if err != nil {
		return err
	}

	err = binary.Write(buff,binary.LittleEndian,&self.jt)
	if err != nil {
		return logError(err)
	}

	err = binary.Write(buff,binary.LittleEndian,&self.length)
	return err
}

type plannerDay struct {
	journies []journey
}

// To implements db/Serialize
func (self *plannerDay) To(buff *bytes.Buffer) error {
	n := int32(len(self.journies))
	err := binary.Write(buff, binary.LittleEndian,&n)
	if err != nil {
		return logError(err)
	}
	for i:=int32(0); i < n; i++ {
		err = self.journies[i].To(buff)
		if (err !=nil) {
			return logError(err)
		}
	}
	return nil
}

// From implemments db/Serialize
func (self *plannerDay) From(buff *bytes.Buffer) error {
	var n int32
	err := binary.Read(buff,binary.LittleEndian,&n)
	if err != nil {
		return logError(err)
	}

	var entry journey
	self.journies =  make([]journey,0,5)
	for  i:=int32(0); i < n; i++ {
		err = entry.From(buff)
		if (err != nil) {
			return logError(err)
		}
		self.journies = append(self.journies,entry)
	}
	return nil
}

type journeyPlanner struct{
	table db.Table
	mux  sync.Mutex
}

var ENOJOURNEYSPLANNED = errors.New("No journeys have been planned for today")
const journeyPlannerTableName = "journeyplanner"

// NewJourneyPlanner is factory function for journeyPlanner
func NewJourneyPlanner(database db.Database) (*journeyPlanner,error) {

	// Create planner
	jp := new(journeyPlanner)

	// Create or open table
	table,err := database.OpenTable(journeyPlannerTableName)
	if  err == db.ETABLENOTFOUND { 
		table,err = database.CreateTable(journeyPlannerTableName)
	}
	if err != nil {
		return jp,err
	}
	jp.table  = table
	return jp,nil
}

// dropJourneyPlanner deletes the table holding all journey planner state
func dropJourneyPlanner(database db.Database) error {
	return database.DropTable(journeyPlannerTableName)
}

// Adds journey to the journey planner table keyed by date and passport. Thread-safe.
func (self *journeyPlanner) addJourney(pp flap.Passport, j journey) error {

	// Prevent mult-threaded write to self.table
	self.mux.Lock()
	defer self.mux.Unlock()

	// Build record key
	t := j.flight.Start.ToTime()
	recordKey := fmt.Sprintf("%s/%s",t.Format("2006-01-02"),pp.ToString())

	// Retreive any existing list for this day/traveller
	var pd plannerDay
	self.table.Get([]byte(recordKey),&pd)

	// Add journey to list
	pd.journies = append(pd.journies,j)

	// Save amended list
	err := self.table.Put([]byte(recordKey),&pd)
	if err != nil {
		return err
	}
	return nil
}

// Used to add a plan for a trip including the details of the
// traveller. Note outbound journey only is planned at this point.
// Return journey is planned only at point submission of outbound
// journey is accepted by Flight
func (self *journeyPlanner) planTrip(from flap.ICAOCode, to flap.ICAOCode, length flap.Days, pp flap.Passport, startOfDay flap.EpochTime, fe *flap.Engine) error {
	f,err := self.buildFlight(from,to,startOfDay,fe)
	if (err != nil) {
		return err
	}
	j:= journey{jt:jtOutbound,flight:*f,length:length}
	return self.addJourney(pp,j)
}

// Plans the inbound journey for given outbound journey 
func (self *journeyPlanner) planInbound(j * journey, pp flap.Passport, startOfDay flap.EpochTime,fe *flap.Engine) error {
	
	// Create return journey for last day of trip
	f,err := self.buildFlight(j.flight.ToAirport,j.flight.FromAirport,startOfDay+flap.EpochTime(j.length*flap.SecondsInDay),fe)
	if (err != nil)  {
		return err
	}
	jin := journey{jt:jtInbound,flight:*f,length:j.length}
	return self.addJourney(pp,jin)
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
func (self *journeyPlanner) buildFlight(from flap.ICAOCode,to flap.ICAOCode, startOfDay flap.EpochTime,fe *flap.Engine) (*flap.Flight, error) {

	// Retrieve airport records
	fromAirport,err := fe.Airports.GetAirport(from)
	if (err != nil) {
		return nil,logError(err)
	}
	toAirport,err := fe.Airports.GetAirport(to)
	if (err != nil) {
		return nil,logError(err)
	}

	// Calculate flight length
	_,duration,err := self.flightLength(fromAirport,toAirport)
	if (err != nil) {
		return nil,logError(err)
	}
	
	// Set start and end time, ensuring flight ends by end of first day to avoid overlap
	// with return journey
	start := startOfDay + flap.EpochTime(rand.Intn(int(flap.SecondsInDay-duration-1)))
	end := start + duration

	// Create flight
	f,e := flap.NewFlight(fromAirport,start,toAirport,end)
	return f,e
}

// Attempts to submit all flights in all journeys for the current day.
// If the journey is outbound and the submission succeeds then the inbound
// journey is planned. Not thread-safe
func (self *journeyPlanner) submitFlights(tb *TravellerBots,fe *flap.Engine, startOfDay flap.EpochTime, fp *flightPaths, debit bool) error {

	// Iterate through all journeys for today
	it,err := self.NewIterator(startOfDay)
	if err != nil {
		return logError(err)
	}
	for it.Next() {

		// Retrieve planned flights and traveller
		plannedFlights := it.Value()
		p,err := it.Passport()
		if err != nil {
			return logError(err)
		}
		
		// Submit all the flights
		for _, j := range(plannedFlights.journies)  {
		
			// Submit flight
			var flights [1]flap.Flight
			flights[0]=j.flight
			err = fe.SubmitFlights(p,flights[:],j.flight.Start,debit)
			var bi botId
			bi.fromPassport(p)

			// If successful  ...
			if err == nil {
				// ... plan journey ...
				tb.GetBot(bi).stats.Submitted(j.flight.Distance)
				if j.jt==jtOutbound {
					err = self.planInbound(&j,p,startOfDay,fe)
					if err != nil {
						return logError(err)
					}
				}
				// ... and report
				fp.addFlight(j.flight.FromAirport,j.flight.ToAirport,j.flight.Start,j.flight.End,fe.Airports,bi.band)
				logDebug("Submitted flight for",p.ToString())
			} else {
				logDebug("Flight submission refused for",p.ToString(),":",j.flight.Start.ToTime())
				tb.GetBot(bi).stats.Refused()
			}
		}

		// Delete the record
		//  TBD
	}

	return nil
}

type journeyPlannerIterator struct {
	iterator db.Iterator
}

func (self *journeyPlannerIterator) Next() (bool) {
	return self.iterator.Next() 
}

func (self *journeyPlannerIterator) Value() *plannerDay {
	var pd plannerDay
	self.iterator.Value(&pd)
	return &pd
}

func (self *journeyPlannerIterator) Passport() (flap.Passport,error) {
	var p flap.Passport
	var key = string(self.iterator.Key())
	parts := strings.Split(key,"/")
	err :=  p.FromString(parts[1])
	return p,err
}

func (self *journeyPlannerIterator) Error() error {
	return self.iterator.Error()
}

func (self *journeyPlannerIterator) Release() error {
	self.iterator.Release()
	return self.iterator.Error()
}

func (self *journeyPlanner) NewIterator(date flap.EpochTime) (*journeyPlannerIterator,error) {
	prefix := date.ToTime().Format("2006-01-02")
	iter := new(journeyPlannerIterator)
	var err error
	iter.iterator,err = self.table.NewIterator([]byte(prefix))
	return iter,err
}

