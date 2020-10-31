package flap

import (
	"testing"
	"reflect"
	//"time"
)

func TestNewFlight(t *testing.T) {
	var from = Airport{}
	var to = Airport{}
	entry,err := NewFlight(from,0,to,0)
	if  err != EINVALIDARGUMENT {
		t.Error("NewFlight accepts zero start datetime")
	}
	entry,err = NewFlight(from,2,to,2)
	if  err != EINVALIDARGUMENT {
		t.Error("NewFlight accepts end equal to start")
	}
	entry,err = NewFlight(from,1,to,2)
	if  entry.Distance !=0 {
		t.Error("km !=0",entry)
	}
	if entry.Start!= 1  {
		t.Error("startTime != 1", entry)
	}
	if entry.End != 2 {
		t.Error("endTime != 2",entry)
	}
	if entry.et != etFlight {
		t.Error("et != etFlight",entry)
	}
}

func TestCalcDistZeros(t* testing.T) {
	ll1 := LatLon{
		Lat:0,
		Lon:0,
	}
	ll2 := LatLon{
		Lat:0,
		Lon:0,
	}
	kms,err := ll1.Distance(ll2)
	if kms != 0 {
		t.Error("distance != 0", kms)
	}
	if err != nil {
		t.Error("Error reported", err)
	}
}

func TestCalcDistInvalidLat(t* testing.T) {
	ll1 := LatLon{
		Lat:-91,
		Lon:0,
	}
	ll2 := LatLon{
		Lat:0,
		Lon:0,
	}
	_,err := ll1.Distance(ll2)
	if err != EINVALIDARGUMENT {
		t.Error("Invalid latitude accepted ", err)
	}
}

func TestCalcDistInvalidLong(t* testing.T) {
	ll1 := LatLon{
		Lat:0,
		Lon:0,
	}
	ll2 := LatLon{
		Lat:0,
		Lon:181,
	}
	_,err:= ll1.Distance(ll2)
	if err != EINVALIDARGUMENT {
		t.Error("Invalid longiture accepted", err)
	}
}

func TestCalcDistLHWtoCAI(t* testing.T) {
	ll1 := LatLon{
		Lat:51.470020,
		Lon:-0.454295,
	}
	ll2 := LatLon{
		Lat:30.12190055847168,
		Lon:31.40559959411621,
	}
	kms,err := ll1.Distance(ll2)
	if int64(kms) != 3533 {
		t.Error("Calculated distance isnt 3533 kms", kms)
	}
	if err != nil {
		t.Error("Error reported", err)
	}
}

func TestCalcDistCAItoLWH(t* testing.T) {
	ll1 := LatLon{
		Lat:51.470020,
		Lon:-0.454295,
	}
	ll2 := LatLon{
		Lat:30.12190055847168,
		Lon:31.40559959411621,
	}
	kms,err := ll2.Distance(ll1)
	if int64(kms) != 3533 {
		t.Error("Calculated distance isnt 3533 kms", kms)
	}
	if err != nil {
		t.Error("Error reported", err)
	}
}

func TextCreateTripHistory(t *testing.T) {
	var th TripHistory
	if len(th.entries) != 0 {
		t.Error("len(entries) !=0", th)
	}
}

func TestAddFlight(t *testing.T) {
	var th TripHistory
	f,_ := NewFlight(Airport{},1,Airport{},2)
	err := th.AddFlight(f)
	if err != nil {
		t.Error("AddFlight failed", err)
	}
}

func createFlight(i int, start int, end int) *Flight {
	f,_ :=  NewFlight(Airport{NewICAOCode(string(i+64)),LatLon{float64(i),float64(i)}},EpochTime(start),
		Airport{NewICAOCode(string(i+65)),LatLon{float64(i+1),float64(i+1)}},EpochTime(end))
	return f
}

func populateFlights(th *TripHistory, num int,interval int) error {
	for i := 1; i <= num; i+=2 {
		err := th.AddFlight(createFlight(i,i*interval,(i+1)*interval))
		if (err != nil) {
			return err
		}
	}
	for i := 2; i <=num; i+=2 {
		err:= th.AddFlight(createFlight(i,i*interval, (i+1)*interval))
		if (err != nil) {
			return err
		}
	}
	return nil
}

