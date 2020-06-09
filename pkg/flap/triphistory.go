// Package flap provides all the functions and stucs required to run a coe Flap system as defined in TBD. This is with the exception of rhe REST interfaces. It uses interfaces to allo plugin of differnt key-value database technologies. 
package flap

import (
	"errors"
	"math"
	"sort"
	"encoding/json"
	"encoding/binary"
	"time"
	"github.com/twpayne/go-kml"
	"bytes"
	"image/color"
)

type Kilometres float64
type EpochTime uint64
type EpochDelta int64
type flightType int8
type tripHistoryIndex int8
type LatLon struct
{
	Lat float64
	Lon float64
}
func (self *LatLon) Valid() bool {
	if math.Abs(self.Lat) > 90 {
		return false
	}
	if math.Abs(self.Lon) > 180 {
		return false
	}

	return true
}

var ENOTIMPLEMENTED = errors.New("Not Implemented")
var EINVALIDARGUMENT= errors.New("Invalid Argument")
var EFLIGHTTOOOLD = errors.New("Flight too old to add")
var EINVALIDFLAPPARAMS = errors.New("Invalid FLAP parameters")
var EGROUNDED = errors.New("Traveller is grounded")
var EEMPTYTRIPHISTORY = errors.New("Empty trip history")
var ELATESTFLIGHTNOTTRIPEND = errors.New("Latest flight is not the end of a trip")
var EEPOCHNOTSTARTOFDAY = errors.New("Epoch time not the start of a UTC day")
var EFLIGHTNOTFOUND = errors.New("Flight not found")
var ENOCHANGEREQUIRED	= errors.New("No update to trip history required")

const (
	etFlight flightType  = iota
	etJourneyEnd
	etTripEnd
	etTravellerTripEnd
	etTripReopen
)

func (self *EpochTime) ToTime() time.Time {
	return time.Unix(int64(*self), 0)
}

func (self *EpochTime) toEpochDays(roundup bool) epochDays {
	if roundup {
		return epochDays((*self + (SecondsInDay-EpochTime(1))) / SecondsInDay)
	} else {
		return epochDays(*self / SecondsInDay)
	}
}

func min(a EpochTime, b EpochTime) EpochTime {
	if a > b {
		return b
	}
	return a
}

const MaxEpochTime=EpochTime(math.MaxUint64)

type Flight struct {
	et flightType
	start EpochTime
	end  EpochTime
	from  ICAOCode
	to ICAOCode
	distance  Kilometres
}

func (self *Flight) Older(the *Flight) bool {
	return bool(the.start >= self.start)
}

func (self *Flight) setType(ft flightType, respectful bool) {
	if respectful == true &&
	(self.et == etTravellerTripEnd  || self.et == etTripReopen) {
		return
	}
	self.et = ft
	return
}

// To implements db/Serialize
func (self *Flight) To(buff *bytes.Buffer) error {
	err:= binary.Write(buff,binary.LittleEndian,&self.et)
	if err != nil {
		return logError(err)
	}
	err = binary.Write(buff,binary.LittleEndian,&self.start)
	if err != nil {
		return logError(err)
	}
	err = binary.Write(buff,binary.LittleEndian,&self.end)
	if err != nil {
		return logError(err)
	}
	err = binary.Write(buff,binary.LittleEndian,&self.from)
	if err != nil {
		return logError(err)
	}
	err = binary.Write(buff,binary.LittleEndian,&self.to)
	if err != nil {
		return logError(err)
	}
	return binary.Write(buff,binary.LittleEndian,&self.distance)
}

// From implemments db/Serialize
func (self *Flight) From(buff *bytes.Buffer) error {
	err:= binary.Read(buff,binary.LittleEndian,&self.et)
	if err != nil {
		return logError(err)
	}
	err = binary.Read(buff,binary.LittleEndian,&self.start)
	if err != nil {
		return logError(err)
	}
	err = binary.Read(buff,binary.LittleEndian,&self.end)
	if err != nil {
		return logError(err)
	}
	err = binary.Read(buff,binary.LittleEndian,&self.from)
	if err != nil {
		return logError(err)
	}
	err = binary.Read(buff,binary.LittleEndian,&self.to)
	if err != nil {
		return logError(err)
	}
	return binary.Read(buff,binary.LittleEndian,&self.distance)
}

