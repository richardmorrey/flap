package model

import (
	"testing"
	"github.com/richardmorrey/flap/pkg/flap"
)

func TestNewJourneyPlannerNil(t *testing.T) {
	_,err := NewJourneyPlanner(nil)
	if (err == nil) {
		t.Error("NewJourneyPlanner accepted nil trip lengths")
	}
}

func TestNewJourneyPlannerEmpty(t *testing.T) {
	var empty []flap.Days
	_,err := NewJourneyPlanner(empty)
	if (err == nil) {
		t.Error("NewJourneyPlanner accepted nil trip lengths")
	}
}

func TestNewJourneyPlannerOneLength(t *testing.T) {
	simple := []flap.Days{5}
	jp,err := NewJourneyPlanner(simple)
	if (err != nil) {
		t.Error("NewJourneyPlanner failed with 1 trip length")
	}
	if len(jp.days) != 7 {
		t.Error("NewJourneyPlanner failed to create the correct number of days",len(jp.days))
	}
}

func TestNewJourneyPlannerSimple(t *testing.T) {
	simple := []flap.Days{1,2,2,11,5,10}
	jp,err := NewJourneyPlanner(simple)
	if (err != nil) {
		t.Error("NewJourneyPlanner failed with 1 trip length")
	}
	if len(jp.days) != 13 {
		t.Error("NewJourneyPlanner failed to set correct maxTripLength",len(jp.days))
	}
}

func TestAddJourneyHigh(t *testing.T) {
	jp,_ := NewJourneyPlanner([]flap.Days{1,2,3})
	err := jp.addJourney(journey{},flap.Days(len(jp.days)))
	if  err != EDAYTOOFARAHEAD {
		t.Error("addJourney accepts day greater than maximum")
	}
}

func TestAddJourney(t *testing.T) {
	jp,_ := NewJourneyPlanner([]flap.Days{1,2,3})
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
	jp,_ := NewJourneyPlanner([]flap.Days{1,2,3})
	bot:= botId{1,2}
	from:= flap.NewICAOCode("A")
	to:=flap.NewICAOCode("B")
	err:=jp.planTrip(from,to,bot) 
	if (err != nil)  {
		t.Error("Failed to plan trip",err)
	}
	j := journey{jt:jtOutbound,flight:journeyFlight{from,to},bot:bot}
	if jp.days[0][0] != j {
		t.Error("Failed to set outbound journey correctly",jp.days[0][0])
	}
}
