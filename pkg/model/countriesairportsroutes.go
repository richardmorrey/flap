package model 

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"github.com/richardmorrey/flap/pkg/db"
	"path/filepath"
	"encoding/csv"
	"encoding/gob"
	"bytes"
	"bufio"
	"io"
	"strconv"
	"os"
	"fmt"
	"sort"
	"errors"
)

var ETOOMANYAIRPORTS = errors.New("Too many airports")
var EAIRPORTDETAILSNOTFOUND = errors.New("Airport details not found")

type ICAOCarrier [2]byte

type Route struct {
	From		flap.ICAOCode
	FromCountry	flap.IssuingCountry
	To 		flap.ICAOCode
	ToCountry	flap.IssuingCountry // Needed for planning multi-flight trips
	Carrier 	ICAOCarrier
}

type RouteWithWeight struct {
	Route
	Weight weight
}

func (self* Country) To(b *bytes.Buffer) error {
	enc := gob.NewEncoder(b) 
	return enc.Encode(self)
}

func (self* Country) From(b *bytes.Buffer) error {
	dec := gob.NewDecoder(b)
	return dec.Decode(self)
}

type Airport struct {
	Weights
	Code 	    flap.ICAOCode
	Routes 	    []Route
}

type Country struct {
	Weights
	Airports    []*Airport
}

func newCountry() *Country {
	country := new(Country)
	country.Airports= make([]*Airport,0,10000)
	return country
}

type CountriesAirportsRoutes struct {
	table 		db.Table
	acs	 	airportCodesMap
	wcs 		airportWeightCountryMap
	res		[]RouteWithWeight
} 


// NewCountriesAirportsRoutes creates a new instance of CountryAirportsRoutes
// ensuring a table is created for its contents in the provided database
func NewCountriesAirportsRoutes(database db.Database) *CountriesAirportsRoutes {
	car := new(CountriesAirportsRoutes)
	table,err := database.OpenTable("countriesairportsroutes")
	if err != nil {
		table,err = database.CreateTable("countriesairportsroutes")
		if err != nil {
			return nil
		}
	}
	car.table = table
	return car
}

type countryState struct  {
	countryCode flap.IssuingCountry
	country 	*Country
	airport		*Airport
	routesCount	int	
	airportsCount	int
}

func (self* countryState) report() {
	fmt.Printf("\r...country:%s\tairports:%d\troutes:%d...         ", string(self.countryCode[:]),self.airportsCount,self.routesCount)
}

// Build builds a database to allow the running model to select routes for flights in a "realistic" pattern.
// It creates a table of entries keyed by ISO 3166-1 alpha-2 code for country. Each country contains an
// ordered array of airports by ICAO airport code. Each airport contains an array of routes to other
// airports that operate from it. Routes for each airport are assigned with ascending weights based on the
// size of the destination airport. Airpots are assigned ascending weights based on their size. Countries
// are assigned ascending weights based on the total weight of all airports it contains.
// Driven by real world data from openflights.org
func (self *CountriesAirportsRoutes) Build(dataFolder string, cw *countryWeights) (error) {
	
	// Load openflightsid-to-airport ICAOCode map
	err := self.loadIDs(dataFolder)
	if (err != nil) {
		return logError(err)
	}
	fmt.Printf("...loaded %d airport codes...\n",len(self.acs))

	// Load airport sizes
	err = self.loadSizes(dataFolder)
	if (err != nil) {
		return logError(err)
	}
	fmt.Printf("...loaded %d airport sizes...\n",len(self.wcs))

	// Load and build slice of routes ordered by country and airport
	self.res = make([]RouteWithWeight,0,60000) 
	err = self.loadRoutes(dataFolder)
	if (err != nil) {
		return logError(err)
	}
	fmt.Printf("...loaded %d routes...\n", len(self.res))

	// Create first country
	cs := countryState{countryCode:self.res[0].FromCountry ,country:newCountry()}
	cs.airport = cs.country.getAirport(self.res[0].From)
	
	// Build table entries
	for _, route := range self.res {

		// Create next country if country code has changed
		if (route.FromCountry != cs.countryCode) {
			
			// Add last airport weight and save current country
			err = self.putCountry(cs)
			if (err != nil) {
				return logError(err)
			}

			// Update country weights
			err = cw.update(&cs)
			if err != nil {
				return logError(err)
			}

			// Create next country
			cs = countryState{countryCode:route.FromCountry,country:newCountry()}
			cs.airport = cs.country.getAirport(route.From)
		}


		// Create the next airport if from code has changed
		if  route.From != cs.airport.Code {
			w,err := cs.airport.topWeight()
			if err != nil {
				return logError(err)
			}
			cs.country.add(w)
			cs.airport = cs.country.getAirport(route.From)
			cs.airportsCount++
		}

		// Add route, ensuring we dont include circular ones
		if route.To != cs.airport.Code {
			cs.airport.Routes = append(cs.airport.Routes,route.Route)
			cs.airport.add(route.Weight)
			cs.routesCount++
		} else {
			fmt.Printf("\nSkipping circular route %s to %s\n", route.To.ToString(),cs.airport.Code.ToString())
		}
	}

	// Save last country
	err = self.putCountry(cs)
	if (err != nil) {
		return logError(err)
	}
	
	// Update country weights with last country
	err = cw.update(&cs)
	if (err != nil) {
		return logError(err)
	}

	fmt.Printf("\r...built weighted model for %d countries...                           \n",len(cw.Countries))
	return nil
}