func checkFlight(t *testing.T,flight Flight, i int,interval int) {
	if  flight.Start != EpochTime(i*interval) {
		t.Error("Incorrect start", i,flight)
	}
	if flight.End !=  EpochTime((i+1)*interval) {
		t.Error("Incorrect end", flight)
	}
	if flight.FromAirport != NewICAOCode(string(i+64)) {
		t.Error("Incorrect from", flight)
	}
	if flight.ToAirport != NewICAOCode(string(i+65)) {
		t.Error("incorrect to", flight)
	}
}

func checkFlights(t *testing.T, th *TripHistory,start int,end int,interval int) {
	for i := start; i <= end; i++ {
		checkFlight(t,th.entries[end-i],i,interval)
	}
	for i := (end-start)+1; i< MaxFlights; i++ {
		if th.entries[i].Start != EpochTime(0) {
			t.Error("Flight not empty",i, th.entries[i])
		}
	}
}

func compareFlights(th1 *TripHistory,th2 *TripHistory) bool {
	return reflect.DeepEqual(th1.entries,th2.entries) 
}

func TestAddFlights(t *testing.T) {
	var th TripHistory
	err := populateFlights(&th,5,1)
	if err != nil {
		t.Error("Failed to populate flights")
	}
	checkFlights(t,&th,1,5,1)
}

func TestOldFlight(t *testing.T) {
	var th TripHistory
	err := populateFlights(&th,MaxFlights,10)
	if err != nil {
		t.Error("Failed to populate flights")
	}
	checkFlights(t,&th,1,MaxFlights,10)
	err = th.AddFlight(createFlight(0,1,2))
	if err != EFLIGHTTOOOLD {
		t.Error("Added old flight",err,th)
	}
}

func TestFullTripHistory(t *testing.T) {
	var th TripHistory
	err := populateFlights(&th,101,1)
	if err != nil {
		t.Error("Failed to populate flights")
	}
	checkFlights(t,&th,2,101,1)
}

func TestRemoveFlight(t *testing.T) {
	var th TripHistory
	th.AddFlight(createFlight(1,1,2))
	th.AddFlight(createFlight(2,1,2))
	th.AddFlight(createFlight(1,2,3))
	th.AddFlight(createFlight(2,2,3))
	e := th.RemoveFlight(createFlight(1,1,3))
	if e != EFLIGHTNOTFOUND {
		t.Error("Removed non-existent flight",th)
	}
	e = th.RemoveFlight(createFlight(1,1,2))
	if e != nil {
		t.Error("Failed to remove flight",th)
	}
	e = th.RemoveFlight(createFlight(1,1,2))
	if e != EFLIGHTNOTFOUND {
		t.Error("Removed flight twice",th)
	}
	e = th.RemoveFlight(createFlight(2,2,3))
	if e != nil {
		t.Error("Failed to remove flight",th)
	}
	e = th.RemoveFlight(createFlight(2,1,2))
	if e != nil {
		t.Error("Failed to remove flight",th)
	}
	e = th.RemoveFlight(createFlight(1,2,3))
	if e != nil {
		t.Error("Failed to remove flight",th)
	}
	var empty [MaxFlights]Flight
	if th.entries !=  empty {
		t.Error("All flights not removed", th)
	}
}

func TestEndTripNoReopenRespect(t *testing.T) {
	ts := tripState{reopened:false,start:SecondsInDay}
	f,_ := NewFlight(Airport{},1,Airport{},2)
	ts.endTrip(f, true)
	if f.et != etTripEnd {
		t.Error("endTrip flight type not set", f)
	}
	var tsEmpty tripState
	if !reflect.DeepEqual(ts,tsEmpty) {
		t.Error("endTrip tripState not empty",ts)
	}
}

func TestEndTripNoReopenNoRespect(t *testing.T) {
	ts := tripState{journeys:5,reopened:false,start:SecondsInDay}
	f,_ := NewFlight(Airport{},1,Airport{},2)
	ts.endTrip(f, false)
	if f.et != etTripEnd {
		t.Error("endTrip flight type not set", f)
	}
	var tsEmpty tripState
	if !reflect.DeepEqual(ts,tsEmpty) {
		t.Error("endTrip tripState not empty",ts)
	}
}

