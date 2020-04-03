package flap

import (
	"testing"
	//"io/ioutil"
	//"path/filepath"
	"os"
	"github.com/richardmorrey/flap/pkg/db"
	"reflect"
)

var TRAVELLERSTESTFOLDER="travellerstest"

func travellerssetup(t *testing.T) *db.LevelDB{
	if err := os.Mkdir(TRAVELLERSTESTFOLDER, 0700); err != nil {
		t.Error("Failed to create test dir", err)
	}
	NewLogger(llDebug,".")
	return db.NewLevelDB(TRAVELLERSTESTFOLDER)
}

func travellersteardown(db *db.LevelDB) {
	db.Release()
	os.RemoveAll(TRAVELLERSTESTFOLDER)
}

func TestNewTravellers(t *testing.T) {
	db:=travellerssetup(t)
	defer travellersteardown(db)
	travellers := NewTravellers(db)
	if travellers == nil {
		t.Error("Failed to create Travellers")
	}
	table,err:= db.OpenTable("travellers")
	if err != nil || table == nil {
		t.Error("Travellers table not created")
	}
}

func TestPutGetEmptyTraveller(t  *testing.T) {
	db:= travellerssetup(t)
	defer travellersteardown(db)
	travellers := NewTravellers(db)
	var passport Passport
	var travellerin Traveller
	travellerin.passport = passport
	err := travellers.PutTraveller(travellerin)
	if err != nil {
		t.Error("PutTraveller failed",err)
	}
	travellerout,err := travellers.GetTraveller(passport)
	if err != nil {
		t.Error("GetTraveller failed", err)
	}
	if !(reflect.DeepEqual(travellerin,travellerout)) {
		t.Error("Got traveller doesnt equal put traveller", travellerout)
	}
}

func TestPutGetFullTraveller(t  *testing.T) {
	db:= travellerssetup(t)
	defer travellersteardown(db)
	travellers := NewTravellers(db)
	passport := NewPassport("012345678","uk")
	var travellerin  Traveller
	travellerin.passport = passport
	populateFlights(&(travellerin.tripHistory),150,1)
	err := travellers.PutTraveller(travellerin)
	if err != nil {
		t.Error("PutTraveller failed",err)
	}
	travellerout,err := travellers.GetTraveller(passport)
	if err != nil {
		t.Error("GetTraveller failed", err)
	}
	if !(reflect.DeepEqual(travellerin,travellerout)) {
		t.Error("Got traveller doesnt equal put traveller", travellerin,travellerout)
	}
}

func TestPutGetSnapshot(t  *testing.T) {
	db:= travellerssetup(t)
	defer travellersteardown(db)
	travellers := NewTravellers(db)
	passport := NewPassport("012345678","uk")
	var travellerin  Traveller
	travellerin.passport = passport
	populateFlights(&(travellerin.tripHistory),150,1)
	err := travellers.PutTraveller(travellerin)
	if err != nil {
		t.Error("PutTraveller failed",err)
	}
	ss,err := travellers.TakeSnapshot()
	if err != nil {
		t.Error("Failed to take snapshot")
	}
	defer ss.Release()
	travellerout,err := ss.Get(passport)
	if err != nil {
		t.Error("GetTraveller failed with snapshot", err)
	}
	if !(reflect.DeepEqual(travellerin,travellerout)) {
		t.Error("Got traveller doesnt equal put traveller with snapshot", travellerout)
	}
}

func TestCleared0(t *testing.T) {
	var traveller Traveller
	if (!traveller.Cleared(0)) {
		t.Error("Traveller not cleared",traveller)
	}
}

func TestCleared1(t *testing.T) {
	var traveller Traveller
	if (!traveller.Cleared(1)) {
		t.Error("Traveller not cleared",traveller)
	}
}

func TestCleared2Default(t *testing.T) {
	var traveller Traveller
	traveller.kept.promise.Clearance=1
	traveller.balance=-1
	populateFlights(&(traveller.tripHistory),1,1)
	traveller.EndTrip()
	if (traveller.Cleared(0)) {
		t.Error("Traveller cleared",traveller)
	}
}

