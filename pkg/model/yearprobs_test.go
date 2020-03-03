package model

import (
	"testing"
	"time"
	"math"
	"github.com/richardmorrey/flap/pkg/flap"
)

func TestNotEnoughMonths(t *testing.T) {
	bs := BotSpec{FlyProbability:0.1,MonthWeights:[]weight{1,2,3,4,5,6,7,8,9,10,11}}
	_,err := newYearProbs(&bs)
	if err != ENOVALIDPROBABILITYINBOTSPEC {
		t.Error("newYearProbs succeeded with eleven months")
	}
}

func TestTooManyMonths(t *testing.T) {
	bs := BotSpec{FlyProbability:0.1,MonthWeights:[]weight{1,2,3,4,5,6,7,8,9,10,11,12,13}}
	_,err := newYearProbs(&bs)
	if err != ENOVALIDPROBABILITYINBOTSPEC {
		t.Error("newYearProbs succeeded with thirteen months")
	}
}

func TestMissingProb(t *testing.T) {
	bs := BotSpec{MonthWeights:[]weight{1,2,3,4,5,6,7,8,9,10,11,12}}
	_,err := newYearProbs(&bs)
	if err != ENOVALIDPROBABILITYINBOTSPEC {
		t.Error("newYearProbs succeeded with no probabilitys")
	}
}

func TestMonthlyProbs(t *testing.T) {
	
	// Create year probs 
	bs := BotSpec{FlyProbability:0.1,MonthWeights:[]weight{1,2,3,4,5,6,7,8,9,10,11,12}}
	yp,err := newYearProbs(&bs)
	if err != nil {
		t.Error("newYearProbs faied with valid monthly probs")
	}

	// Check calcualted day weights
	yearWeight := Probability(31+2*29+3*31+4*30+5*31+6*30+7*31+8*31+9*30+10*31+11*30+12*31)
	cd := time.Date(2020, time.January, 1, 1, 0, 0, 0, time.UTC);
	var total Probability
	for i:=0; i < 366; i+=1 {
		expected := (Probability(cd.Month())/yearWeight)*0.1*366 
		if yp.days[i] != expected {
			t.Error("newYearPass calculated wrong prob for day",i,expected,yp.days[i])
		}
		total += yp.days[i]
		cd = cd.Add(time.Hour*24) 
	}
	if math.Round(float64(total)*100)/100 != 36.6 {
		t.Error("newYearPass calculated incorrect total year fly probability", total)
	}
}

func TestDayProb(t *testing.T) {
	
	bs := BotSpec{FlyProbability:0.1,MonthWeights:[]weight{1,2,3,4,5,6,7,8,9,10,11,12}}
	yp,err := newYearProbs(&bs)
	if err != nil {
		t.Error("newYearProbs faied with valid monthly probs")
	}
	yearWeight := Probability(31+2*29+3*31+4*30+5*31+6*30+7*31+8*31+9*30+10*31+11*30+12*31)
	 
	p := yp.getDayProb(flap.EpochTime(time.Date(2020, time.January, 1, 1, 0, 0, 0, time.UTC).Unix()))
	if (p != (1/yearWeight)*0.1*366) {
		t.Error("Unexpected prob for 2020-01-01", p)
	}
	p = yp.getDayProb(flap.EpochTime(time.Date(2020, time.February, 29, 1, 0, 0, 0, time.UTC).Unix()))
	if (p != (2/yearWeight)*0.1*366) {
		t.Error("Unexpected prob for 2020-02-29", p)
	}
	p = yp.getDayProb(flap.EpochTime(time.Date(2020, time.March, 1, 1, 0, 0, 0, time.UTC).Unix()))
	if (p != (3/yearWeight)*0.1*366) {
		t.Error("Unexpected prob for 2020-03-01", p)
	}
	p = yp.getDayProb(flap.EpochTime(time.Date(2020, time.December, 31, 1, 0, 0, 0, time.UTC).Unix()))
	if (p != (12/yearWeight)*0.1*366) {
		t.Error("Unexpected prob for 2020-12-31", p)
	}

	p = yp.getDayProb(flap.EpochTime(time.Date(2021, time.February, 28, 1, 0, 0, 0, time.UTC).Unix()))
	if (p != (2/yearWeight)*0.1*366) {
		t.Error("Unexpected prob for 2021-02-28", p)
	}
	p = yp.getDayProb(flap.EpochTime(time.Date(2021, time.March, 1, 1, 0, 0, 0, time.UTC).Unix()))
	if (p != (3/yearWeight)*0.1*366) {
		t.Error("Unexpected prob for 2021-03-01", p)
	}
	p = yp.getDayProb(flap.EpochTime(time.Date(2021, time.December, 31, 1, 0, 0, 0, time.UTC).Unix()))
	if (p != (12/yearWeight)*0.1*366) {
		t.Error("Unexpected prob for 2021-12-31", p)
	}
}