func TestEndTripTravellerEnd(t *testing.T) {
	ts := tripState{journeys:5,reopened:false,start:SecondsInDay}
	f,_ := NewFlight(Airport{},1,Airport{},2)
	f.et=etTravellerTripEnd
	ts.endTrip(f, false)
	if f.et != etTravellerTripEnd {
		t.Error("endTrip traveller end overwritten", f)
	}
	var tsEmpty tripState
	if !reflect.DeepEqual(ts,tsEmpty) {
		t.Error("endTrip tripState not empty",ts)
	}
}

func TestEndJourneyNotReopened(t *testing.T) {
	ts := tripState{journeys:5,reopened:false,start:SecondsInDay}
	f,_ := NewFlight(Airport{},1,Airport{},2)
	ts.endJourney(f)
	if f.et != etJourneyEnd {
		t.Error("endJourney failed to set correct state", f)
	}
	tsAfter := tripState{journeys:6,reopened:false,start:SecondsInDay}
	if !reflect.DeepEqual(ts,tsAfter) {
		t.Error("endJourney didnt increment journeys",ts)
	}
}

func TestEndJourneyReopened(t *testing.T) {
	ts := tripState{journeys:5,reopened:true,start:SecondsInDay}
	f,_ := NewFlight(Airport{},1,Airport{},2)
	f.et=etTripReopen
	tsBefore := ts
	ts.endJourney(f)
	if !reflect.DeepEqual(ts,tsBefore) {
		t.Error("endJourney reponen not respected",ts)
	}
	if f.et != etTripReopen {
		t.Error("endJourmey overwrting etTripReopened")
	}
}

func TestEndJourneyTripReopened(t *testing.T) {
	ts := tripState{journeys:5,reopened:true,start:SecondsInDay}
	f,_ := NewFlight(Airport{},1,Airport{},2)
	tsAfter := ts
	tsAfter.journeys++
	ts.endJourney(f)
	if !reflect.DeepEqual(ts,tsAfter) {
		t.Error("endJourney not working for reopened trip")
	}
	if f.et != etJourneyEnd {
		t.Error("endJourney not setting etJourneyEnd for repoened trip")
	}
}

func TestEndTripRespect(t *testing.T) {
	ts := tripState{journeys:5,reopened:true,start:SecondsInDay}
	f,_ := NewFlight(Airport{},1,Airport{},2)
	ts.endTrip(f, true)
	if f.et != etFlight {
		t.Error("endTrip flight not respecting reopen", f)
	}
	var tsEmpty tripState
	if reflect.DeepEqual(ts,tsEmpty) {
		t.Error("endTrip tripState not respecting reopen",ts)
	}
	ts.endTrip(f, false)
	if f.et != etTripEnd {
		t.Error("endTrip flight type not set", f)
	}
	if !reflect.DeepEqual(ts,tsEmpty) {
		t.Error("endTrip tripState not empty",ts)
	}
}

func TestUpdateJourney(t *testing.T) {
	ts := tripState{journeys:5,reopened:true,start:1}
	f,_ := NewFlight(Airport{},1,Airport{},2)
	params := FlapParams{FlightInterval:2}
	ts.updateJourney(f, SecondsInDay, &params,true)
	tsAfter := tripState{journeys:6,reopened:true,start:1}
	if reflect.DeepEqual(ts,tsAfter) {
		t.Error("updateJourney changing state when it shouldnt", ts)
	}
	ts.updateJourney(f,(SecondsInDay*2)+1,&params,true)
	if reflect.DeepEqual(ts,tsAfter) {
		t.Error("updateJourney changing state when it shouldnt", ts)
	}
	ts.updateJourney(f,SecondsInDay*2+2,&params,true)
	if !reflect.DeepEqual(ts,tsAfter) {
		t.Error("updateJourney not enforcing journey end", ts)
	}
	if f.et != etJourneyEnd {
		t.Error("updateJourney not setting flight state to etJourneyEnd",f)
	}
}

func TestUpdateJourneyReturnAirports(t *testing.T) {
	ts := tripState{journeys:5,reopened:true,start:1}
	params := FlapParams{FlightInterval:2}
	f := createFlight(1,1,2)
	ts.updateJourney(f,SecondsInDay, &params,false)
	if f.et == etJourneyEnd {
		t.Error("updateJourney not setting flight state to etJourneyEnd",f)
	}
	f = createFlight(3,3,4)
	ts.updateJourney(f,SecondsInDay, &params,false)
	if f.et == etJourneyEnd {
		t.Error("updateJourney not setting flight state to etJourneyEnd",f)
	}
	f = createFlight(2,5,6)
	ts.updateJourney(f,SecondsInDay, &params,true)
	if f.et != etJourneyEnd {
		t.Error("updateJourney not setting flight state to etJourneyEnd",f)
	}

}

