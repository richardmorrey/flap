package flap

import (
	"testing"
	"math"
)

func TestEmptyLine(t *testing.T) {
	bf := newBestFit(0)
	err := bf.calculateLine()
	if err != ENOTENOUGHDATAPOINTS {
		t.Error("Calculated a line with no points")
	}
}

func TestOnePointLine(t *testing.T) {
	bf := newBestFit(0)
	bf.add(0)
	err := bf.calculateLine()
	if err != ENOTENOUGHDATAPOINTS {
		t.Error("Calculated a line with no points")
	}
}

func TestHorzontalLine(t *testing.T) {
	bf := newBestFit(0)
	bf.add(10)
	bf.add(10)
	err := bf.calculateLine()
	if err != nil{
		t.Error("Can't calculate line with more than 1 data point",err)
	}
	if bf.m !=0 {
		t.Error("Calculated non-zero gradient for horizontal line", bf.m)
	}
	if bf.c != 10 {
		t.Error("Failed to calculate c correctly for horizontal line", bf.c)
	}
}

func TestLongHorizontal(t *testing.T) {
	bf := newBestFit(0)
	for x:=1; x<1000;x++ {
		bf.add(999)
	}
	err := bf.calculateLine()
	if err != nil{
		t.Error("Can't calculate line with more than 1 data point",err)
	}
	if bf.m !=0 {
		t.Error("Calculated non-zero gradient for horizontal line", bf.m)
	}
	if bf.c != 999 {
		t.Error("Failed to calculate c correctly for horizontla line", bf.c)
	}
}

func TestAscending(t *testing.T) {
	bf := newBestFit(0)
	for x:=Kilometres(37); x<1000;x+=5 {
		bf.add(x)
	}
	err := bf.calculateLine()
	if err != nil{
		t.Error("Can't calculate line with more than 1 data point",err)
	}
	if bf.m !=5 {
		t.Error("Calculated incorrect gradient for ascending line", bf.m)
	}
	if bf.c != 37 {
		t.Error("Calcualted incorrect constant for ascending line", bf.c)
	}
}

func TestDescending(t *testing.T) {
	bf := newBestFit(0)
	for x:=Kilometres(567); x>0;x-=5 {
		bf.add(x)
	}
	err := bf.calculateLine()
	if err != nil{
		t.Error("Can't calculate line with more than 1 data point",err)
	}
	if bf.m !=-5 {
		t.Error("Calculated incorrect gradient for descending line", bf.m)
	}
	if bf.c != 567 {
		t.Error("Calculated incorrect constant for descending line", bf.c)
	}
}

func to3DecimalPlaces(x float64) float64 {
	x *= 1000
	t := math.Trunc(x)
	if math.Abs(x-t) >= 0.5 {
		t += math.Copysign(1, x)
	}
	return t/1000
}

func TestWobbly(t *testing.T) {

	// Create line with following x and y
	//0,1,2,3,4,5,6,7,8,9
	//510,440,410,340,310,240,210,140,110,40,10
	bf := newBestFit(0)
	for x:=Kilometres(500); x>0;x-=50 {
		if x % 100 ==0 {
			bf.add(x+10)
		} else {
			bf.add(x-10)
		}
	}
	err := bf.calculateLine()

	// Expected results from http://www.endmemo.com/statistics/lr.php
	if err != nil{
		t.Error("Can't calculate line with more than 1 data point",err)
	}
	if to3DecimalPlaces(bf.m) !=-50.606 {
		t.Error("Calculated incorrect gradient for wobbly line", bf.m)
	}
	if to3DecimalPlaces(bf.c) != 502.727 {
		t.Error("Calculated incorrect constant for wobbly line", bf.c)
	}
}