// NewFlight constructs a new flight from one airpot to another and
// with the given start and end date/times. It also calculates and
// stores distance in kilometres of the flight based on coordinates
// of the two airports. Thus distance field is correctly set for
// all Flight instances
func NewFlight(from Airport,start EpochTime,to Airport,end EpochTime) (*Flight,error) {
	if start <= 0 {
		return nil, EINVALIDARGUMENT
	}
	if end <= start {
		return nil,EINVALIDARGUMENT
	}
	flight := new(Flight)
	flight.start=start
	flight.end=end
	flight.from=from.Code
	flight.to=to.Code

	var err error
	flight.distance,err = from.Loc.Distance(to.Loc)
	return flight,err
}

func haversine(angle float64) float64 {
	return .5 * (1 - math.Cos(angle))
}

func degPos(ll LatLon) pos {
	return pos{ll.Lat * math.Pi / 180, ll.Lon * math.Pi / 180}
}

const rEarth = 6372.8 // km

type pos struct {

	φ float64 // latitude, radians
	ψ float64 // longitude, radians
}

// Distance calcuates the approximate distance between two sets of latitude/longitude coordinates.
// It uses the harvesine formula and returns a result in whole kilometres.
// Implementation based on https://rosettacode.org/wiki/Haversine_formula#Go
func (self* LatLon) Distance(ll2 LatLon) (Kilometres,error) {
	if !self.Valid() || !ll2.Valid() {
		return 0,EINVALIDARGUMENT
	}
	p1 := degPos(*self)
	p2 := degPos(ll2)
	distance := 2 * rEarth * math.Asin(math.Sqrt(haversine(p2.φ-p1.φ)+
	        math.Cos(p1.φ)*math.Cos(p2.φ)*haversine(p2.ψ-p1.ψ)))
	return Kilometres(distance),nil
}

const MaxFlights=100
const SecondsInDay=86400

// TripHistory manages a record of all the Flights, Journeys and Trips made by a Traveller. 
// It is a list of events ordered by the start epoch time of each event, with the most recent event first.
// Each Events is a Flight but some have special status as a JourneyEnd or a TripEnd. 
// The start of a Journey is implicitly the first Flight after a JourneyEnd or the first Flight in the history.
// The start of a Trip is implicity the first Flight after a TripEnd or the first Flight in the history.
// The history stores a maximum of 100 events, dropping the oldest event if necessary to maintain this.
type TripHistory struct {
	entries			[MaxFlights]Flight
	oldestChange		tripHistoryIndex
}

// AddFlight inserts a Flight into the trip history, in the correct place to maintain ordering
func (self *TripHistory) AddFlight(f *Flight ) error {	
	
	// Find index to add flight
	i := sort.Search(MaxFlights, func(i int) bool { return self.entries[i].Older(f)})
	if  i >= MaxFlights {
		return EFLIGHTTOOOLD
	}
	
	// Copy older entries down one - the oldest is dropped if history is full -
	// and insert
	copy(self.entries[i+1:], self.entries[i:])
	self.entries[i] = *f

	// Set oldestAdded if necessary to speed up updating
	if i > int(self.oldestChange)  {
		self.oldestChange = tripHistoryIndex(i)
	} else {
		if self.oldestChange < MaxFlights-1 {
			self.oldestChange++
		}
	}
	return  nil
}

// RemoveFlight removed a Flight from the trip history. Only
// removes if all field values match the given flight
func (self *TripHistory) RemoveFlight(f *Flight) error {
	
	// Find index to look for flight to remove flight
	i := sort.Search(MaxFlights, func(i int) bool { return self.entries[i].Older(f)})
	if  i >= MaxFlights {
		return EFLIGHTNOTFOUND
	}
	
	// Move forward flight by flight until we find a match
	for ; self.entries[i] != *f && self.entries[i].start==f.start; i++ {}
	if self.entries[i] != *f {
		return EFLIGHTNOTFOUND
	}

	// Copy older entries up one, thus overwriting the fight to be removed
	copy(self.entries[i:], self.entries[i+1:])

	// Set oldestAdded if necessary to speed up updating
	if i >= int(self.oldestChange)  {
		self.oldestChange = tripHistoryIndex(i)
	} else {
		self.oldestChange--
	}
	return  nil
}