func TestUpdateJourneySequence(t *testing.T) {
	ts := tripState{journeys:5,reopened:true,start:1}
	params := FlapParams{FlightInterval:1}
	f := createFlight(1,1,2)
	ts.updateJourney(f,SecondsInDay, &params,false)
	if f.et == etJourneyEnd {
		t.Error("updateJourney setting flight state to etJourneyEnd",f)
	}
	f = createFlight(3,3,4)
	ts.updateJourney(f,SecondsInDay*2, &params,false)
	if f.et != etJourneyEnd {
		t.Error("updateJourney not setting flight state to etJourneyEnd",f)
	}
	f = createFlight(2,5,6)
	ts.updateJourney(f,SecondsInDay, &params,true)
	if f.et == etJourneyEnd {
		t.Error("updateJourney setting flight state to etJourneyEnd",f)
	}
}

func TestUpdateTrip2Flights(t *testing.T) {
	ts := tripState{journeys:1,reopened:true,start:1}
	f,_ := NewFlight(Airport{},1,Airport{},2)
	params := FlapParams{TripLength:50,FlightsInTrip:50,FlightInterval:50}
	tsAfter := ts
	ts.updateTrip(f, SecondsInDay, &params)
	if !reflect.DeepEqual(ts,tsAfter) {
		t.Error("updateTrip incorrectlr changing trip state", ts)
	}
	ts.journeys =2
	tsAfter = ts
	ts.updateTrip(f, SecondsInDay, &params)
	if !reflect.DeepEqual(ts,tsAfter) {
		t.Error("updateTrip incorrectly changing trip state", ts)
	}
	if  f.et != etFlight {
		t.Error("updateTrip applying soft trip end to reopened trip.", f)
	}
	ts.reopened=false
	ts.updateTrip(f, SecondsInDay, &params)
	var emptyState tripState
	if !reflect.DeepEqual(ts,emptyState) {
		t.Error("updateTrip not resetting trip state on trip end", ts)
	}
	if  f.et != etTripEnd {
		t.Error("updateTrip not setting flight state at trip end", f)
	}
}

func TestUpdateTrip2FlightsWithPromises(t *testing.T) {
	ts := tripState{journeys:2,reopened:false,start:1}
	f,_ := NewFlight(Airport{},1,Airport{},2)
	params := FlapParams{TripLength:50,FlightsInTrip:50,FlightInterval:50,Promises:PromisesConfig{Algo:paLinearBestFit}}
	tsAfter := ts
	ts.updateTrip(f, SecondsInDay, &params)
	if  f.et != etFlight {
		t.Error("updateTrip applying soft trip end when running with promises enabled", f)
	}
	if !reflect.DeepEqual(ts,tsAfter) {
		t.Error("updateTrip incorrectly changing trip state", ts)
	}
}

func TestUpdateTripTripAge(t *testing.T) {
	ts := tripState{reopened:true,start:SecondsInDay}
	f,_ := NewFlight(Airport{},SecondsInDay,Airport{},SecondsInDay+1)
	tsAfter := ts
	params := FlapParams{TripLength:50,FlightsInTrip:10,FlightInterval:10}
	ts.updateTrip(f, SecondsInDay*51, &params)
	if !reflect.DeepEqual(ts,tsAfter) {
		t.Error("updateTrip incorrectly changing trip state", ts)
	}
	ts.updateTrip(f, SecondsInDay*52, &params)
	var emptyState tripState
	if !reflect.DeepEqual(ts,emptyState) {
		t.Error("updateTrip not resetting trip state on trip end", ts)
	}
	if  f.et != etTripEnd {
		t.Error("updateTrip not setting flight state at trip end", f)
	}
}

