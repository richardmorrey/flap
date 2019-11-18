# FLying is A Privelege

## Overview
FLAP is a non-judgemental system/process for reducing global air travel in order to help save the planet. This repository includes an implementation of the core FLAP algorithm together with a modelling tool to exercise it and other information about FLAP. For a gentle introduction please go to http://www.flapyourarms.org . 

## Getting Started
The best way to get started with FLAP is to install and run the modelling tool, flapmodel, on Linux. Follow these steps.

1. Install and set up a Golang development environment (see https://golang.org/doc/install ) and go to the golang source folder you have set up:

		cd $GOPATH/src	

2. Install and build flap:

		go get "github.com/richardmorrey/flap/cmd/flapmodel"
		cd github.com/richardmorrey/flap/cmd/flapmodel/
		go build

3. Create directories and download data files:

		mkdir working
		mkdir data
		cd data
		wget https://raw.githubusercontent.com/jpatokal/openflights/master/data/airports.dat
  		wget https://raw.githubusercontent.com/jpatokal/openflights/master/data/routes.dat
  		wget https://ourairports.com/data/airports.csv
		cd ..

4. Build the model:

		./flapmodel --configfile=../../configs/default.yaml build

5. Run the model:

		./flapmodel --configfile=../../configs/default.yaml run

6. Explore the results:

		ls working/*.csv

Use "flapmodel --help" and view comments in "pkg/configs/default.yaml" to explore ways of varying the model.

Note you do not need to rebuild the model for every run.

## Getting Around

This section summarizes what you will find in each top level folder of the FLAP GIT respository.

### doc/
Contains flapwhitepaper.pdf, a white paper that explains the FLAP idea in detail.

### pkg/flap/
This a full implementation of the functionality as defined in Sections 2-3 of the white paper and implemented consistent with Section 4. The three core processes defined in Section 4 can be involved via two public functions 
	Engine.SubmitFlights - Flight data processing.
	Engine.UpdateAndBackfill - Trip completion enforcement and backfilling.		       		
In a full deployment the first of these would be driven by REST interfaces invoked by airline systems. For example usage see pkg/model/engine.go.

Note this package has good working test coverage. Use "go test" to invoke.

### pkg/flap/db/
Ths is a package containing database implementations for use by pkg/flap encapsulated behind simple key/value-store stule interfaces. As of now there is only support for a single database technology - leveldb. This is suitable for modelling purposes only.

Note this package has good working test coverage. Use "go test" to invoke.

### pkg/model/
This is a package for modelling FLAP behaviour by exercising pkg/flap with realistic data for thousands/millions of travellers. It is driven by genuine data about world airports and the routes between them as curated by openflights.org. For example usage see cmd/flapmodel/main.go 

Note this package has limited but useful and valid test coverage. Use "go test" to invoke.

### cmd/flapmodel/
This is a modelling tool for build and running different models using pkg/model and pkg/flap. It is the best starting point for anyone - please someone ;) - interested in FLAP. It has command-line help and a fully documented default configuration file: cmd/flapmodel/config.yaml.

### configs/
Contained example and documented configuration files to use with flapmodel.

### website/
This contains all the website content with more material including a summary of key modelling findings: http://www.flapyourarms.org .
