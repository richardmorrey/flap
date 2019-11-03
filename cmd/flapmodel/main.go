package main
 
import (
	"fmt"
	"github.com/richardmorrey/flap/pkg/model"
	"flag"
	"os"
	"strconv"
)

func ShowHelp() {
	fmt.Println(

`
Usage: flapmodel [-configpath=<configpath>] <command>

where

<configfile> is path to a flapmodel yaml configuration file. Defaults to
"./config.yaml".

Available commands are as follows.

build
Builds the following in the working folder specified in <configfile>:  
airports		- leveldb table of airports and their latitude/
                          longitude. Used by the core flap package for
			  calculating flight distances
countriesairportsroutes - leveldb table holding weighted model of routes
                          categorised by source airport and country. Used
			  to give pseudo-realistic travel pattern for
			  travellers.
countryweights.json	- list of countries weighted according to the number
                          and size of airports they host. Used to
			  give pseudo-realistic country of origin spread for
			  travellers
This process relies on the following files downloaded and stored in the data
folder specified in config file:
https://raw.githubusercontent.com/jpatokal/openflights/master/data/routes.dat 
https://raw.githubusercontent.com/jpatokal/openflights/master/data/airports.dat 
http://ourairports.com/data/airports.csv

run
Runs the model for the number of days and for the number of travellers and
traveller bands specified in <configfile>. Results of the run are written out
to working folder as follows:
botspec_<x>.csv		- statistics for botspec <x> as defined in config.yaml
summary.csv		- summary statistics for each day the model was run.
Can only be executed following a succesful execution of "build". Can be
executed multiple times for each execution of "build". Model state, including
deletion of all traveller records is reset at the start of each execution.

show <botspec> <index>
Reports trip history and other details for specified bot traveller after the
last execution of run. <botspec> is a value between 0 and the number of botspec
entries in config.yaml. <botindex> 0-indexed number defining the specific bot
within that band. The number of travellers in each band is derived from
the weights across all the specs and the value of "totalTravellers".

kml <botspec> <index>
Works like show but returns kml for import into Google Earth instead of JSON.
`)
	os.Exit(0)
}

func main() {
	configfile := flag.String("configfile","./config.yaml","File path of yaml config file to use")
	flag.Parse()
	switch flag.Arg(0){
		case "build":
			engine,err := model.NewEngine(*configfile)
			if err != nil {
				fmt.Printf("\nFailed to initialize model engine with error '%s'\n",err)
			} else {
				defer engine.Release()
	 			err := engine.Build()
	 			if err != nil {
		 			fmt.Printf("\nFailed to build model with error '%s'\n",err)
				}
			}
		case "run":
			engine,err := model.NewEngine(*configfile)
			if err != nil {
				fmt.Printf("\nFailed to initialize model engine with error '%s'\n",err)
			} else {
				defer engine.Release()
				err := engine.Run()
				if err != nil {
					fmt.Printf("\nFailed to run model with error '%s'\n",err)
				}
			}
		case "show":
			spec,_ := strconv.ParseUint(flag.Arg(1), 10, 64)
			index,_ := strconv.ParseUint(flag.Arg(2), 10, 64)
		 	engine,err := model.NewEngine(*configfile)
			if err != nil {
				fmt.Printf("\nFailed to initialize model engine with error '%s'\n",err)
			} else {
				defer engine.Release()
				_,json,_,err := engine.ShowTraveller(spec,index)
				if err != nil {
					fmt.Printf("\nFailed to find traveller with error '%s'\n",err)
				} else {
					fmt.Printf("\n%s\n",json)
				}
			}
		case "kml":
			spec,_ := strconv.ParseUint(flag.Arg(1), 10, 64)
			index,_ := strconv.ParseUint(flag.Arg(2), 10, 64)
		 	engine,err := model.NewEngine(*configfile)
			if err != nil {
				fmt.Printf("\nFailed to initialize model engine with error '%s'\n",err)
			} else {
				defer engine.Release()
				_,_,kml, err := engine.ShowTraveller(spec,index)
				if err != nil {
					fmt.Printf("\nFailed to find traveller with error '%s'\n",err)
				} else {
					fmt.Printf("\n%s\n",kml)
				}
			}

		case "help":
		default:
			ShowHelp()
	}
}