func TestUpdateTripFlightsInTrip(t *testing.T) {
	ts := tripState{reopened:true,flights:1,start:SecondsInDay}
	f,_ := NewFlight(Airport{},SecondsInDay,Airport{},SecondsInDay+1)
	params := FlapParams{TripLength:50,FlightsInTrip:10,FlightInterval:5}
	tsAfter := ts
	ts.updateTrip(f, SecondsInDay, &params)
	if !reflect.DeepEqual(ts,tsAfter) {
		t.Error("updateTrip incorrectly changing trip state", ts)
	}
	ts.flights=10
	ts.updateTrip(f, SecondsInDay, &params)
	var emptyState tripState
	if !reflect.DeepEqual(ts,emptyState) {
		t.Error("updateTrip not resetting trip state on trip end", ts)
	}
	if  f.et != etTripEnd {
		t.Error("updateTrip not setting flight state at trip end", f)
	}
}

func TestUpdateTripEmpty(t *testing.T) {
	var tsEmpty tripState
	f,_ := NewFlight(Airport{},SecondsInDay-1,Airport{},SecondsInDay)
	params := FlapParams{TripLength:50,FlightsInTrip:50,FlightInterval:2}
	tsEmpty.updateTrip(f, SecondsInDay, &params)
	tsAfter := tripState{start:SecondsInDay-1}
	if !reflect.DeepEqual(tsEmpty,tsAfter) {
		t.Error("updateTrip not setting tripstart", tsEmpty)
	}
	if f.et != etFlight {
		t.Error("updateTrup incorrectly changing flight type",f)
	}
}

func TestUpdateEmpty(t *testing.T) {
	var th TripHistory
	var params FlapParams
	d,err := th.Update(&params,0)
	if err == nil {
		t.Error("No error on empty history")
	}
	if d != 0 {
		t.Error("Returned distance for empty history")
	}
}

func TestUpdateNowNotWholeDays(t *testing.T) {
	var th TripHistory
	populateFlights(&th,1,1)
	var params FlapParams
	_,err := th.Update(&params,SecondsInDay-1)
	if err == nil {
		t.Error("No error on now not whole days")
	}
}

func TestUpdateOneFlight(t *testing.T) {
	var th TripHistory
	params := FlapParams{TripLength:50,FlightsInTrip:50,FlightInterval:1}
	populateFlights(&th,1,1)
	d,err := th.Update(&params,SecondsInDay)
	if err != nil {
		t.Error("Update failed")
	}
	checkFlights(t,&th,1,1,1)
	if th.entries[0].et != etFlight {
		t.Error("Changed flight state",th.entries[0])
	}
	if d != th.entries[0].Distance {
		t.Error("Returned wrong distance for one flight today",d)
	}
	d,err = th.Update(&params,SecondsInDay*2)
	if err != nil {
		t.Error("Update failed")
	}
	checkFlights(t,&th,1,1,1)
	if th.entries[0].et != etJourneyEnd {
		t.Error("Failed to end journey with one flight",th.entries[0])
	}
	if d != 0 {
		t.Error("Returned distance for no flights today",d)
	}
}

func TestUpdateOneTrip(t *testing.T) {
	var th TripHistory
	params := FlapParams{TripLength:50,FlightsInTrip:50,FlightInterval:1}
	th.AddFlight(createFlight(1,SecondsInDay,SecondsInDay+1))
	th.AddFlight(createFlight(2,SecondsInDay*2+1,SecondsInDay*2+3))
	th2 := th
	th2.entries[1].et = etJourneyEnd
	th2.entries[0].et = etTripEnd
	_,err := th.Update(&params,SecondsInDay*4)
	if (err != nil) {
		t.Error("Update failed", err)
	}
	if !compareFlights(&th,&th2) {
		t.Error("Update failed to end trip",th)
	}
	_,err = th.Update(&params,SecondsInDay*5)
	if (err != ENOCHANGEREQUIRED) {
		t.Error("Update didnt realise no update required", err)
	}
	if !compareFlights(&th,&th2) {
		t.Error("Update failed to retain trip end a day later",th)
	}

}

