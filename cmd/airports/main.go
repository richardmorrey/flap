package main

import (
	"os"
	"fmt"
	"flag"
	"bufio"
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"io"
	"strconv"
)

type Airport struct {
	Iata string
	Lng  float64
	Lat  float64
	Tz   string
	Nm   string
	Cy   string
	Co   string
}

// LoadAirports populates a map  "airports" from a csv file
// CSV file must be fomatted as per "airpots.dat" file from https://openflights.org/data.html.
// Each entry table is keyed by ICAOCode and holds the latitude and longitude of the airport
func LoadAirports(filepath string) (map[string]Airport,error) {

	Airports := make(map[string]Airport)

	// Open and iterate through CSV file
	csvFile, err := os.Open(filepath)
	if (err != nil) {
		return nil,err
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else
		if err != nil {
			return nil,err
		}
		
		// Extract fields of interest
		var ap Airport
		ap.Nm = line[1] 
		ap.Cy = line[2] 
		ap.Co = line[3]
		ap.Iata = line[4]
		ap.Lat,err=strconv.ParseFloat(line[6],64)
		if err != nil {
			return nil,err
		}
		ap.Lng,err=strconv.ParseFloat(line[7],64)
		if err != nil {
			return nil,err
		}
		ap.Tz = line[11] 
		Airports[line[5]] = ap
	}
	return Airports,nil
}

// main is main
func main() {
	
	// Parse command-line
	flag.Parse()

	// Load airpots
	airports,err := LoadAirports(flag.Arg(0))
	if err != nil {
		fmt.Printf("Failed to load airports with error %s\n", err)
		os.Exit(0)
	}

	// Output as a single javascript global
	jsonData, _ := json.Marshal(airports)
	src := "var gAirports=" + string(jsonData) +";"
	ioutil.WriteFile("airports.js", []byte(src), 0644)
}

