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
	if (traveller.Cleared(0) == CRGrounded) {
		t.Error("Traveller not cleared",traveller)
	}
}

func TestCleared1(t *testing.T) {
	var traveller Traveller
	if (traveller.Cleared(1) == CRGrounded) {
		t.Error("Traveller not cleared",traveller)
	}
}

func TestCleared2Default(t *testing.T) {
	var traveller Traveller
	traveller.Kept.Clearance=1
	traveller.Balance=-1
	populateFlights(&(traveller.tripHistory),1,1)
	traveller.EndTrip()
	if (traveller.Cleared(0) != CRGrounded) {
		t.Error("Traveller cleared",traveller)
	}
}

func TestClearedByPromise(t *testing.T) {
	var tr Traveller
	tr.Balance=-1
	tr.Kept = Promise{TripStart:SecondsInDay,TripEnd:SecondsInDay*2,Clearance:SecondsInDay*4,Distance:10}
	tr.Promises.entries[0]=tr.Kept
	tr.tripHistory.entries[0] = *createFlight(1,SecondsInDay,SecondsInDay*2)
	tr.tripHistory.entries[0].Distance=10
	tr.tripHistory.entries[0].et = etTravellerTripEnd
	if (tr.Cleared(4*SecondsInDay) != CRKeptPromise) {
		t.Error("Traveller not cleared with valid clearance date",tr.Kept)
	}
	if (tr.Cleared(3*SecondsInDay) != CRGrounded)  {
		t.Error("Traveller cleared with early clearance date",tr.Kept)
	}
	tr.Promises.entries[0].Clearance = SecondsInDay*3
	if (tr.Cleared(3*SecondsInDay) != CRKeptPromise) {
		t.Error("Traveller not cleared when clearance date changed",tr.Kept)
	}
}

func TestSubmitFlight(t *testing.T) {
	var traveller Traveller
	traveller.Kept.Clearance=2
	traveller.Balance=0
	oneflight := *createFlight(1,1,2)
	oneflight.Distance=1
	_,_,err:=traveller.submitFlight(&oneflight,2,10,true)
	if err != nil {
		t.Error("submitFlight failed for cleared traveller",traveller)
	}
	if (traveller.Balance !=-11) {
		t.Error("submitFlight didnt update balance",traveller.Balance)
	}
	traveller.EndTrip()
	_,_,err=traveller.submitFlight(&oneflight,1,10,true)
	if err == nil {
		t.Error("submitFlight accepted flight when grounded",traveller)
	}
	if (traveller.Balance !=-11) {
		t.Error("submitFlight changed balance when grounded",traveller)
	}
}

func TestSubmitFlightNoDebit(t *testing.T) {
	var traveller Traveller
	traveller.Kept.Clearance=2
	traveller.Balance=0
	oneflight := *createFlight(1,1,2)
	oneflight.Distance=1
	_,_,err:=traveller.submitFlight(&oneflight,2,10,false)
	if err != nil {
		t.Error("submitFlight failed for cleared traveller",traveller)
	}
	if (traveller.Balance !=0) {
		t.Error("submitFlight updated balance when debit is false",traveller.Balance)
	}
	if traveller.Cleared(0) == CRGrounded {
		t.Error("traveller grounded after no debit submissions",traveller)
	}
}

func TestKeepNoFlights(t *testing.T) {
	var tr Traveller
	kept := tr.keep()
	if kept != false {
		t.Error("keepkept.for an empty traveller")
	}
	var tEmpty Traveller
	if !reflect.DeepEqual(tr,tEmpty) {
		t.Error("keep changes state of an empty traveller")
	}
}

func TestKeepMatchingPromise(t *testing.T) {
	var tr Traveller
	tr.tripHistory.entries[0] = *createFlight(1,1,2)
	tr.tripHistory.entries[0].Distance=55
	tr.Promises.entries[0]=Promise{TripStart:1,TripEnd:2,Clearance: epochDays(88).toEpochTime(), Travelled:55}
	if ! tr.keep()  {
		t.Error("keep didnt keep matching  promise")
	}
	if (tr.Kept != tr.Promises.entries[0]) {
		t.Error("keep didnt set kept  for matching made promise",tr)
	}
	if tr.MidTrip() {
		t.Error("keep failed to end trip on matching promise",tr)
	}
}

func TestKeepNonMatchingPromise(t *testing.T) {
	var tr Traveller
	tr.tripHistory.entries[0] = *createFlight(1,1,2)
	tr.tripHistory.entries[0].Distance=54
	tr.Promises.entries[0]=Promise{TripStart:1,TripEnd:2,Clearance: epochDays(88).toEpochTime(), Travelled:55}
	if tr.keep()  {
		t.Error("keepkept.that didnt match")
	}
	if (tr.Kept.Clearance != 0) {
		t.Error("keep changed cleared for non-matching promise",tr)
	}
	if !tr.MidTrip() {
		t.Error("keep ended trip on non-matching promise",tr)
	}

}

func TestPassportFromString(t *testing.T) {
	p1 := NewPassport("012345678","uk")
	s := p1.ToString()
	var p2 Passport
	err := p2.FromString(s)
	if err != nil {
		t.Error("failed to parse passport from string",s)
	}
	if !reflect.DeepEqual(p1,p2) {
		t.Error("passport from string doesnt match original",p1,p2)
	}
}

func TestPassportFromStringFail(t *testing.T) {
	s := "nonsense"
	var p2 Passport
	err := p2.FromString(s)
	if err != EPASSPORTSTRINGWRONGLENGTH {
		t.Error("FromString accepted string of incorrect length",err)
	}
}
