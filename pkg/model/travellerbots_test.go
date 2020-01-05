package model

import (
	"testing"
	"reflect"
	"github.com/richardmorrey/flap/pkg/flap"
)

func buildCountryWeights(nEntries int) *CountryWeights{
	countryWeights := NewCountryWeights()
	for j:=1; j <= nEntries; j++  {
		countryWeights.Countries= append(countryWeights.Countries,string(j+64))
		countryWeights.add(weight(j*10))
	}
	return countryWeights
}

func TestEmptySpecs(t *testing.T) {
	ts := NewTravellerBots(buildCountryWeights(1),flap.FlapParams{})
	params := ModelParams{TotalTravellers:0}
	params.BotSpecs= make([]BotSpec,0,10)
	err := ts.Build(&params)
	if (err == nil) {
		t.Error("Accepted zero bot specs",err)
	}
}

func TestOneSpec(t *testing.T) {
	ts := NewTravellerBots(buildCountryWeights(1),flap.FlapParams{})
	params := ModelParams{TotalTravellers:2}
	params.BotSpecs = make([]BotSpec,0,10)
	params.BotSpecs = append(params.BotSpecs,BotSpec{PlanProbability:0.1,Weight:12345})
	err := ts.Build(&params)
	if (err != nil) {
		t.Error("Failed b:uild from one bot spec",err)
	}
	if (len(ts.bots) != 1) {
		t.Error("Failed to Create 1 bot from 1 bot spec",err)
	}
	expected := travellerBot{countryStep:5,numInstances:2,flyProb:0.1}
	if !reflect.DeepEqual(ts.bots[0],expected) {
		t.Error("traveller bot has incorrect value",ts.bots[0],expected)
	}
}

func TestTwoSpecs(t *testing.T) {
	ts := NewTravellerBots(buildCountryWeights(1),flap.FlapParams{})
	params := ModelParams{TotalTravellers:2}
	params.BotSpecs = make([]BotSpec,0,10)
	params.BotSpecs = append(params.BotSpecs,BotSpec{PlanProbability:0.1,Weight:1})
	params.BotSpecs = append(params.BotSpecs,BotSpec{PlanProbability:0.1,Weight:1})
	err := ts.Build(&params)
	if (err != nil) {
		t.Error("Failed build from one bot spec",err)
	}
	if (len(ts.bots) != 2) {
		t.Error("Failed to create 2 bots from 2 bot specs",ts.bots)
	}
	expected := travellerBot{countryStep:10,numInstances:1,flyProb:0.1}
	if !reflect.DeepEqual(ts.bots[0],expected) {
		t.Error("traveller bot has incorrect value",ts.bots[0],expected)
	}
	if !reflect.DeepEqual(ts.bots[1],expected) {
		t.Error("traveller bot has incorrect value",ts.bots[1],expected)
	}
}

func TestThreeSpecs(t *testing.T) {
	ts := NewTravellerBots(buildCountryWeights(3),flap.FlapParams{})
	params := ModelParams{TotalTravellers:11}
	params.BotSpecs = make([]BotSpec,0,10)
	params.BotSpecs = append(params.BotSpecs,BotSpec{PlanProbability:0.1,Weight:1})
	params.BotSpecs = append(params.BotSpecs,BotSpec{PlanProbability:0.2,Weight:2})
	params.BotSpecs = append(params.BotSpecs,BotSpec{PlanProbability:0.3,Weight:8})
	err := ts.Build(&params)
	if (err != nil) {
		t.Error("Failed build from three bot specs",err)
	}
	if (len(ts.bots) != 3) {
		t.Error("Failed to create 2 bots from 2 bot specs",ts.bots)
	}
	expected := travellerBot{countryStep:60,numInstances:1,flyProb:0.1}
	if !reflect.DeepEqual(ts.bots[0],expected) {
		t.Error("traveller bot has incorrect value",ts.bots[0],expected)
	}
	expected = travellerBot{countryStep:30,numInstances:2,flyProb:0.2}
	if !reflect.DeepEqual(ts.bots[1],expected) {
		t.Error("traveller bot has incorrect value",ts.bots[1],expected)
	}
	expected = travellerBot{countryStep:7.5,numInstances:8,flyProb:0.3}
	if !reflect.DeepEqual(ts.bots[2],expected) {
		t.Error("traveller bot has incorrect value",ts.bots[2],expected)
	}
}

func TestGetBot(t *testing.T) {
	ts := NewTravellerBots(buildCountryWeights(2),flap.FlapParams{})
	params := ModelParams{TotalTravellers:10}
	params.BotSpecs = make([]BotSpec,0,10)
	params.BotSpecs = append(params.BotSpecs,BotSpec{PlanProbability:0.1,Weight:1})
	params.BotSpecs = append(params.BotSpecs,BotSpec{PlanProbability:0.2,Weight:9})
	ts.Build(&params)
	p,_ := ts.getPassport(botId{0,0})
	if p != flap.NewPassport("000000000","A")  {
		t.Error("getPassport returned wrong passport for 0,0 ",p)
	}
	p,_ = ts.getPassport(botId{1,0})
	if p != flap.NewPassport("010000000","A")  {
		t.Error("getPassport returned wrong passport for 1,0 ",p)
	}
	p,_ = ts.getPassport(botId{1,1})
	p2 := flap.NewPassport("010000001","A")
	if p != p2 {
		t.Error("getPassport returned wrong passport for 1,1 ",p,p2)
	}
	p,_ = ts.getPassport(botId{1,2})
	if p != flap.NewPassport("010000002","A")  {
		t.Error("getPassport returned wrong passport for 1,2 ",p)
	}
	p,_ = ts.getPassport(botId{1,3})
	if p != flap.NewPassport("010000003","A")  {
		t.Error("getPassport returned wrong passport for 1,3 ",p)
	}
	p,_ = ts.getPassport(botId{1,4})
	if p != flap.NewPassport("010000004","B")  {
		t.Error("getPassport returned wrong passport for 1,4 ",p)
	}
	p,_ = ts.getPassport(botId{1,5})
	if p != flap.NewPassport("010000005","B")  {
		t.Error("getPassport returned wrong passport for 1,5 ",p)
	}
	p,_ = ts.getPassport(botId{1,6})
	if p != flap.NewPassport("010000006","B")  {
		t.Error("getPassport returned wrong passport for 1,6 ",p)
	}
	p,_ = ts.getPassport(botId{1,7})
	if p != flap.NewPassport("010000007","B")  {
		t.Error("getPassport returned wrong passport for 1,7 ",p)
	}
	p,_ = ts.getPassport(botId{1,8})
	if p != flap.NewPassport("010000008","B")  {
		t.Error("getPassport returned wrong passport for 1,7 ",p)
	}
	p,_ = ts.getPassport(botId{1,9})
	if p != flap.NewPassport("010000009","B")  {
		t.Error("getPassport returned wrong passport for 1,7 ",p)
	}

}