func TestUpdateTwoTrips(t *testing.T) {
	var th TripHistory
	params := FlapParams{TripLength:50,FlightsInTrip:50,FlightInterval:1}
	th.AddFlight(createFlight(1,SecondsInDay,SecondsInDay+1))
	th.AddFlight(createFlight(2,SecondsInDay*2+1,SecondsInDay*2+3))
	th.AddFlight(createFlight(3,SecondsInDay*3+4,SecondsInDay*3+5))
	th.AddFlight(createFlight(4,SecondsInDay*4+6,SecondsInDay*4+7))
	th2 := th
	th2.entries[3].et = etJourneyEnd
	th2.entries[2].et = etTripEnd
	th2.entries[1].et = etJourneyEnd
	th2.entries[0].et = etTripEnd
	_,err := th.Update(&params,SecondsInDay*6)
	if (err != nil) {
		t.Error("Update failed", err)
	}
	if !compareFlights(&th,&th2) {
		t.Error("Update failed to end trip",th)
	}
}

func TestUpdateReopedTrip(t *testing.T) {
	var th TripHistory
	params := FlapParams{TripLength:50,FlightsInTrip:50,FlightInterval:1}
	th.AddFlight(createFlight(1,SecondsInDay,SecondsInDay+1))
	th.AddFlight(createFlight(2,SecondsInDay*2+1,SecondsInDay*2+3))
	th.AddFlight(createFlight(3,SecondsInDay*3+4,SecondsInDay*3+5))
	th.AddFlight(createFlight(4,SecondsInDay*4+6,SecondsInDay*4+7))
	th.entries[3].et=etTripReopen
	th2 := th
	th2.entries[2].et = etJourneyEnd
	th2.entries[1].et = etJourneyEnd
	th2.entries[0].et = etJourneyEnd
	_,err := th.Update(&params,SecondsInDay*6)
	if (err != nil) {
		t.Error("Update failed", err)
	}
	if !compareFlights(&th,&th2) {
		t.Error("Update failed to keep reopened trip open",th)
	}
}

func TestUpdateMultiFlightJourneys(t *testing.T) {
	var th TripHistory
	params := FlapParams{TripLength:50,FlightsInTrip:50,FlightInterval:1}
	th.AddFlight(createFlight(1,1,10))
	th.AddFlight(createFlight(2,10,30))
	th.AddFlight(createFlight(3,SecondsInDay*2, (SecondsInDay*2)+1))
	th.AddFlight(createFlight(4,(SecondsInDay*3)-2,(SecondsInDay*3)-1))
	th2:=th
	th2.entries[0].et = etTripEnd
	th2.entries[2].et = etJourneyEnd
	_,err := th.Update(&params,SecondsInDay*4)
	if (err != nil) {
		t.Error("Update failed", err)
	}
	if !compareFlights(&th,&th2) {
		t.Error("Update failed to end trip and journeys",th)
	}
}

func TestUpdateTripLength(t *testing.T) {
	var th TripHistory
	params := FlapParams{TripLength:2,FlightsInTrip:50,FlightInterval:10}
	th.AddFlight(createFlight(1,SecondsInDay,SecondsInDay+1))
	th.AddFlight(createFlight(2,SecondsInDay*2+1,SecondsInDay*2+3))
	th.AddFlight(createFlight(3,SecondsInDay*3+4,SecondsInDay*3+5))
	th.AddFlight(createFlight(4,SecondsInDay*4+6,SecondsInDay*4+7))
	th2 := th
	th2.entries[1].et = etTripEnd
	_,err := th.Update(&params,SecondsInDay*5)
	if (err != nil) {
		t.Error("Update failed", err)
	}
	if !compareFlights(&th,&th2) {
		t.Error("Update failed to enforce trip length",th)
	}
}

func TestUpdateTripLengthLeftReopened(t *testing.T) {
	var th TripHistory
	params := FlapParams{TripLength:10,FlightsInTrip:50,FlightInterval:10}
	th.AddFlight(createFlight(1,SecondsInDay,SecondsInDay+1))
	th.AddFlight(createFlight(2,SecondsInDay*2,SecondsInDay*2+1))
	th.entries[0].et = etTripReopen
	_,err := th.Update(&params,SecondsInDay*11)
	if (err != nil) {
		t.Error("Update failed", err)
	}
	if th.entries[0].et != etTripReopen {
		t.Error("Update closeed trip left reopened too early",th)
	}
	_,err = th.Update(&params,SecondsInDay*12)
	if (err != nil) {
		t.Error("Update failed", err)
	}
	if th.entries[0].et != etTripEnd {
		t.Error("Update failed to close trip left reopened",th)
	}
}

