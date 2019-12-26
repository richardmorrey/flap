package flap

import (
	"testing"
	"os"
	"github.com/richardmorrey/flap/pkg/db"
	"reflect"
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
	engine := NewEngine(db)
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
	engine := NewEngine(db)
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
	engine := NewEngine(db)
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
	engine := NewEngine(db)
	paramsIn := FlapParams{TripLength:200,FlightsInTrip:50,FlightInterval:101,DailyTotal:1000}
	err := engine.Administrator.SetParams(paramsIn)
	if err == nil {
		t.Error("Invalid flight interval accepted",paramsIn)
	}
}

func TestInvalidFlightInTrip(t  *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db)
	paramsIn := FlapParams{TripLength:365,FlightsInTrip:51,FlightInterval:2,DailyTotal:1000}
	err := engine.Administrator.SetParams(paramsIn)
	if err == nil {
		t.Error("Invalid flights in trip  accepted",paramsIn)
	}
}

func TestEngineSubmitFlightsEmpty(t *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db)
	var flights []Flight
	passport := NewPassport("987654321","uk")
	err := engine.SubmitFlights(passport,flights,0,true)
	if err == nil {
		t.Error("Allowed to add empty list of flights")
	}
}

func TestEngineSubmitFlights(t *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db)
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
	if traveller.tripHistory.entries[3].start != 0 {
		t.Error("SubmitFlights failed to submit flights to traveller")
	}
}

func TestEngineSubmitFlightsInBatches(t *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db)
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
	if traveller.tripHistory.entries[2].start != 0 {
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
	if traveller.tripHistory.entries[3].start != 0 {
		t.Error("SubmitFlights failed to submit flights to traveller")
	}
}

func TestEngineSubmitFlightsGrounded(t *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db)
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
	engine := NewEngine(db)
	ts,d,g,err := engine.UpdateTripsAndBackfill(1)
	if (err == nil) {
		t.Error("Update accepted now that isnt the start of a day")
	}
	ts,d,g,err = engine.UpdateTripsAndBackfill(SecondsInDay)
	if err != nil {
		t.Error("Update failed with empty Travellers table",err)
	}
	if ts != 0 {
		t.Error("Update returned some travellers for an empty table",err)
	}
	if d != 0 {
		t.Error("Update returned some distance for an empty table",err)
	}
	if g != 0 {
		t.Error("Update returned some grounded travllers  for an empty table",err)
	}

}

func TestUpdateTripsAndBackfillOne(t  *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db)
	paramsIn := FlapParams{DailyTotal:100, MinGrounded:1,FlightInterval:1,FlightsInTrip:50,TripLength:365}
	engine.Administrator.SetParams(paramsIn)
	var flights []Flight
	flights = append(flights,*createFlight(1,SecondsInDay,SecondsInDay+1),*createFlight(1,SecondsInDay*3,SecondsInDay*3+1))
	passport := NewPassport("987654321","uk")
	err := engine.SubmitFlights(passport,flights,SecondsInDay,true)
	ts,d,g,err := engine.UpdateTripsAndBackfill(SecondsInDay*5)
	if err != nil {
		t.Error("Update failed for one Traveller",err)
	}
	traveller,_ := engine.Travellers.GetTraveller(passport) 
	if traveller.tripHistory.entries[0].et != etTripEnd {
		t.Error("UpdateTripsAndBackfill failed to end trip of one traveller",traveller.tripHistory.AsJSON())
	}
	expectedBalance := 100 - (flights[0].distance+flights[1].distance)
	if traveller.balance !=  expectedBalance {
		t.Error("UpdateTripsAndBackfill didnt backfill correctly", traveller.balance)
	}
	if engine.totalGrounded != 1 {
		t.Error("UpdateTripsAndBackfill set wrong value for totalGrounded", engine.totalGrounded)
	}
	if ts != 0 {
		t.Error("Update returned some travellers when no one travelled yesterday",err)
	}
	if d !=0 {
		t.Error("Update returned some distance when no one travelled yesterday",err)
	}
	if g != engine.totalGrounded {
		t.Error("Update returned wroung grounded value",g)
	}


}

