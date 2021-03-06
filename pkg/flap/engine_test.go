package flap

import (
	"testing"
	"os"
	"github.com/richardmorrey/flap/pkg/db"
	"reflect"
	//"fmt"
)

var ENGINETESTFOLDER="travellerstest"

func enginesetup(t *testing.T) *db.LevelDB{
	if err := os.Mkdir(ENGINETESTFOLDER, 0700); err != nil {
		t.Error("Failed to create test dir", err)
	}
	return db.NewLevelDB(ENGINETESTFOLDER)
} 

func engineteardown(db *db.LevelDB) {
	db.Release()
	os.RemoveAll(ENGINETESTFOLDER)
}

func TestNewEngine(t *testing.T) {
	db:=enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	if engine  == nil {
		t.Error("Failed to create engine")
	}
	table,err:= db.OpenTable("administrator")
	if err != nil || table == nil {
		t.Error("Administrator table not created")
	}
}

func TestEmptyParams(t  *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	paramsIn := FlapParams{}
	err := engine.Administrator.SetParams(paramsIn)
	if err != nil {
		t.Error("GetParams failed",err)
	}
	paramsOut := engine.Administrator.GetParams()
	if !(reflect.DeepEqual(paramsIn,paramsOut)) {
		t.Error("Got traveller doesnt equal put traveller", paramsOut)
	}
}

func TestValidParams(t  *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	paramsIn := FlapParams{TripLength:365,FlightsInTrip:2,FlightInterval:50,DailyTotal:1000}
	err := engine.Administrator.SetParams(paramsIn)
	if err != nil {
		t.Error("GetParams failed",err)
	}
	paramsOut := engine.Administrator.GetParams()
	if !(reflect.DeepEqual(paramsIn,paramsOut)) {
		t.Error("Got traveller doesnt equal put traveller", paramsOut)
	}
}

func TestInvalidFlightInterval(t  *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	paramsIn := FlapParams{TripLength:200,FlightsInTrip:50,FlightInterval:101,DailyTotal:1000}
	err := engine.Administrator.SetParams(paramsIn)
	if err == nil {
		t.Error("Invalid flight interval accepted",paramsIn)
	}
}

func TestInvalidFlightInTrip(t  *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	paramsIn := FlapParams{TripLength:365,FlightsInTrip:51,FlightInterval:2,DailyTotal:1000}
	err := engine.Administrator.SetParams(paramsIn)
	if err == nil {
		t.Error("Invalid flights in trip  accepted",paramsIn)
	}
}

func TestEngineSubmitFlightsEmpty(t *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	var flights []Flight
	passport := NewPassport("987654321","uk")
	err := engine.SubmitFlights(passport,flights,0,true)
	if err == nil {
		t.Error("Allowed to add empty list of flights")
	}
}

func TestEngineSubmitFlightsTaxiOverhead(t *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	engine.Administrator.SetParams(FlapParams{TaxiOverhead:100})
	var flights []Flight
	flights = append(flights,*createFlight(1,1,2))
	flights[0].Distance = 10
	passport := NewPassport("987654321","uk")
	err := engine.SubmitFlights(passport,flights,SecondsInDay,true)
	if err != nil {
		t.Error("SubmitFlights failed")
	}
	traveller,err := engine.Travellers.GetTraveller(passport) 
	if err != nil {
		t.Error("SubmitFlights failed to create traveller")
	}
	if traveller.Balance != -110 {
		t.Error("SubmitFlights didnt result in correct balance for traveller",traveller.Balance)
	}
}

func TestEngineSubmitFlights(t *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	var flights []Flight
	flights = append(flights,*createFlight(1,1,2),*createFlight(2,2,3),*createFlight(3,3,4))
	passport := NewPassport("987654321","uk")
	err := engine.SubmitFlights(passport,flights,SecondsInDay,true)
	if err != nil {
		t.Error("SubmitFlights failed")
	}
	traveller,err := engine.Travellers.GetTraveller(passport) 
	if err != nil {
		t.Error("SubmitFlights failed to create traveller")
	}
	if traveller.tripHistory.entries[0] != flights[2] {
		t.Error("SubmitFlights failed to submit flights to traveller")
	}
	if traveller.tripHistory.entries[1] != flights[1] {
		t.Error("SubmitFlights failed to submit flights to traveller")
	}
	if traveller.tripHistory.entries[2] != flights[0] {
		t.Error("SubmitFlights failed to submit flights to traveller")
	}
	if traveller.tripHistory.entries[3].Start != 0 {
		t.Error("SubmitFlights failed to submit flights to traveller")
	}
}