func TestClearedByPromise(t *testing.T) {
	var tr Traveller
	tr.balance=-1
	tr.kept.promise=Promise{TripStart:SecondsInDay,TripEnd:SecondsInDay*2,Clearance:SecondsInDay*4,Distance:10}
	tr.Promises.entries[0]=tr.kept.promise
	tr.tripHistory.entries[0] = *createFlight(1,SecondsInDay,SecondsInDay*2)
	tr.tripHistory.entries[0].distance=10
	tr.tripHistory.entries[0].et = etTravellerTripEnd
	if (!tr.Cleared(4*SecondsInDay)) {
		t.Error("Traveller not cleared with valid clearance date",tr.kept.promise)
	}
	if (tr.Cleared(3*SecondsInDay)) {
		t.Error("Traveller cleared with early clearance date",tr.kept.promise)
	}
	tr.Promises.entries[0].Clearance = SecondsInDay*3
	if (!tr.Cleared(3*SecondsInDay)) {
		t.Error("Traveller not cleared when clearance date changed",tr.kept.promise)
	}
}

func TestSubmitFlight(t *testing.T) {
	var traveller Traveller
	traveller.kept.promise.Clearance=2
	traveller.balance=0
	oneflight := *createFlight(1,1,2)
	oneflight.distance=1
	_,err:=traveller.submitFlight(&oneflight,2,10,true)
	if err != nil {
		t.Error("submitFlight failed for cleared traveller",traveller)
	}
	if (traveller.balance !=-11) {
		t.Error("submitFlight didnt update balance",traveller.balance)
	}
	traveller.EndTrip()
	_,err=traveller.submitFlight(&oneflight,1,10,true)
	if err == nil {
		t.Error("submitFlight accepted flight when grounded",traveller)
	}
	if (traveller.balance !=-11) {
		t.Error("submitFlight changed balance when grounded",traveller)
	}
}

func TestSubmitFlightNoDebit(t *testing.T) {
	var traveller Traveller
	traveller.kept.promise.Clearance=2
	traveller.balance=0
	oneflight := *createFlight(1,1,2)
	oneflight.distance=1
	_,err:=traveller.submitFlight(&oneflight,2,10,false)
	if err != nil {
		t.Error("submitFlight failed for cleared traveller",traveller)
	}
	if (traveller.balance !=0) {
		t.Error("submitFlight updated balance when debit is false",traveller.balance)
	}
	if !traveller.Cleared(0) {
		t.Error("traveller grounded after no debit submissions",traveller)
	}
}

func TestKeepNoFlights(t *testing.T) {
	var tr Traveller
	kept := tr.keep()
	if kept != false {
		t.Error("keep kept promise for an empty traveller")
	}
	var tEmpty Traveller
	if !reflect.DeepEqual(tr,tEmpty) {
		t.Error("keep changes state of an empty traveller")
	}
}

func TestKeepMatchingPromise(t *testing.T) {
	var tr Traveller
	tr.tripHistory.entries[0] = *createFlight(1,1,2)
	tr.tripHistory.entries[0].distance=55
	tr.Promises.entries[0]=Promise{TripStart:1,TripEnd:2,Clearance: epochDays(88).toEpochTime(), Travelled:55}
	if ! tr.keep()  {
		t.Error("keep didnt keep matching  promise")
	}
	if (tr.kept.promise != tr.Promises.entries[0]) {
		t.Error("keep didnt set kept  for matching made promise",tr)
	}
	if tr.MidTrip() {
		t.Error("keep failed to end trip on matching promise",tr)
	}
}

func TestKeepNonMatchingPromise(t *testing.T) {
	var tr Traveller
	tr.tripHistory.entries[0] = *createFlight(1,1,2)
	tr.tripHistory.entries[0].distance=54
	tr.Promises.entries[0]=Promise{TripStart:1,TripEnd:2,Clearance: epochDays(88).toEpochTime(), Travelled:55}
	if tr.keep()  {
		t.Error("keep kept promise that didnt match")
	}
	if (tr.kept.promise.Clearance != 0) {
		t.Error("keep changed cleared for non-matching promise",tr)
	}
	if !tr.MidTrip() {
		t.Error("keep ended trip on non-matching promise",tr)
	}

}