// startOfTrip returns the index for the start of the trip of which flight at given index
// is a part or 0 if start of trip has been lost. Assumes the trip history is not empty 
func (self *TripHistory) startOfTrip(j tripHistoryIndex) (tripHistoryIndex,error) {

	// Check for valid argument
	if j >= MaxFlights {
		return 0,EINVALIDARGUMENT
	}

	// Check for emptry trip history
	if self.empty() {
		return 0,EEMPTYTRIPHISTORY
	}

	// Find the previous trip end flight and return the index of flight
	// immediately after
	var i tripHistoryIndex
	for i=j; i < MaxFlights && self.entries[i].start !=0; i++ {
		f := &(self.entries[i])
		if f.et == etTripEnd || f.et == etTravellerTripEnd {
			return i-1,nil
		}
	}

	// Return the oldest entry if nothing found
	return i-1,nil
}

// tripStartEndLength returns the start time of the first flight in the
// latest trip,the end time of the last flight taken and the total distance
// so far flown across the whole trip
func (self *TripHistory) tripStartEndLength() (EpochTime,EpochTime,Kilometres) {
	var d Kilometres
	var st EpochTime
	for i:=0; i < MaxFlights && self.entries[i].start !=0 && self.entries[i].et != etTripEnd && self.entries[i].et != etTravellerTripEnd; i++ {
		d += self.entries[i].distance
		st = self.entries[i].start
	}
	if d > 0 {
		return st,self.entries[0].end,d
	} else {
		return 0,0,0
	}
}

// empty returns true if their are no flights in the trip history
func (self *TripHistory) empty() bool {
	return self.entries[0].start == 0
}

// daysBetween returns whole days from time1 to time2
// if time2 is less than time1 it returns 0
func daysBetween(time1 EpochTime, time2 EpochTime) Days {
	if (time2 < time1) {
		return 0
	}
	return Days((time2-time1) / SecondsInDay)
}

// seconds converts a Days value into seconds
func  seconds(days Days) int64 {
	return int64(days*SecondsInDay)
}

type ICAOCodeSet map[ICAOCode] bool
type tripState struct {
	journeys int
	reopened bool
	flights uint64
	start EpochTime
	visited ICAOCodeSet
}

func (self *tripState) endTrip(f *Flight,respectReopen bool) {
	if !(respectReopen && self.reopened) {
		*self = tripState{}
		if f.et != etTravellerTripEnd {
			f.et=etTripEnd
		}
	}
}

func (self *tripState) endJourney(f *Flight) {
	if !(f.et==etTripReopen) && !(f.et==etTravellerTripEnd) {
		self.journeys++
		f.et=etJourneyEnd
		self.visited=nil 
	}
}

func (self *tripState) updateTrip(f* Flight, now EpochTime, params *FlapParams) {

	if self.start == 0 {
		self.start = f.start
	}

	if self.journeys == 2 && params.Promises.Algo == paNone  {
		self.endTrip(f,true)
	}

	if daysBetween(self.start,now) > params.TripLength {
		self.endTrip(f,false)
	}

	if self.flights >= params.FlightsInTrip  {
		self.endTrip(f,false)
	}
}

func (self *tripState) updateJourney(f* Flight, now EpochTime, params *FlapParams,lastFlight bool) {

	if !lastFlight && self.visited==nil {
		self.visited = make(ICAOCodeSet)
	}

	if (self.visited != nil) {
		if self.visited[f.to] {
			self.endJourney(f)
		} else {
			self.visited[f.from] = true
			self.visited[f.to]=true
		}
	}

	if (daysBetween(f.end,now) >= params.FlightInterval) {
		self.endJourney(f)
	}
}

func (self* tripState) nextEntry(th *TripHistory,i tripHistoryIndex) *Flight {
	self.flights++
	entry:=&(th.entries[i])
	if entry.et == etTripReopen {
		self.reopened = true
	}
	return entry
}

func (self *Flight) yesterday(now EpochTime) Kilometres {
	if (self.start < now) && (now-self.start <= SecondsInDay) {
		return self.distance
	}
	return 0
}