func TestEngineSubmitFlightsInBatches(t *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	var flights []Flight
	flights = append(flights,*createFlight(1,1,2),*createFlight(2,2,3))
	passport := NewPassport("987654321","uk")
	err := engine.SubmitFlights(passport,flights,SecondsInDay,true)
	if err != nil {
		t.Error("SubmitFlights failed")
	}
	traveller,err := engine.Travellers.GetTraveller(passport) 
	if err != nil {
		t.Error("SubmitFlights failed to create traveller")
	}
	if traveller.tripHistory.entries[0] != flights[1] {
		t.Error("SubmitFlights failed to submit flights to traveller")
	}
	if traveller.tripHistory.entries[1] != flights[0] {
		t.Error("SubmitFlights failed to submit flights to traveller")
	}
	if traveller.tripHistory.entries[2].Start != 0 {
		t.Error("SubmitFlights failed to submit flights to traveller")
	}
	var flights2 []Flight
	flights2 = append(flights2,*createFlight(3,3,4))
	err=engine.SubmitFlights(passport,flights2,SecondsInDay,true)
	if (err != nil) {
		t.Error("SubmitFlights failed to add to existing traveller")
	}
	traveller,err = engine.Travellers.GetTraveller(passport) 
	if err != nil {
		t.Error("SubmitFlights failed to get traveller")
	}
	if traveller.tripHistory.entries[0] != flights2[0] {
		t.Error("SubmitFlights failed to submit flights to traveller")
	}
	if traveller.tripHistory.entries[1] != flights[1] {
		t.Error("SubmitFlights failed to submit flights to traveller",traveller.tripHistory.entries[1],flights[0])
	}

	if traveller.tripHistory.entries[2] != flights[0] {
		t.Error("SubmitFlights failed to submit flights to traveller",traveller.tripHistory.entries[2],flights[1])
	}
	if traveller.tripHistory.entries[3].Start != 0 {
		t.Error("SubmitFlights failed to submit flights to traveller")
	}
}

func TestEngineSubmitFlightsGrounded(t *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	var flights []Flight
	flights = append(flights,*createFlight(1,1,2),*createFlight(2,2,3),*createFlight(3,3,4))
	passport := NewPassport("987654321","uk")
	err := engine.SubmitFlights(passport,flights,SecondsInDay,true)
	if err != nil {
		t.Error("SubmitFlights failed")
	}
	traveller,err := engine.Travellers.GetTraveller(passport) 
	if err != nil {
		t.Error("SubmitFlights failed to create traveller",err)
	}
	traveller.EndTrip()
	engine.Travellers.PutTraveller(traveller)
	thBefore := traveller.tripHistory
	moreFlights := append(flights,*createFlight(4,4,5))
	err = engine.SubmitFlights(passport,moreFlights,SecondsInDay,true)
	if err == nil {
		t.Error("SubmitFlights succeeded for grounded traveller",err)
	}
	traveller,err = engine.Travellers.GetTraveller(passport) 
	if err != nil {
		t.Error("SubmitFlights traveller disappeared",err)
	}
	if traveller.tripHistory != thBefore {
		t.Error("SubmitFlights changed TripHistory of grounded traveller")
	}
}

func TestUpdateTripsAndBackfillEmpty(t  *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	us,err := engine.UpdateTripsAndBackfill(1)
	if (err == nil) {
		t.Error("Update accepted now that isnt the start of a day")
	}
	us,err = engine.UpdateTripsAndBackfill(SecondsInDay)
	if err != nil {
		t.Error("Update failed with empty Travellers table",err)
	}
	if us.Travellers != 0 {
		t.Error("Update returned some travellers for an empty table",us.Travellers)
	}
	if us.Distance != 0 {
		t.Error("Update returned some distance for an empty table",us.Distance)
	}
	if us.Flights != 0 {
		t.Error("Update returned some flights for an empty table",us.Flights)
	}

	if us.Grounded != 0 {
		t.Error("Update returned some grounded travllers  for an empty table",err)
	}

}

