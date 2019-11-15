# flap
FLAP is a system/process for reducing global air travel in order to help save the planet

## doc/
Contains flapwhitepaper.pdf, a white paper that explains the idea in detail.

## pkg/flap
This a full implementation of the functionality defined in Sections 2-3 of the white paper and implemented consistent with Section 4. The three core processes defined in section 4 can be involved via two public functions 
	Engine.SubmitFlights - Flight data processing.
	Engine.UpdateAndBackfill - Trip completion enforcement and backfilling.		       		
In a full deployment the first of these would be driven by REST interfaces invoked by airline systems.

## pkg/flap/db
Ths is a sub-package containing database implementations for use by pkg/flap encapsulated behind simple key/value-store stule interfaces. As of now there is only support for a single database technology - leveldb. This is suitable for modelling purposes only.

## pkg/model
This is a package for modelling FLAP behaviour by exercising pkg/flap with realistic data for thousands/millions of travellers. It is driven by gegenuine data about world airports and the routes between them as curated by openflights.org. 

## cmd/flapmodel
This is a modelling tool for build and running different models using pkg/model and pkg/flap. It is the best starting point for anyone - please someone ;) - interested in FLAP. It has command-line help and a fully documented default configuration file: cmd/flapmodel/config.yaml.

## configs
Contained example and documented configuration files to use with flapmodel

## website/
This contains all the website content with more material including a summary of key modelling findings: http://www.flapyourarms.org .
