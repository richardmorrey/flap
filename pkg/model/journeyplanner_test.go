package model

import (
	"testing"
	"github.com/richardmorrey/flap/pkg/flap"
)

func TestNewJourneyPlannerZero(t *testing.T) {
	_,err := NewJourneyPlanner(0)
	if (err == nil) {
		t.Error("NewJourneyPlanner accepted zero trip lengths")
	}
}

func TestNewJourneyPlanner(t *testing.T) {
	jp,err := NewJourneyPlanner(10)
	if (err != nil) {
		t.Error("NewJourneyPlanner failed with 1 trip length")
	}
	if len(jp.days) != 12 {
		t.Error("NewJourneyPlanner failed to create the correct number of days",len(jp.days))
	}
}

func TestAddJourneyHigh(t *testing.T) {
	jp,_ := NewJourneyPlanner(3)
	err := jp.addJourney(journey{},flap.Days(len(jp.days)))
	if  err != EDAYTOOFARAHEAD {
		t.Error("addJourney accepts day greater than maximum")
	}
}

func TestAddJourney(t *testing.T) {
	jp,_ := NewJourneyPlanner(3)
	j := journey{jt:jtInbound}
	err := jp.addJourney(j,2)
	if  err != nil {
		t.Error("Failed to add journey")
	}
	if len(jp.days[2]) != 1 {
		t.Error("Failed to add journey")
	}
	if jp.days[2][0] != j {
		t.Error("Failed to add journey with correct value", jp.days[2][0])
	}
}

func TestplanTrip(t *testing.T) {
	jp,_ := NewJourneyPlanner(3)
	bot:= botId{1,2}
	from:= flap.NewICAOCode("A")
	to:=flap.NewICAOCode("B")
	err:=jp.planTrip(from,to,10,bot,0)
	if (err != nil)  {
		t.Error("Failed to plan trip",err)
	}
	j := journey{jt:jtOutbound,flight:journeyFlight{from,to},length:10,bot:bot}
	if jp.days[0][0] != j {
		t.Error("Failed to set outbound journey correctly",jp.days[0][0])
	}
}