func TestUpdateTripsAndBackfillOne(t  *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	paramsIn := FlapParams{DailyTotal:100, MinGrounded:1,FlightInterval:1,FlightsInTrip:50,TripLength:365}
	engine.Administrator.SetParams(paramsIn)
	var flights []Flight
	flights = append(flights,*createFlight(1,SecondsInDay,SecondsInDay+1),*createFlight(1,SecondsInDay*3,SecondsInDay*3+1))
	passport := NewPassport("987654321","uk")
	err := engine.SubmitFlights(passport,flights,SecondsInDay,true)
	us,err := engine.UpdateTripsAndBackfill(SecondsInDay*5)
	if err != nil {
		t.Error("Update failed for one Traveller",err)
	}
	traveller,_ := engine.Travellers.GetTraveller(passport) 
	if traveller.tripHistory.entries[0].et != etTripEnd {
		t.Error("UpdateTripsAndBackfill failed to end trip of one traveller",traveller.tripHistory.AsJSON())
	}
	expectedBalance := 100 - (flights[0].Distance+flights[1].Distance)
	if traveller.Balance !=  expectedBalance {
		t.Error("UpdateTripsAndBackfill didnt backfill correctly", traveller.Balance)
	}
	if engine.Administrator.bs.totalGrounded != 1 {
		t.Error("UpdateTripsAndBackfill set wrong value for totalGrounded", engine.Administrator.bs.totalGrounded)
	}
	if us.Travellers != 0 {
		t.Error("Update returned some travellers when no one travelled yesterday",err)
	}
	if us.Distance !=0 {
		t.Error("Update returned some distance when no one travelled yesterday",err)
	}
	if us.Flights !=0 {
		t.Error("Update returned some flights when no one travelled yesterday",err)
	}

	if us.Grounded != engine.Administrator.bs.totalGrounded {
		t.Error("Update returned wroung grounded value",us.Grounded)
	}


}

func TestUpdateTripsAndBackfillThreadedDatastore(t *testing.T) {
	for threads:=1; threads <= 16; threads *=2 {
		db := db.NewDatastoreDB("flaptest")
		if db == nil {
			t.Error("Failed to create db object")
		}
		testUpdateTripsThreaded(t,threads,db)
		dropTravellers(db)
		db.Release()
	}
}

func TestUpdateTripsAndBackfillThreaded(t *testing.T) {
	for threads:=1; threads <= 16; threads *=2 {
		db := enginesetup(t)
		testUpdateTripsThreaded(t,threads,db)
		engineteardown(db)
	}
}

