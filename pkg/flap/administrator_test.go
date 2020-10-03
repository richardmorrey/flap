package flap

import (
	"testing"
	"os"
	"reflect"
	"github.com/richardmorrey/flap/pkg/db"
)

var ADMINTESTFOLDER="admintest"

func setupAdmin(t *testing.T) *db.LevelDB{
	if err := os.Mkdir(ADMINTESTFOLDER, 0700); err != nil {
		t.Error("Failed to create test dir", err)
	}
	return db.NewLevelDB(ADMINTESTFOLDER)
} 

func teardownAdmin(db *db.LevelDB) {
	db.Release()
	os.RemoveAll(ADMINTESTFOLDER)
}

func TestSaveLoadBackfill(t *testing.T) {

	db:=setupAdmin(t)
	defer teardownAdmin(db)
	
	admin := newAdministrator(db)
	if admin == nil {
		t.Error("Failed to create administrator")
	}

	admin.bs.totalGrounded=10
	err := admin.Save()
	if err != nil {
		t.Error("Failed to save modified backfill state",err)
	}
	
	admin2 := newAdministrator(db)
	if admin2 == nil {
		t.Error("Failed to create administrator")
	}

        if admin2.bs.totalGrounded != 10 {
		t.Error("Failed to load saved backfill state", admin2.bs)
	}

}

func TestSaveLoadParams(t *testing.T) {

	db:=setupAdmin(t)
	defer teardownAdmin(db)
	
	admin := newAdministrator(db)
	if admin == nil {
		t.Error("Failed to create administrator")
	}

	paramsIn := FlapParams{DailyTotal:100, MinGrounded:1,FlightInterval:1,FlightsInTrip:50,TripLength:365,
		Promises:PromisesConfig{Algo:paLinearBestFit,MaxPoints:10,MaxDays:100}}
	admin.SetParams(paramsIn)
	err := admin.Save()
	if err != nil {
		t.Error("Failed to save modified params",err)
	}
	
	admin2 := newAdministrator(db)
	if admin2 == nil {
		t.Error("Failed to create administrator")
	}

        if admin2.params != paramsIn {
		t.Error("Failed to load saved params", admin2.params)
	}

}

func TestSaveLoadCorrections(t *testing.T) {

	db:=setupAdmin(t)
	defer teardownAdmin(db)
	
	admin := newAdministrator(db)
	if admin == nil {
		t.Error("Failed to create administrator")
	}

	admin.pc.change(10,100)
	err := admin.Save()
	if err != nil {
		t.Error("Failed to save modified params",err)
	}
	
	admin2 := newAdministrator(db)
	if admin2 == nil {
		t.Error("Failed to create administrator")
	}

        if !reflect.DeepEqual(admin2.pc.state,admin.pc.state) {
		t.Error("Failed to load saved params", admin2.pc.state)
	}

}

func TestSaveLoadPredictor(t *testing.T) {

	db:=setupAdmin(t)
	defer teardownAdmin(db)
	
	admin := newAdministrator(db)
	if admin == nil {
		t.Error("Failed to create administrator")
	}

	admin.SetParams(FlapParams{DailyTotal:100, MinGrounded:1,FlightInterval:1,FlightsInTrip:50,TripLength:365,
		Promises:PromisesConfig{Algo:paLinearBestFit,MaxPoints:10,MaxDays:100}})
	if admin.predictor == nil {
		t.Error("Administrator didn't create predictor")
	}
	admin.predictor.add(SecondsInDay,1000)
	admin.predictor.add(2*SecondsInDay,900)
	admin.predictor.add(3*SecondsInDay, 800)
	err := admin.Save()
	if err != nil {
		t.Error("Failed to save modified params",err)
	}
	
	admin2 := newAdministrator(db)
	if admin2 == nil {
		t.Error("Failed to create administrator")
	}

	ys,mcs,_ := admin.predictor.state()
	ys2,mcs2,_ := admin2.predictor.state()
        if !reflect.DeepEqual(ys2,ys) {
		t.Error("Failed to load saved predictor points", ys,ys2)
	}
        if !reflect.DeepEqual(mcs2,mcs) {
		t.Error("Failed to load saved predictor fit", mcs,mcs2)
	}
}

func TestLinearBestFitPredictor(t *testing.T) {

	db:=setupAdmin(t)
	defer teardownAdmin(db)
	
	admin := newAdministrator(db)
	if admin == nil {
		t.Error("Failed to create administrator")
	}

	admin.SetParams(FlapParams{Promises:PromisesConfig{Algo:paLinearBestFit,MaxPoints:10,MaxDays:100}})
	if admin.predictor == nil {
		t.Error("Administrator didn't create predictor")
	}

	switch v := admin.predictor.(type) {
		case *bestFit:
		break
		default:
			t.Error("Failed to create Linear Best Fit predictor",v)
		break
	}
}

func TestPolyBestFitPredictor(t *testing.T) {

	db:=setupAdmin(t)
	defer teardownAdmin(db)
	
	admin := newAdministrator(db)
	if admin == nil {
		t.Error("Failed to create administrator")
	}

	admin.SetParams(FlapParams{Promises:PromisesConfig{Algo:paPolyBestFit,MaxPoints:10,MaxDays:100}})
	if admin.predictor == nil {
		t.Error("Administrator didn't create predictor")
	}

	switch v := admin.predictor.(type) {
		case *polyBestFit:
		break
		default:
			t.Error("Failed to create Poly Best Fit predictor",v)
		break
	}
}