// Update updates the trip history so that JourneyEnd and TripEnd flights are correct according to the provide values
// for maxFlightDuratioan, maxTripDuration and maxFlightsinTrip.
// It automatically sets a flight to TripEnd at the point where two JourneyEnds have been added since the start of a Trip.
// It is written to be optimised for the common case where no flights have been added since the last  update.
// It ensures history is correct from start of trip before current/last onwards, to cope with removed flights.
func (self *TripHistory) Update(params *FlapParams,now EpochTime) (Kilometres,error) {
	
	var distanceYesterday  Kilometres

	// Check for empty trip history
	if (self.empty()) {
		return 0,EEMPTYTRIPHISTORY
	}

	// Check now is start of day
	if now % SecondsInDay !=  0 {
		return 0,EEPOCHNOTSTARTOFDAY
	}

	// Choose where in trip history to start from.
	// Note this paragraph is just an optimisation.
	// It is functionally valid to reevaluate the
	// whole history every time ...
	// ... if there are no changes since last update
	// and we are at trip end there is nothing to do... 
	if self.oldestChange ==0 && (self.entries[0].et== etTripEnd || self.entries[0].et== etTravellerTripEnd) {
		return 0, ENOCHANGEREQUIRED
	}
	// ... always do entireity of latest trip otherwise ...
	j,err := self.startOfTrip(self.oldestChange)
	if err != nil {
		return 0,err
	}
	// ... but if we have to do older flights go back to start of trip before.
	if self.oldestChange>0 {
		if j != MaxFlights-1 && self.entries[j+1].start != 0 {
	      		j,_ = self.startOfTrip(tripHistoryIndex(j+1))
		}
	}
	
	//Iterate through all flights forwards to the latest
	var state tripState
	for i := j; i >0 ; i-- {

		// Update boilerplate stuff
		entry:= state.nextEntry(self,i)
		nowthen := self.entries[i-1].start

		// If the traveller has called the trip closed at this
		// flight, enforce it
		if entry.et == etTravellerTripEnd {
			state.endTrip(entry,false)
			continue
		}

		// Update journey and then trip
		state.updateJourney(entry,nowthen,params,false)
		state.updateTrip(entry,nowthen,params)

		// Add up distance travelled yesterday for statistics and reporting
		distanceYesterday+=entry.yesterday(now)
	}
	
	// Update latest flight using passed current day
	entry := state.nextEntry(self,0)
	state.updateJourney(entry,now,params,true)
	state.updateTrip(entry,now,params)
	distanceYesterday+=entry.yesterday(now)

	// Reset oldest added
	self.oldestChange = 0
	return distanceYesterday,nil
}

// EndTrip changes the type of the latest flight to traveller to a traveller trip end..
// Flights of this type cannot be amended amended by Update.
func (self *TripHistory) EndTrip() error {
	
	// Check for empty trip history
	if (self.empty()) {
		return EEMPTYTRIPHISTORY
	}
	
	self.entries[0].et = etTravellerTripEnd 
	return nil
}

// ReopenTrip changes the type of the latest flight to traveller to a traveller trip end..
// Flights of this type cannot be amended amended by Update.
func (self *TripHistory) ReopenTrip() error {
	
	// Check for empty trip history
	if (self.empty()) {
		return EEMPTYTRIPHISTORY
	}
	
	// Confirm that latest flight is a tripend
	if !(self.entries[0].et==etTripEnd  || self.entries[0].et == etTravellerTripEnd) {
		return ELATESTFLIGHTNOTTRIPEND
	}
	self.entries[0].et = etTripReopen 
	return nil
}

type jsonFlight struct {
	Start time.Time
	End time.Time
	From string
	To string
	Distance Kilometres
}
type jsonJourney struct {
	Flights []jsonFlight
}
type jsonTrip struct {
	TripStatus string
	Journeys []jsonJourney
}