func testUpdateTripsThreaded(t *testing.T,threads int, db db.Database) {
	engine := NewEngine(db,3,".")
	paramsIn := FlapParams{DailyTotal:100, MinGrounded:1,FlightInterval:1,FlightsInTrip:50,TripLength:365,Threads:byte(threads)}
	err := engine.Administrator.SetParams(paramsIn)
	if err != nil {
		t.Error("SetParams failed",err)
	}
	var flights13,flights2 []Flight
	passport1 := NewPassport("111111111","uk")
	flights13 = append(flights13,*createFlight(1,SecondsInDay,SecondsInDay+1),*createFlight(1,SecondsInDay*3,SecondsInDay*3+1))
	err = engine.SubmitFlights(passport1,flights13,SecondsInDay,true)
	passport2 := NewPassport("222222222","uk")
	flights2 = append(flights2,*createFlight(10,SecondsInDay,SecondsInDay+1),*createFlight(11,SecondsInDay*4,SecondsInDay*4+1))
	err = engine.SubmitFlights(passport2,flights2,SecondsInDay,true)
	passport3 := NewPassport("333333333","uk")
	err = engine.SubmitFlights(passport3,flights13,SecondsInDay,true)
	us,err := engine.UpdateTripsAndBackfill(SecondsInDay*5)
	if err != nil {
		t.Error("Update failed for three travellers",err)
	}
	traveller,_ := engine.Travellers.GetTraveller(passport1) 
	if traveller.tripHistory.entries[0].et != etTripEnd {
		t.Error("UpdateTripsAndBackfill failed to end trip of traveller 1",traveller.tripHistory.AsJSON())
	}
	expectedBalance := 100 - (flights13[0].Distance+flights13[1].Distance)
	if traveller.Balance !=  expectedBalance {
		t.Error("UpdateTripsAndBackfill didnt backfill traveller 1correctly", expectedBalance,traveller.Balance)
	}
	traveller,_ = engine.Travellers.GetTraveller(passport3) 
	if traveller.tripHistory.entries[0].et != etTripEnd {
		t.Error("UpdateTripsAndBackfill failed to end trip of traveller 3",traveller.tripHistory.AsJSON())
	}
	if traveller.Balance !=  expectedBalance {
		t.Error("UpdateTripsAndBackfill didnt backfill traveller 3 correctly", traveller.Balance)
	}
	traveller,_ = engine.Travellers.GetTraveller(passport2) 
	if traveller.tripHistory.entries[0].et != etFlight {
		t.Error("UpdateTripsAndBackfill ended trip of traveller 2",traveller.tripHistory.AsJSON())
	}
	expectedBalance = -(flights2[0].Distance+flights2[1].Distance)
	if traveller.Balance !=  expectedBalance {
		t.Error("UpdateTripsAndBackfill backfilled traveller 2", traveller.Balance)
	}
	if engine.Administrator.bs.totalGrounded != 2 {
		t.Error("UpdateTripsAndBackfill set wrong value for totalGrounded", engine.Administrator.bs.totalGrounded)
	}
	if us.Travellers != 1 {
		t.Error("Update returned no travellers when someone travelled yesterday",us.Travellers,flights2[1].Start)
	}
	if us.Distance != flights2[1].Distance {
		t.Error("Update returned incorrect distance travelled yesterday",us.Distance)
	}
	if us.Flights !=  1 {
		t.Error("Update returned Incorrect flight count",us.Flights)
	}
	if us.Grounded != 2 {
		t.Error("Update returned wrong value for grounded",us.Grounded)
	}
}

func TestUpdateTripsAndBackfillPromises(t  *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	paramsIn := FlapParams{DailyTotal:100, MinGrounded:5,FlightInterval:1,FlightsInTrip:50,TripLength:365,
		Promises: PromisesConfig{Algo:paLinearBestFit,MaxPoints:10}}
	engine.Administrator.SetParams(paramsIn)
	var flights []Flight
	flights = append(flights,*createFlight(1,SecondsInDay,SecondsInDay+1),*createFlight(1,SecondsInDay*3,SecondsInDay*3+1))
	passport := NewPassport("987654321","uk")
	err := engine.SubmitFlights(passport,flights,SecondsInDay,true)
	_,err = engine.UpdateTripsAndBackfill(SecondsInDay*5)
	if err != nil {
		t.Error("Update failed for one Traveller",err)
	}
	if reflect.TypeOf(engine.Administrator.predictor) == nil {
		t.Error("Failed to create predictor when promises are active")
	}
	if engine.Administrator.predictor.version() != 0 {
		t.Error("predictor has more than one point after one call to Update")
	}
	pexpected := engine.Administrator.predictor
	_,err = engine.UpdateTripsAndBackfill(SecondsInDay*6)
	if err != nil {
		t.Error("Second Update failed with promises activated",err)
	}
	if engine.Administrator.predictor != pexpected {
		t.Error("predictor replaced on second call to Update when promises are active")
	}
	if engine.Administrator.predictor.version() != 1 {
		t.Error("predictor add not successfully invoked twice when promises are active")
	}
	clearance,_ := engine.Administrator.predictor.predict(60,5)
	if clearance != 8 {
		t.Error("Update not populating predictor  with points to give expected prediction",clearance)
	}
}

