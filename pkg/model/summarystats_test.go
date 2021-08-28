package model

import (
	"testing"
	"time"
	"github.com/richardmorrey/flap/pkg/flap"
)


func TestMeanNoData(t *testing.T) {
	var ss summaryStats
	mp := ModelParams{BotSpecs:[]BotSpec{{FlyProbability: 1,Weight: 100,MonthWeights: []weight{10,10,10,10,10,10,10,10,10,10,10,10}}}}
	_,err := ss.calculateMeanDaily(&mp,flap.EpochTime(time.Date(2020, time.January, 1, 1, 0, 0, 0, time.UTC).Unix()))
	if err != ESSNODATA {
		t.Error("Incorrect error code on no data",err)
	}
}

func TestMeanZero(t *testing.T) {
	var ss summaryStats
	ss.update(summaryStatsRow{Travelled:0},1,nil)
	mp := ModelParams{ReportDayDelta:1,BotSpecs:[]BotSpec{{FlyProbability: 1,Weight: 100,MonthWeights: []weight{10,10,10,10,10,10,10,10,10,10,10,10}}}}
	m,err := ss.calculateMeanDaily(&mp,flap.EpochTime(time.Date(2020, time.January, 1, 1, 0, 0, 0, time.UTC).Unix()))
	if err != nil {
		t.Error("Error calculating mean daily with data for a full RDD",err)
	}
	if m != 0 {
		t.Error("incorrect mean daily",m)
	}
}

func TestMeanOne(t *testing.T) {
	var ss summaryStats
	ss.update(summaryStatsRow{Travelled:10},1,nil)
	mp := ModelParams{ReportDayDelta:1,BotSpecs:[]BotSpec{{FlyProbability: 1,Weight: 100,MonthWeights: []weight{10,10,10,10,10,10,10,10,10,10,10,10}}}}
	m,err := ss.calculateMeanDaily(&mp,flap.EpochTime(time.Date(2020, time.January, 1, 1, 0, 0, 0, time.UTC).Unix()))
	if err != nil {
		t.Error("Error calculating mean daily with data for a full RDD",err)
	}
	if m != 10 {
		t.Error("incorrect mean daily",m)
	}
}

func TestMeanDouble(t *testing.T) {
	var ss summaryStats
	ss.update(summaryStatsRow{Travelled:10},1,nil)
	mp := ModelParams{ReportDayDelta:1,BotSpecs:[]BotSpec{{FlyProbability: 1,Weight: 100,MonthWeights: []weight{24,24,12,24,36,24,24,24,24,24,24,24}}}}
	m,err := ss.calculateMeanDaily(&mp,flap.EpochTime(time.Date(2020, time.March, 1, 1, 0, 0, 0, time.UTC).Unix()))
	if err != nil {
		t.Error("Error calculating mean daily with data for a year",err)
	}
	if m != 20 {
		t.Error("incorrect mean daily",m)
	}
}

func TestMeanDoubleEnd(t *testing.T) {
	var ss summaryStats
	ss.update(summaryStatsRow{Travelled:10},1,nil)
	mp := ModelParams{ReportDayDelta:1,BotSpecs:[]BotSpec{{FlyProbability: 1,Weight: 100,MonthWeights: []weight{24,24,12,24,36,24,24,24,24,24,24,24}}}}
	m,err := ss.calculateMeanDaily(&mp,flap.EpochTime(time.Date(2020, time.March, 31, 1, 0, 0, 0, time.UTC).Unix()))
	if err != nil {
		t.Error("Error calculating mean daily with data for a year",err)
	}
	if m != 20 {
		t.Error("incorrect mean daily",m)
	}
}

func TestMeanDoubleBefore(t *testing.T) {
	var ss summaryStats
	ss.update(summaryStatsRow{Travelled:10},1,nil)
	mp := ModelParams{ReportDayDelta:1,BotSpecs:[]BotSpec{{FlyProbability: 1,Weight: 100,MonthWeights: []weight{24,24,12,24,36,24,24,24,24,24,24,24}}}}
	m,err := ss.calculateMeanDaily(&mp,flap.EpochTime(time.Date(2020, time.February, 29, 1, 0, 0, 0, time.UTC).Unix()))
	if err != nil {
		t.Error("Error calculating mean daily with data for a year",err)
	}
	if m != 10 {
		t.Error("incorrect mean daily",m)
	}
}

func TestMeanDoubleAfter(t *testing.T) {
	var ss summaryStats
	ss.update(summaryStatsRow{Travelled:10},1,nil)
	mp := ModelParams{ReportDayDelta:1,BotSpecs:[]BotSpec{{FlyProbability: 1,Weight: 100,MonthWeights: []weight{24,24,12,24,36,24,24,24,24,24,24,24}}}}
	m,err := ss.calculateMeanDaily(&mp,flap.EpochTime(time.Date(2020, time.April, 1, 1, 0, 0, 0, time.UTC).Unix()))
	if err != nil {
		t.Error("Error calculating mean daily with data for a year",err)
	}
	if m != 10 {
		t.Error("incorrect mean daily",m)
	}
}

func TestMeanBigRDD(t *testing.T) {
	var ss summaryStats
	for i :=0; i  < 10; i++ {
		ss.update(summaryStatsRow{Travelled:40},10,nil)
	}
	mp := ModelParams{ReportDayDelta:10,BotSpecs:[]BotSpec{{FlyProbability: 1,Weight: 100,MonthWeights: []weight{24,24,12,24,36,24,24,24,24,24,24,24}}}}
	m,err := ss.calculateMeanDaily(&mp,flap.EpochTime(time.Date(2020, time.March, 1, 1, 0, 0, 0, time.UTC).Unix()))
	if err != nil {
		t.Error("Error calculating mean daily",err)
	}
	if m != 800 {
		t.Error("incorrect mean daily",m)
	}
}

func TestMeanIncompleteRDD(t *testing.T) {
	var ss summaryStats
	for i :=0 ; i < 10; i ++ {
		ss.update(summaryStatsRow{Travelled:10},10,nil)
	}
	for i :=0; i  < 9; i++ {
		ss.update(summaryStatsRow{Travelled:40},10,nil)
	}
	mp := ModelParams{ReportDayDelta:10,BotSpecs:[]BotSpec{{FlyProbability: 1,Weight: 100,MonthWeights: []weight{10,10,10,10,10,10,10,10,10,10,10,10}}}}
	m,err := ss.calculateMeanDaily(&mp,flap.EpochTime(time.Date(2020, time.January, 1, 1, 0, 0, 0, time.UTC).Unix()))
	if err != nil {
		t.Error("Error calculating mean daily",err)
	}
	if m != 100 {
		t.Error("incorrect mean daily",m)
	}
}