func TestUpdateTravellerTripEnd(t *testing.T) {
	var th TripHistory
	params := FlapParams{TripLength:50,FlightsInTrip:50,FlightInterval:1}
	th.AddFlight(createFlight(1,1,10))
	th.AddFlight(createFlight(2,10,30))
	th.AddFlight(createFlight(3,SecondsInDay*2, (SecondsInDay*2)+1))
	th.AddFlight(createFlight(4,(SecondsInDay*3)-2,(SecondsInDay*3)-1))
	th.entries[2].et = etTravellerTripEnd
	th2:=th
	th2.entries[0].et = etJourneyEnd
	_,err := th.Update(&params,SecondsInDay*4)
	if (err != nil) {
		t.Error("Update failed", err)
	}
	if !compareFlights(&th,&th2) {
		t.Error("Update didnt respect trip closed by traveller",th)
	}
}

func TestUpdateFlightsInTrip(t *testing.T) {
	var th TripHistory
	params := FlapParams{TripLength:50,FlightsInTrip:10, FlightInterval:1}
	populateFlights(&th,10,1)
	th2 := th
	th2.entries[0].et=etTripEnd
	_,err := th.Update(&params,SecondsInDay*4)
	if (err != nil) {
		t.Error("Update failed", err)
	}
	if !compareFlights(&th,&th2) {
		t.Error("Update didnt enforce flights in trip",th)
	}
}

func TestUpdateFlightsFullHistory(t *testing.T) {
	var th TripHistory
	params := FlapParams{TripLength:1,FlightsInTrip:50, FlightInterval:1}
	populateFlights(&th,101,1)
	th.entries[99].et=etTripReopen
	th.entries[50].et=etTripReopen
	th2 := th
	th2.entries[50].et=etTripEnd
	th2.entries[0].et=etTripEnd
	_,err := th.Update(&params,SecondsInDay)
	if (err != nil) {
		t.Error("Update failed", err)
	}
	if !compareFlights(&th,&th2) {
		t.Error("Update failed to update flights in full history",th)
	}
}


func TestEndTripEmpty(t *testing.T) {
	var th TripHistory
	err := th.EndTrip()
	if err == nil {
		t.Error("EndTrip successful with empty history")
	}
}

func TestEndTrip(t *testing.T) {
	var th TripHistory
	populateFlights(&th,1,1)
	err := th.EndTrip()
	if err != nil {
		t.Error(err.Error())
	}
	if th.entries[0].et != etTravellerTripEnd {
		t.Error("Latest flight not set to trip end",th.entries[0])
	}
}

func TestRepoenTripEmpty(t *testing.T) {
	var th TripHistory
	err := th.ReopenTrip()
	if err == nil {
		t.Error("Reopen successful with empty history")
	}
}

func TestReopenOpenTrip(t *testing.T) {
	var th TripHistory
	populateFlights(&th,1,1)
	err := th.ReopenTrip()
	if err == nil {
		t.Error("Reopen successful when trip not closed")
	}
}

func TestReopenTrip(t *testing.T) {
	var th TripHistory
	populateFlights(&th,1,1)
	th.entries[0].et = etTripEnd
	err := th.ReopenTrip()
	if err != nil {
		t.Error("Failed to reopen trip")
	}
	if th.entries[0].et != etTripReopen {
		t.Error("ReopenTrip didnt work",th.entries[0])
	}
}

func TestReopenTravellerTrip(t *testing.T) {
	var th TripHistory
	populateFlights(&th,1,1)
	th.entries[0].et = etTravellerTripEnd
	err := th.ReopenTrip()
	if err != nil {
		t.Error("Failed to reopen trip")
	}
	if th.entries[0].et != etTripReopen {
		t.Error("ReopenTrip didnt work",th.entries[0])
	}
}

func TestEmpty(t *testing.T) {
	var th TripHistory
	if !th.empty() {
		t.Error("empty returning false")
	}
	populateFlights(&th,1,1)
	if th.empty() {
		t.Error("empty returning true")
	}
}

func TeststartOfTripEmpty(t *testing.T) {
	var th TripHistory
	_,err := th.startOfTrip(0)
	if err == nil {
		t.Error("startOfTrip working for an empty history")
	}
}

func TeststartOfTripInvalid(t *testing.T) {
	var th TripHistory
	_,err := th.startOfTrip(MaxFlights)
	if err == nil {
		t.Error("Invalid index accepted",err)
	}
}