func TestUpdateTripsAndBackfillKeepPromises(t  *testing.T) {
	
	// Create engine with promises enabled
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	paramsIn := FlapParams{DailyTotal:100, MinGrounded:5,FlightInterval:1,FlightsInTrip:50,TripLength:365,
		Promises:PromisesConfig{Algo:paLinearBestFit,MaxPoints:10,MaxDays:3},TaxiOverhead:100}
	engine.Administrator.SetParams(paramsIn)

	// Get and make a promise for flights
	var flights []Flight
	passport := NewPassport("987654321","uk")
	flights = append(flights,*createFlight(2,SecondsInDay*2,SecondsInDay*2+1),*createFlight(3,SecondsInDay*3,SecondsInDay*3+1))
	p,err := engine.Propose(passport,flights,0,SecondsInDay)
	if err != nil {
		t.Error("Couldnt propose promise for testing keep",err)
	}
	err = engine.Make(passport,p,SecondsInDay)
	if err != nil {
		t.Error("Couldnt make promise for testing keep",err)
	}
	
	// Submit flights
	err = engine.SubmitFlights(passport,flights,SecondsInDay,true)

	// Carry out Update on date when promise should be enforced
	_,err = engine.UpdateTripsAndBackfill(SecondsInDay*4)
	if err != nil {
		t.Error("Update failed when trying to test keep",err)
	}

	// Confirm traveller now has a clearance date
	traveller,err := engine.Travellers.GetTraveller(passport)
	if err != nil {
		t.Error("Failed to get traveller when testing keep")
	}
	if traveller.Kept.Clearance != SecondsInDay*4 {
		t.Error("UpdateTripsAndBackfill failed to set expected clearance date",traveller.Kept.Clearance)
	}

	// Check traveller is backfilled even though they are cleared to fly
	startbalance := traveller.Balance
	_,err = engine.UpdateTripsAndBackfill(SecondsInDay*4)
	if err != nil {
		t.Error("Update failed when trying to test keep",err)
	}
	traveller,err = engine.Travellers.GetTraveller(passport)
	if err != nil {
		t.Error("Failed to get traveller when testing keep")
	}
	if traveller.Balance != startbalance+20 {
		t.Error("Failed to backfill cleared traveller with negative balance",startbalance,traveller.Balance)
	}
	if traveller.Kept.Clearance != SecondsInDay*4 {
		t.Error("UpdateTripsAndBackfill failed to set expected clearance date",traveller.Kept.Clearance)
	}

}

func TestUpdateTripsAndBackfillGap(t  *testing.T) {
	
	// Create engine with promises enabled
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	paramsIn := FlapParams{DailyTotal:100, MinGrounded:5,FlightInterval:1,FlightsInTrip:50,TripLength:365,
		Promises: PromisesConfig{Algo:paLinearBestFit,MaxPoints:10,MaxDays:3}}
	engine.Administrator.SetParams(paramsIn)

	// Get and make a promise for flights
	var flights []Flight
	passport := NewPassport("987654321","uk")
	flights = append(flights,*createFlight(2,SecondsInDay*2,SecondsInDay*2+1),*createFlight(3,SecondsInDay*3,SecondsInDay*3+1))
	p,err := engine.Propose(passport,flights,0,SecondsInDay)
	if err != nil {
		t.Error("Couldnt propose promise for testing keep",err)
	}
	err = engine.Make(passport,p,SecondsInDay)
	if err != nil {
		t.Error("Couldnt make promise for testing keep",err)
	}
	
	// Submit flights
	err = engine.SubmitFlights(passport,flights,SecondsInDay,true)

	// Carry out Update on date when promise should be enforced
	_,err = engine.UpdateTripsAndBackfill(SecondsInDay*4)
	if err != nil {
		t.Error("Update failed when trying to test keep",err)
	}

	// Confirm traveller now has a clearance date
	traveller,err := engine.Travellers.GetTraveller(passport)
	if traveller.Kept.Clearance != SecondsInDay*4 {
		t.Error("UpdateTripsAndBackfill failed to set expected clearance date",traveller.Kept.Clearance)
	}
}