// chooseTrip chooses a trtip start and end airport for given traveller
func (self *CountriesAirportsRoutes) chooseTrip(p flap.Passport) (flap.ICAOCode,flap.ICAOCode,error) {

	// Retrieve source country record for traveller
	var empty flap.ICAOCode
	car,err := self.getCountry(p.Issuer)
	if err != nil {
		return empty,empty,logError(err)
	}

	// Choose source airport
	ap,err := car.choose()
	if err != nil {
		return empty,empty,logError(err)
	}
	airport := car.Airports[ap]

	// Choose route (destination airport)
	route,err := airport.choose()
	if err != nil {
		return empty,empty,logError(err)
	}

	return airport.Code,airport.Routes[route].To,nil
}

// getAirport returns reference to airport with given code.If airport
// doesnt exist it is created.
func (self *Country) getAirport(code flap.ICAOCode) *Airport {
	
	// Seach for airport and return if found ...
	i := sort.Search(len(self.Airports), func(i int) bool { return string(self.Airports[i].Code[:]) >= string(code[:]) })
	if i < len(self.Airports) && self.Airports[i].Code == code {
		return self.Airports[i]
	} else {
		// ... otherwise create a new airport - assuming being called in airport order
		// so no need for ordered insert
		airport := new(Airport)
		airport.Code = code
		self.Airports= append(self.Airports,airport)
		airport.Routes = make([]Route,0,100)
		return airport
	}
}

// getAirportDetails returns weight and and country code for airport with given openflights id
func (self* CountriesAirportsRoutes) getAirportDetails(id uint64) (weightCountry,flap.ICAOCode,error) {
	code := self.acs[id]
	var emptycode flap.ICAOCode
	var emptywc weightCountry
	if code == emptycode {
		return emptywc,emptycode,EAIRPORTDETAILSNOTFOUND
	}
	wc := self.wcs[code]
	if wc == emptywc {
		return emptywc,emptycode,EAIRPORTDETAILSNOTFOUND
	}
	return wc,code,nil
}

// getCounty finds and returns a record matching the given country code
func (self *CountriesAirportsRoutes) getCountry(countryCode flap.IssuingCountry) (Country,error) {
	
	if self.table == nil {
		return Country{},flap.ETABLENOTOPEN
	}

	var country Country
	err := self.table.Get(countryCode[:],&country)
	return country,err
}

// putCountry stores a record for the given country code in the
// current table. Any existing record is overwritten.
func (self  *CountriesAirportsRoutes) putCountry(cs countryState) error {
		
	// Add last airport weight
	if cs.airport != nil {
		w,err := cs.airport.topWeight()
		if err != nil {
			return logError(err)
		}	
		cs.country.add(w)
	}
	
	// Check for a table
	if self.table == nil {
		return flap.ETABLENOTOPEN
	}

	// Put record
	err := self.table.Put(cs.countryCode[:], cs.country)
	if err != nil {
		return logError(err)
	}
	cs.report()
	return nil
}

