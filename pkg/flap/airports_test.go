package flap

import (
	"testing"
	"io/ioutil"
	"path/filepath"
	"os"
	"flap/db"
)

var AIRPORTSTESTFOLDER="airportstest"

func setup(t *testing.T) *db.LevelDB{
	if err := os.Mkdir(AIRPORTSTESTFOLDER, 0700); err != nil {
		t.Error("Failed to create test dir", err)
	}
	return db.NewLevelDB(AIRPORTSTESTFOLDER)
} 

func teardown(db *db.LevelDB) {
	db.Release()
	os.RemoveAll(AIRPORTSTESTFOLDER)
}

func checkairport(code ICAOCode, loc LatLon,airports *Airports, t *testing.T) {
	airport,err := airports.GetAirport(code)
	if err != nil {
		t.Error("Failed to get airport",code)
	}
	if airport.Loc.Lat != loc.Lat {
		t.Error("Wrong latitude", airport)
	}
	if airport.Loc.Lon != loc.Lon {
		t.Error("Wrong longitude",airport)
	}
}

func TestLoadAirpots(t *testing.T) {
	db:=setup(t)
	defer teardown(db)
	airports := NewAirports(db)
	var s = `1,"Goroka Airport","Goroka","Papua New Guinea","GKA","AYGA",-6.081689834590001,145.391998291,5282,10,"U","Pacific/Port_Moresby","airport","OurAirports"
2,"Madang Airport","Madang","Papua New Guinea","MAG","AYMD",-5.20707988739,145.789001465,20,10,"U","Pacific/Port_Moresby","airport","OurAirports"
3,"Mount Hagen Kagamuga Airport","Mount Hagen","Papua New Guinea","HGU","AYMH",-5.826789855957031,144.29600524902344,5388,10,"U","Pacific/Port_Moresby","airport","OurAirports"
4,"Nadzab Airport","Nadzab","Papua New Guinea","LAE","AYNZ",-6.569803,146.725977,239,10,"U","Pacific/Port_Moresby","airport","OurAirports"`
	csvpath:=filepath.Join(AIRPORTSTESTFOLDER,"airportsin.csv")
	if err := ioutil.WriteFile(csvpath, []byte(s), 0644); err != nil {
		t.Error("Failed to write csv", err)
	}
	if err := airports.LoadAirports(csvpath); err != nil {
		t.Error("LoadAirports failed",err)
	}
	checkairport(NewICAOCode("AYGA"), LatLon{-6.081689834590001, 145.391998291}, airports,t)
	checkairport(NewICAOCode("AYMD"), LatLon{-5.20707988739, 145.789001465}, airports,t)
	checkairport(NewICAOCode("AYNZ"), LatLon{-6.569803, 146.725977}, airports,t)
}