func TeststartOfTrip2Flights(t *testing.T) {
	var th TripHistory
	th.AddFlight(createFlight(1,SecondsInDay,SecondsInDay+1))
	th.AddFlight(createFlight(2,SecondsInDay*2+1,SecondsInDay*2+3))
	j,err := th.startOfTrip(0)
	if err != nil {
		t.Error("startOfTrip failed for 2 flight history")
	}
	if j != 1 {
		t.Error("startOfTrip returned wrong value for 2 flight history")
	}
	j,err = th.startOfTrip(1)
	if err != nil {
		t.Error("startOfTrip failed for 2 flight history")
	}
	if j != 1 {
		t.Error("startOfTrip returned wrong value for 2 flight history")
	}

}

func TeststartOfTripFull(t *testing.T) {
	var th TripHistory
	populateFlights(&th,100,1)
	i,_:= th.startOfTrip(0)
	if i != 99 {
		t.Error("startOfTrip not returning 99",)
	}
	i,_ = th.startOfTrip(99)
	if i != 99 {
		t.Error("startOfTrip not returning 99",)
	}
}

func TeststartOfTripEnds(t *testing.T) {
	var th TripHistory
	populateFlights(&th,10,1)
	th.entries[3].et= etTripEnd
	th.entries[8].et= etTravellerTripEnd
	i,_ := th.startOfTrip(0)
	if i != 2 {
		t.Error("startOfTrip not returning 2",i)
	}
	i,_ = th.startOfTrip(2)
	if i != 2 {
		t.Error("startOfTrip not returning 2",i)
	}
	i,_ = th.startOfTrip(4)
	if i != 7 {
		t.Error("startOfTrip not returning 7",i)
	}
}

func TeststartOfTripNoEnd(t *testing.T) {
	var th TripHistory
	populateFlights(&th,3,1)
	i,_ := th.startOfTrip(0)
	if i != 2 {
		t.Error("startOfTrip not returning 2",)
	}
	i,_ = th.startOfTrip(1)
	if i != 2 {
		t.Error("startOfTrip not returning 2",)
	}
}

func TeststartOfTripOneFlight(t *testing.T) {
	var th TripHistory
	populateFlights(&th,1,1)
	i,_ := th.startOfTrip(0)
	if i != 0 {
		t.Error("startOfTrip not returning 0",)
	}
}

func TestTripStartEndLengthEmpty(t *testing.T) {
	var th TripHistory
	s,e,d := th.tripStartEndLength()
	if !(s==0 && e==0 && d==0) {
		t.Error("tripStartEndDistance not returning zeros for empty history",s,e,d)
	}
}

func TestTripStartEndLengthOneFlight(t *testing.T) {
	var th TripHistory
	populateFlights(&th,1,1)
	th.entries[0].Distance = 5
	s,e,d := th.tripStartEndLength()
	if !(s==1 && e==2 && d==5) {
		t.Error("tripStartEndDistance returning incorrect value for 1 flight",s,e,d)
	}
}

func TestTripStartEndLengthThreeFlights(t *testing.T) {
	var th TripHistory
	populateFlights(&th,3,1)
	th.entries[0].Distance = 5
	th.entries[1].Distance = 6
	th.entries[2].Distance = 7
	s,e,d := th.tripStartEndLength()
	if !(s==1 && e==4 && d==18) {
		t.Error("tripStartEndDistance returning incorrect value for 1 flight",s,e,d)
	}
}

func TestTripStartEndLengthTwoTrips(t *testing.T) {
	var th TripHistory
	populateFlights(&th,3,1)
	th.entries[0].Distance = 5
	th.entries[2].et = etTripEnd
	th.entries[1].Distance = 6
	th.entries[2].Distance = 7
	s,e,d := th.tripStartEndLength()
	if !(s==2 && e==4 && d==11) {
		t.Error("tripStartEndDistance returning incorrect value for two trips",s,e,d)
	}
}

func TestTripStartEndLengthTripEnd(t *testing.T) {
	var th TripHistory
	populateFlights(&th,1,1)
	th.entries[0].et = etTravellerTripEnd
	s,e,d := th.tripStartEndLength()
	if !(s==0 && e==0 && d==0) {
		t.Error("tripStartEndDistance returning values when there is no open trip",s,e,d)
	}
}