type weightCountry struct {
	weight 	weight
	country flap.IssuingCountry 
}
type airportCodesMap map[uint64]flap.ICAOCode
type airportWeightCountryMap map[flap.ICAOCode]weightCountry


func routeBefore(r1 RouteWithWeight,r2 RouteWithWeight) bool {
	if r1.FromCountry==r2.FromCountry {
		return string(r1.From[:]) < string(r2.From[:])
	}
	return string(r1.FromCountry[:]) < string(r2.FromCountry[:])
}

// loadRoutes loads routes into an ordered array from
// https://raw.githubusercontent.com/jpatokal/openflights/master/data/routes.dat 
func (self *CountriesAirportsRoutes) loadRoutes(folderPath string) error {
	
	filepath := filepath.Join(folderPath,"routes.dat")
	csvFile, err := os.Open(filepath)
	if (err != nil) {
		return logError(err)
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	for {
		// Parse one line of CSV into slice of strings
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else
		if err != nil {
			return logError(err)
		}
		
		// Retrieve openflight from and to ids
		fromId,err:=strconv.ParseUint(line[3],10,0)
		if err != nil {
			continue
		}
		toId,err:=strconv.ParseUint(line[5],10,0)
		if err != nil {
			continue
		}

		// Build route
		var route RouteWithWeight
		var wc weightCountry
		wc,route.From,err = self.getAirportDetails(fromId)
		if  err != nil {
			continue
		}
		route.FromCountry = wc.country
		wc,route.To,err = self.getAirportDetails(toId)
		if  err != nil {
			continue
		}
		copy(route.Carrier[:],line[0])
		route.Weight =  wc.weight
		route.ToCountry = wc.country
	
		// Inserted sort into slice
		i := sort.Search(len(self.res), func(i int) bool {return routeBefore(route,self.res[i])})
		self.res = append(self.res, RouteWithWeight{})
		copy(self.res[i+1:], self.res[i:])
		self.res[i] = route
	}
	return nil
}

// loadIDs loads airport ICAOCodes into a map keyed by openflights airport id from
// https://raw.githubusercontent.com/jpatokal/openflights/master/data/airports.dat 
func (self *CountriesAirportsRoutes) loadIDs(folderPath string) error {
	
	filepath := filepath.Join(folderPath,"airports.dat")
	csvFile, err := os.Open(filepath)
	if (err != nil) {
		return logError(err)
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	self.acs = make(airportCodesMap)
	for {
		// Parse one line of CSV into slice of strings
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else
		if err != nil {
			return logError(err)
		}
		
		// Add record to map
		var code flap.ICAOCode
		copy(code[:],line[5])
		id,err := strconv.ParseUint(line[0],10,0)
		if err != nil {
			continue
		}
		self.acs[id]=code
	}
	return nil
}

// loadSizes calculates airport weights based on size as contained in following, keyed by ICAOCode:
// http://ourairports.com/data/airports.csv
func (self *CountriesAirportsRoutes) loadSizes(folderPath string) error {
	
	filepath := filepath.Join(folderPath,"airports.csv")
	csvFile, err := os.Open(filepath)
	if (err != nil) {
		return logError(err)
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	self.wcs = make(airportWeightCountryMap)
	for {
		// Parse one line of CSV into slice of strings
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else
		if err != nil {
			return logError(err)
		}
		
		// Add record to map if the row contains an airpot size category
		var code flap.ICAOCode
		var sc weightCountry
		copy(code[:],line[1])
		copy(sc.country[:],line[8])
		switch (line[2]) {
			case "medium_airport":
				sc.weight=10
			case "large_airport":
				sc.weight=100
			case "small_airport":
				sc.weight=1
		}
		if (sc.weight > 0) {
			self.wcs[code]=sc
		}
	}
	return nil
}