// AsKJSON renders current trip history as readable JSON
func (self *TripHistory) AsJSON() string {
	trips := make([]jsonTrip,0)
	var currentTrip *jsonTrip
	var currentJourney *jsonJourney
	tripStatus := "Open"
	for i:=0; self.entries[i].start !=0 && i < MaxFlights; i++ {
		if (self.entries[i].et == etTripEnd) {
			tripStatus="Closed by FLAP"
		}
		if (self.entries[i].et == etTravellerTripEnd) {
			tripStatus="Closed by Traveller"
		}
		if currentTrip == nil  || tripStatus != "Open" {
			trips = append(trips,jsonTrip{})
			currentTrip=&trips[len(trips)-1]
			currentTrip.Journeys = make([]jsonJourney,0)
			currentTrip.TripStatus = tripStatus
			tripStatus="Open"
			currentJourney=nil
		}
		if (currentJourney == nil || self.entries[i].et==etJourneyEnd) {
			currentTrip.Journeys = append(currentTrip.Journeys, jsonJourney{})
			currentJourney=&currentTrip.Journeys[len(currentTrip.Journeys)-1]
			currentJourney.Flights = make([]jsonFlight,0)
		}
		f := &(self.entries[i])
		currentJourney.Flights = 
			append(currentJourney.Flights,
			jsonFlight{f.start.ToTime(),f.end.ToTime(),f.from.ToString(),f.to.ToString(),f.distance})
	}

	jsonData, _ := json.MarshalIndent(trips, "", "    ")
	return string(jsonData)
}

// AsKML renders the current trip history as KML file for import into Google Earth 
func (self* TripHistory) AsKML(airports *Airports) string {

	d := kml.Document(kml.Name("Flight Paths"),
        			kml.SharedStyle("fps",
	                	kml.LineStyle(
			                kml.Color(color.RGBA{R: 255, G: 0, B: 0, A: 127}),
				        kml.Width(10),
					),
				))

	for i:=0; self.entries[i].start !=0 && i < MaxFlights; i++ {
		
		// Get airport locations
		from,_ := airports.GetAirport(self.entries[i].from)
		to,_ := airports.GetAirport(self.entries[i].to)

		// Build tracker:
		gt := kml.GxTrack()
		gt.Add(kml.When(self.entries[i].start.ToTime()))
		gt.Add(kml.When(self.entries[i].end.ToTime()))
		gt.Add(kml.GxCoord(kml.Coordinate{Lon: from.Loc.Lon, Lat: from.Loc.Lat, Alt: 30000}))
		gt.Add(kml.GxCoord(kml.Coordinate{Lon: to.Loc.Lon, Lat: to.Loc.Lat, Alt: 30000}))
		
		// Create placemark
		d.Add(kml.Placemark(
			       kml.Name(from.Code.ToString()+"-"+to.Code.ToString()),
			       kml.StyleURL("#fps"),
		       	       gt))
		
	}
	var buf bytes.Buffer
	kml.GxKML(d).WriteIndent(&buf, "", "  ")
	return buf.String()
}

// MidTrip returns true if the trip history indicates the Traveller is in a middle of a trip - i.e.
//the latest flight is not a trip end event - or the traveller has never travelled.
func (self *TripHistory) MidTrip() bool {
	if self.entries[0].start == 0 {
		return true
	}
	if self.entries[0].et != etTripEnd && self.entries[0].et != etTravellerTripEnd {
		return true
	}
	return false
}

// To implements db/Serialize
func (self *TripHistory) To(buff *bytes.Buffer) error {
	n := int32(sort.Search(MaxFlights,  func(i int) bool {return self.entries[i].start==0}))
	err := binary.Write(buff, binary.LittleEndian,&n)
	if err != nil {
		return logError(err)
	}
	logDebug("processing ",n," trips")
	for i:=int32(0); i < n; i++ {
		err = self.entries[i].To(buff)
		if (err !=nil) {
			return logError(err)
		}
	}

	return binary.Write(buff,binary.LittleEndian,&(self.oldestChange))
}

// From implemments db/Serialize
func (self *TripHistory) From(buff *bytes.Buffer) error {
	var n int32
	err := binary.Read(buff,binary.LittleEndian,&n)
	if err != nil {
		return logError(err)
	}
	for  i:=int32(0); i < n; i++ {
		err =  self.entries[i].From(buff)
		if err != nil {
			return logError(err)
		}
	}
	return binary.Read(buff,binary.LittleEndian,&(self.oldestChange))
}
