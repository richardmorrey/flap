package model

import (
	"testing"
)


func TestMeanNoData(t *testing.T) {
	var ss summaryStats
	_,err := ss.calculateMeanDaily(1)
	if err != ESSNODATA {
		t.Error("Incorrect error code on no data",err)
	}
}

func TestMeanNotEnoughData(t *testing.T) {
	var ss summaryStats
	for i :=0 ; i < 365; i ++ {
		ss.update(summaryStatsRow{Travellers:0,Travelled:0},1,nil)
	}
	_,err := ss.calculateMeanDaily(1)
	if err != ESSNOTENOUGHDATA {
		t.Error("Incorrect error code on no data",err)
	}
}

func TestMeanZero(t *testing.T) {
	var ss summaryStats
	for i :=0 ; i < 366; i ++ {
		ss.update(summaryStatsRow{Travellers:0,Travelled:0},1,nil)
	}
	m,err := ss.calculateMeanDaily(1)
	if err != nil {
		t.Error("Error calculating mean daily with data for a year",err)
	}
	if m != 0 {
		t.Error("incorrect mean daily",m)
	}
}

func TestMeanOne(t *testing.T) {
	var ss summaryStats
	for i :=0 ; i < 366; i ++ {
		ss.update(summaryStatsRow{Travellers:10,Travelled:10},1,nil)
	}
	m,err := ss.calculateMeanDaily(1)
	if err != nil {
		t.Error("Error calculating mean daily with data for a year",err)
	}
	if m != 1 {
		t.Error("incorrect mean daily",m)
	}
}

func TestMeanDouble(t *testing.T) {
	var ss summaryStats
	for i :=0 ; i < 365; i ++ {
		ss.update(summaryStatsRow{Travellers:10,Travelled:10},1,nil)
	}
	ss.update(summaryStatsRow{Travellers:20,Travelled:40},1,nil)
	m,err := ss.calculateMeanDaily(1)
	if err != nil {
		t.Error("Error calculating mean daily with data for a year",err)
	}
	if m != 2 {
		t.Error("incorrect mean daily",m)
	}
}

func TestMeanBigRDD(t *testing.T) {
	var ss summaryStats
	for i :=0 ; i < 370; i ++ {
		ss.update(summaryStatsRow{Travellers:10,Travelled:10},10,nil)
	}
	for i :=0; i  < 10; i++ {
		ss.update(summaryStatsRow{Travellers:20,Travelled:40},10,nil)
	}
	m,err := ss.calculateMeanDaily(10)
	if err != nil {
		t.Error("Error calculating mean daily with data for a year",err)
	}
	if m != 2 {
		t.Error("incorrect mean daily",m)
	}
}

func TestMeanIncompleteRDD(t *testing.T) {
	var ss summaryStats
	for i :=0 ; i < 370; i ++ {
		ss.update(summaryStatsRow{Travellers:10,Travelled:10},10,nil)
	}
	for i :=0; i  < 9; i++ {
		ss.update(summaryStatsRow{Travellers:20,Travelled:40},10,nil)
	}
	m,err := ss.calculateMeanDaily(10)
	if err != nil {
		t.Error("Error calculating mean daily with data for a year",err)
	}
	if m != 1 {
		t.Error("incorrect mean daily",m)
	}
}
