package model

import (
	"testing"
	"os"
	"github.com/richardmorrey/flap/pkg/db"
	"github.com/richardmorrey/flap/pkg/flap"
	"reflect"
	//"fmt"
)

var COUNTRIESAIRPORTSROUTESTESTFOLDER="cartest"

func setup(t *testing.T) (*db.LevelDB, *CountriesAirportsRoutes) {
	if err := os.Mkdir(COUNTRIESAIRPORTSROUTESTESTFOLDER, 0700); err != nil {
		t.Error("Failed to create test dir", err)
	}
	db := db.NewLevelDB(COUNTRIESAIRPORTSROUTESTESTFOLDER)
	car := NewCountriesAirportsRoutes(db)
	if (car==nil) {
		t.Error("Failed to create CountriesAirportRoutes instance")
	}
	return db,car
} 

func teardown(db *db.LevelDB, car *CountriesAirportsRoutes) {
	db.Release()
	os.RemoveAll(COUNTRIESAIRPORTSROUTESTESTFOLDER)
}

func TestPutEmptyCountry(t *testing.T) {
	db,car := setup(t)
	defer teardown(db,car)
	country := newCountry()
	var code flap.IssuingCountry
	country.Airports = nil
	err := car.putCountry(countryState{countryCode:code,country:country})
	if (err != nil) {
		t.Error("Failed to put empty country",err)
	}
	gotCountry,err := car.getCountry(code)
	if (err != nil) {
		t.Error("Failed to get emtpy country",err)
	}
	if !reflect.DeepEqual(gotCountry,*country){
		t.Error("Got country doesnt equal put country",country,gotCountry)
	}
}

func TestPutOneAirportCountry(t *testing.T) {
	db,car := setup(t)
	defer teardown(db,car)
	country := newCountry()
	var code flap.IssuingCountry
	airport := country.getAirport(flap.NewICAOCode("a"))
	airport.Routes=nil // Force Routes to nil to make test pass. "Decode" for an empty slice leaves it as nil.
	country.add(10)
	err := car.putCountry(countryState{countryCode:code,country:country})
	if (err != nil) {
		t.Error("Failed to put country",err)
	}
	gotCountry,err := car.getCountry(code)
	if (err != nil) {
		t.Error("Failed to get country",err)
	}
	if !reflect.DeepEqual(gotCountry,*country) {
		t.Error("Got country doesnt equal put country",gotCountry)
	}
}

func TestPutThreeAirportCountry(t *testing.T) {
	db,car := setup(t)
	defer teardown(db,car)
	country := newCountry()
	var code flap.IssuingCountry
	airport:=country.getAirport(flap.NewICAOCode("a"))
	airport.Routes=nil
	airport=country.getAirport(flap.NewICAOCode("b"))
	airport.Routes=nil
	airport=country.getAirport(flap.NewICAOCode("c"))
	airport.Routes=nil
	err := car.putCountry(countryState{countryCode:code,country:country})
	if (err != nil) {
		t.Error("Failed to put empty country",err)
	}
	gotCountry,err := car.getCountry(code)
	if (err != nil) {
		t.Error("Failed to get emtpy country",err)
	}
	if !reflect.DeepEqual(gotCountry,*country) {
		t.Error("Got country doesnt equal put country",gotCountry)
	}
}

func TestPutAirportsRoute(t *testing.T) {
	db,car := setup(t)
	defer teardown(db,car)
	country := newCountry()
	var code flap.IssuingCountry
	airport := country.getAirport(flap.NewICAOCode("a"))
	airport.Routes=nil
	airportb := country.getAirport(flap.NewICAOCode("b"))
	airportb.Routes=append(airport.Routes,Route{From:flap.NewICAOCode("a"),To:flap.NewICAOCode("b")})
	airport = country.getAirport(flap.NewICAOCode("c"))
	airport.Routes=nil
	err := car.putCountry(countryState{countryCode:code,country:country})
	if (err != nil) {
		t.Error("Failed to put empty country",err)
	}
	gotCountry,err := car.getCountry(code)
	if (err != nil) {
		t.Error("Failed to get emtpy country",err)
	}
	if !reflect.DeepEqual(gotCountry,*country) {
		t.Error("Got country doesnt equal put country",gotCountry)
	}
	gotAirport := gotCountry.getAirport(flap.NewICAOCode("b"))
	if !reflect.DeepEqual(gotAirport,airportb) {
		t.Error("Got airport doesnt equal put airport",gotAirport)
	}
}