func TestPromisesCorrectBalances(t *testing.T) {
	
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	paramsIn := FlapParams{DailyTotal:100, MinGrounded:1,FlightInterval:1,FlightsInTrip:50,TripLength:365,
				Promises:PromisesConfig{Algo:paLinearBestFit,MaxPoints:100}}
	engine.Administrator.SetParams(paramsIn)
	engine.Administrator.pc.change(-25,0)
	us,err := engine.UpdateTripsAndBackfill(SecondsInDay*4)
	if err != nil {
		t.Error("Update failed when testing promises correction",err)
	}
	if us.Share != 100 {
		t.Error("Miscalculated share when not correcting promises",us.Share)
	}

	paramsIn.Promises.Algo = paLinearBestFit | pamCorrectDailyTotal
	engine.Administrator.SetParams(paramsIn)
	us,err = engine.UpdateTripsAndBackfill(SecondsInDay*4)
	if err != nil {
		t.Error("Update failed when testing promises correction",err)
	}
	if us.Share != 75 {
		t.Error("Miscalculated share when correcting promises",us.Share)
	}

}

func TestProposePromisesActive(t *testing.T) {
	
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	paramsIn := FlapParams{DailyTotal:100, MinGrounded:1,FlightInterval:1,FlightsInTrip:50,TripLength:365,
		Promises:PromisesConfig{Algo:paLinearBestFit,MaxPoints:10,MaxDays:100}}
	engine.Administrator.SetParams(paramsIn)
	passport := NewPassport("987654321","uk")

	var plannedflights []Flight
	plannedflights = append(plannedflights,*createFlight(2,SecondsInDay*2,SecondsInDay*2+1),*createFlight(3,SecondsInDay*3,SecondsInDay*3+1))
	p,err := engine.Propose(passport,plannedflights,0,SecondsInDay)
	if err != nil {
		t.Error("Propose not succeding when promises are enabled",err)
	}

	if p.entries[0].Clearance != SecondsInDay*4 {
		t.Error("Proposal doesnt include expected Clearance date",p.entries[0])
	}
	
	if p.entries[0].TripStart != SecondsInDay*2 {
		t.Error("Proposal doesnt include expected trip start date",p.entries[0])
	}

	if p.entries[0].TripEnd != SecondsInDay*3+1 {
		t.Error("Proposal doesnt include expected trip end date",p.entries[0])
	}

	if p.entries[0].Distance != plannedflights[0].Distance+plannedflights[1].Distance {
		t.Error("Proposal doesnt include expected trip distance",p.entries[0])
	}
	engine2 := NewEngine(db,0,"")
	engine2.Administrator.SetParams(paramsIn)
	if (!reflect.DeepEqual(*engine.Administrator.predictor.(*bestFit),*engine2.Administrator.predictor.(*bestFit))) {
		t.Error("predictor state not being persisted across engine instances")
	}
}

func TestProposePromisesTooFarAhead(t *testing.T) {
	
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	paramsIn := FlapParams{DailyTotal:100, MinGrounded:1,FlightInterval:1,FlightsInTrip:50,TripLength:365,
		Promises:PromisesConfig{Algo:paLinearBestFit,MaxPoints:10,MaxDays:1}}
	engine.Administrator.SetParams(paramsIn)
	passport := NewPassport("987654321","uk")

	var plannedflights []Flight
	plannedflights = append(plannedflights,*createFlight(3,SecondsInDay*3,SecondsInDay*3+1))
	_,err := engine.Propose(passport,plannedflights,0,SecondsInDay)
	if err != ETRIPTOOFARAHEAD {
		t.Error("Propose not rejecting trip that is too far ahead in time",err)
	}
}


func TestProposePromisesActiveiWithTripEnd(t *testing.T) {
	
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	paramsIn := FlapParams{DailyTotal:100, MinGrounded:1,FlightInterval:1,FlightsInTrip:50,TripLength:365,
		Promises:PromisesConfig{Algo:paLinearBestFit,MaxPoints:10,MaxDays:100}}
	engine.Administrator.SetParams(paramsIn)
	passport := NewPassport("987654321","uk")

	var plannedflights []Flight
	plannedflights = append(plannedflights,*createFlight(2,SecondsInDay*2,SecondsInDay*2+1),*createFlight(3,SecondsInDay*3,SecondsInDay*3+1))
	p,err := engine.Propose(passport,plannedflights,SecondsInDay*10,SecondsInDay)
	if err != nil {
		t.Error("Propose not succeding when promises are enabled",err)
	}

	if p.entries[0].Clearance != SecondsInDay*11 {
		t.Error("Proposal doesnt include expected Clearance date",p.entries[0])
	}
	
	if p.entries[0].TripStart != SecondsInDay*2 {
		t.Error("Proposal doesnt include expected trip start date",p.entries[0])
	}

	if p.entries[0].TripEnd != SecondsInDay*10 {
		t.Error("Proposal doesnt include expected trip end date",p.entries[0])
	}

	if p.entries[0].Distance != plannedflights[0].Distance+plannedflights[1].Distance {
		t.Error("Proposal doesnt include expected trip distance",p.entries[0])
	}

}

