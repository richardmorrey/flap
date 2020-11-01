package model

import (
	"testing"
	"github.com/richardmorrey/flap/pkg/flap"
	"github.com/richardmorrey/flap/pkg/db"
	"os"
)

var JPTESTFOLDER="jptest"

func setupJP(t *testing.T) *db.LevelDB{
	if err := os.Mkdir(JPTESTFOLDER, 0700); err != nil {
		t.Error("Failed to create test dir", err)
	}
	return db.NewLevelDB(JPTESTFOLDER)
} 

func teardownJP(db *db.LevelDB) {
	db.Release()
	os.RemoveAll(JPTESTFOLDER)
}


func TestNewJourneyPlanner(t *testing.T) {

	db := setupJP(t)
	defer teardownJP(db)
	jp,err := NewJourneyPlanner(db)
	if (err != nil) {
		t.Error("Failed to create new journey planner")
	}
	if (jp == nil) {
		t.Error("Failed to return journey planner")
	}
}

func TestAddJourney(t *testing.T) {

	db := setupJP(t)
	defer teardownJP(db)
	jp,err := NewJourneyPlanner(db)
	if (err != nil) {
		t.Error("Failed to create new journey planner")
	}

	f,_ := flap.NewFlight(flap.Airport{},flap.SecondsInDay,flap.Airport{},flap.SecondsInDay+1)
	pp := flap.NewPassport("987654321","uk")

	j := journey{jt:jtInbound,flight:*f}
	err = jp.addJourney(pp,j)
	if  err != nil {
		t.Error("Failed to add journey")
	}

	it,err := jp.NewIterator(flap.EpochTime(flap.SecondsInDay))
	if err != nil {
		t.Error("Failed to create iterator")
	}
	for it.Next() {

		pf := it.Value()
		if len(pf.journies) != 1 {
			t.Error("No planned jouurnies")
		}
		if pf.journies[0] !=  j {
			t.Error("Retrieved journey doesnt match what was stored", pf.journies[0],j)
		}
		pp2,err := it.Passport()
		if err != nil {
			t.Error("Could retrieve passport")
		}
		if pp2 != pp {
			t.Error("Retrieved passport doesnt match what was stored", pp2,pp)
		}
	}

}

/*
func TestplanTrip(t *testing.T) {
	db := setupJP(t)
	defer teardownJP()
	jp,err := NewJourneyPlanner(jp)
	if (err != nil) {
		t.Error("Failed to create new journey planner")
	}

	jp,_ := NewJourneyPlanner(3)
	bot:= botId{1,2}
	from:= flap.NewICAOCode("A")
	to:=flap.NewICAOCode("B")
	err:=jp.planTrip(from,to,10,bot,0)
	if (err != nil)  {
		t.Error("Failed to plan trip",err)
	}
}
*/