func TestUpdateTripsAndBackfillThree(t  *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db)
	paramsIn := FlapParams{DailyTotal:100, MinGrounded:1,FlightInterval:1,FlightsInTrip:50,TripLength:365}
	engine.Administrator.SetParams(paramsIn)
	var flights13,flights2 []Flight
	passport1 := NewPassport("111111111","uk")
	flights13 = append(flights13,*createFlight(1,SecondsInDay,SecondsInDay+1),*createFlight(1,SecondsInDay*3,SecondsInDay*3+1))
	err := engine.SubmitFlights(passport1,flights13,SecondsInDay,true)
	passport2 := NewPassport("222222222","uk")
	flights2 = append(flights2,*createFlight(10,SecondsInDay,SecondsInDay+1),*createFlight(11,SecondsInDay*4,SecondsInDay*4+1))
	err = engine.SubmitFlights(passport2,flights2,SecondsInDay,true)
	passport3 := NewPassport("333333333","uk")
	err = engine.SubmitFlights(passport3,flights13,SecondsInDay,true)
	ts,d,g,err := engine.UpdateTripsAndBackfill(SecondsInDay*5)
	if err != nil {
		t.Error("Update failed for three travellers",err)
	}
	traveller,_ := engine.Travellers.GetTraveller(passport1) 
	if traveller.tripHistory.entries[0].et != etTripEnd {
		t.Error("UpdateTripsAndBackfill failed to end trip of traveller 1",traveller.tripHistory.AsJSON())
	}
	expectedBalance := 100 - (flights13[0].distance+flights13[1].distance)
	if traveller.balance !=  expectedBalance {
		t.Error("UpdateTripsAndBackfill didnt backfill traveller 1correctly", traveller.balance)
	}
	traveller,_ = engine.Travellers.GetTraveller(passport3) 
	if traveller.tripHistory.entries[0].et != etTripEnd {
		t.Error("UpdateTripsAndBackfill failed to end trip of traveller 3",traveller.tripHistory.AsJSON())
	}
	if traveller.balance !=  expectedBalance {
		t.Error("UpdateTripsAndBackfill didnt backfill traveller 3 correctly", traveller.balance)
	}
	traveller,_ = engine.Travellers.GetTraveller(passport2) 
	if traveller.tripHistory.entries[0].et != etFlight {
		t.Error("UpdateTripsAndBackfill ended trip of traveller 2",traveller.tripHistory.AsJSON())
	}
	expectedBalance = -(flights2[0].distance+flights2[1].distance)
	if traveller.balance !=  expectedBalance {
		t.Error("UpdateTripsAndBackfill backfilled traveller 2", traveller.balance)
	}
	if engine.totalGrounded != 2 {
		t.Error("UpdateTripsAndBackfill set wrong value for totalGrounded", engine.totalGrounded)
	}
	if ts != 1 {
		t.Error("Update returned no travellers when someone travelled yesterday",ts,flights2[1].start)
	}
	if d != flights2[1].distance {
		t.Error("Update returned no distance when someone travelled yesterday",d)
	}
	if g != 2 {
		t.Error("Update returned wrong value for grounded",g)
	}

}

func TestUpdateTripsAndBackfillPromises(t  *testing.T) {
	db:= enginesetup(t)
	defer engineteardown(db)
	engine := NewEngine(db)
	paramsIn := FlapParams{DailyTotal:100, MinGrounded:5,FlightInterval:1,FlightsInTrip:50,TripLength:365,
				PromisesAlgo:paLinearBestFit,PromisesMaxPoints:10}
	engine.Administrator.SetParams(paramsIn)
	var flights []Flight
	flights = append(flights,*createFlight(1,SecondsInDay,SecondsInDay+1),*createFlight(1,SecondsInDay*3,SecondsInDay*3+1))
	passport := NewPassport("987654321","uk")
	err := engine.SubmitFlights(passport,flights,SecondsInDay,true)
	_,_,_,err = engine.UpdateTripsAndBackfill(SecondsInDay*5)
	if err != nil {
		t.Error("Update failed for one Traveller",err)
	}
	if engine.predictor == nil {
		t.Error("Failed to create predictor when promises are active")
	}
	if engine.predictor.version() != 0 {
		t.Error("predictor has more than one point after one call to Update")
	}
	pexpected := engine.predictor
	_,_,_,err = engine.UpdateTripsAndBackfill(SecondsInDay*6)
	if err != nil {
		t.Error("Second Update failed with promises activated",err)
	}
	if engine.predictor != pexpected {
		t.Error("predictor replaced on second call to Update when promises are active")
	}
	if engine.predictor.version() != 1 {
		t.Error("predictor add not successfully invoked twice when promises are active")
	}
	clearance,_ := engine.predictor.predict(60,5)
	if clearance != 8 {
		t.Error("Update not populating predictor  with points to give expected prediction",clearance)
	}
}