func TestProposePromisesInactive(t *testing.T) {
	
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	paramsIn := FlapParams{DailyTotal:100, MinGrounded:1,FlightInterval:1,FlightsInTrip:50,TripLength:365}
	engine.Administrator.SetParams(paramsIn)
	passport := NewPassport("987654321","uk")

	_,err := engine.Propose(passport,nil,0,0)
	if err != EINVALIDARGUMENT {
		t.Error("Propose accepted nil flight list as a valid argument",err)
	}

	var plannedflights []Flight
	_,err = engine.Propose(passport,plannedflights,0,SecondsInDay)
	if err != EINVALIDARGUMENT {
		t.Error("Propose accepted empty flight list as a valid argument",err)
	}

	plannedflights = append(plannedflights,*createFlight(2,SecondsInDay*3,SecondsInDay*3+1))
	_,err = engine.Propose(passport,plannedflights,0,SecondsInDay)
	if err != EPROMISESNOTENABLED  {
		t.Error("Propose succeding when promises aren't enabled",err)
	}
}

func TestMakePromisesInactive(t *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	paramsIn := FlapParams{DailyTotal:100, MinGrounded:1,FlightInterval:1,FlightsInTrip:50,TripLength:365}
	engine.Administrator.SetParams(paramsIn)
	passport := NewPassport("987654321","uk")

	err := engine.Make(passport,nil,SecondsInDay)
	if err != EINVALIDARGUMENT {
		t.Error("Make doesnt report expected error when nil proposal provided",err)
	}

	var p Proposal
	err = engine.Make(passport,&p,SecondsInDay)
	if err != EPROMISESNOTENABLED {
		t.Error("Make doesnt report expected error when promises aren't enabled",err)
	}

}

func TestMake(t *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	paramsIn := FlapParams{DailyTotal:100, MinGrounded:1,FlightInterval:1,FlightsInTrip:50,TripLength:365,
		Promises:PromisesConfig{Algo:paLinearBestFit,MaxPoints:10}}
	engine.Administrator.SetParams(paramsIn)
	passport := NewPassport("987654321","uk")

	var p Proposal
	fillpromises(&(p.Promises))
	err := engine.Make(passport,&p,SecondsInDay)
	if err != nil {
		t.Error("Make returns error when promises are enabled",err)
	}

	traveller,err := engine.Travellers.GetTraveller(passport) 
	if err != nil {
		t.Error("Make failed to create traveller")
	}

	if traveller.Created != SecondsInDay {
		t.Error("Make failed to set traveller creation date/time")
	}

	if !reflect.DeepEqual(p.Promises.entries,traveller.Promises.entries) {
		t.Error("Make failed to update promises in traveller record", traveller.Promises)
	}
}

func TestMakeOldProposal(t *testing.T) {

	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db,0,"")
	paramsIn := FlapParams{DailyTotal:100, MinGrounded:1,FlightInterval:1,FlightsInTrip:50,TripLength:365,
		Promises:PromisesConfig{Algo:paLinearBestFit,MaxPoints:10}}
	engine.Administrator.SetParams(paramsIn)
	passport := NewPassport("987654321","uk")

	var p Proposal
	engine.Administrator.predictor.add(1,50)
	engine.Administrator.predictor.add(2,15)
	err := engine.Make(passport,&p,SecondsInDay)
	if err != EPROPOSALEXPIRED {
		t.Error("Make returns error when promises are enabled",err)
	}
}

