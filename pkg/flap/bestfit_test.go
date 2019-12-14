package flap

import (
	"testing"
	"math"
	"reflect"
)

func TestZeroXOrigin(t *testing.T) {
	bf,err := newBestFit(SecondsInDay-1,10)
	if err != EXORIGINZERO {
		t.Error("Accepting a zero xorigin")
	}
	if bf != nil {
		t.Error("Returning a bestfit instance with a zero xorigin")
	}
}

func TestEmptyLine(t *testing.T) {
	bf,_ := newBestFit(SecondsInDay,10)
	err := bf.calculateLine()
	if err != ENOTENOUGHDATAPOINTS {
		t.Error("Calculated a line with no points")
	}
}

func TestZeroMaxpoints(t *testing.T) {
	_,err := newBestFit(SecondsInDay,1)
	if err != EMAXPOINTSBELOWTWO {
		t.Error("Allowing a maxpoints value of less than 2")
	}
}

func TestMaxpoints(t* testing.T) {
	bf,_ := newBestFit(SecondsInDay,2)
	bf.add(1)
	bf.add(2)
	bf.add(3)
	if !reflect.DeepEqual(bf.ys,[]Kilometres{2,3}) {
		t.Error("maxpoints not being enforced",bf.ys)
	}
}

func TestOnePointLine(t *testing.T) {
	bf,_ := newBestFit(SecondsInDay,10)
	bf.add(0)
	err := bf.calculateLine()
	if err != ENOTENOUGHDATAPOINTS {
		t.Error("Calculated a line with no points")
	}
}

func TestHorzontalLine(t *testing.T) {
	bf,_ := newBestFit(SecondsInDay,10)
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
	bf,_ := newBestFit(SecondsInDay,10)
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
	bf,_ := newBestFit(SecondsInDay,1000)
	for y:=Kilometres(37); y<1000;y+=5 {
		bf.add(y)
	}
	err := bf.calculateLine()
	if err != nil{
		t.Error("Can't calculate line with more than 1 data point",err)
	}
	if bf.m !=5 {
		t.Error("Calculated incorrect gradient for ascending line", bf.m)
	}
	if bf.c != 32 {
		t.Error("Calcualted incorrect constant for ascending line", bf.c)
	}
}

func TestDescending(t *testing.T) {
	bf,_ := newBestFit(SecondsInDay,1000)
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
	if bf.c != 572 {
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
	//1,2,3,4,5,6,7,8,9,10
	//510,440,410,340,310,240,210,140,110,40
	bf,_ := newBestFit(SecondsInDay,10)
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
	if to3DecimalPlaces(bf.c) != 553.333 {
		t.Error("Calculated incorrect constant for wobbly line", bf.c)
	}
}

func TestPredictFlat(t *testing.T) {
	bf,_ := newBestFit(SecondsInDay,10)
	bf.m=0
	bf.c=10
	clear,err := bf.predict(100,1)
	if err !=  nil {
		t.Error("prediced returned error for flat line",err)
	}
	if clear != 11 {
		t.Error("predicted returned wrong prediction for flat line", clear)
	}
	
	clear,err = bf.predict(200,1)
	if err !=  nil {
		t.Error("prediced returned error for flat line",err)
	}
	if clear != 21 {
		t.Error("predicted returned wrong prediction for flat line", clear)
	}
}

func TestBackfilledFlat(t *testing.T) {
	bf,_ := newBestFit(SecondsInDay,10)
	bf.m=0
	bf.c=10
	dist,err := bf.backfilled(1,2)
	if err !=  nil {
		t.Error("backfilled returned error for flat line",err)
	}
	if dist != 10 {
		t.Error("backfilled, returned wrong prediction for flat line", dist)
	}
	dist,err = bf.backfilled(17,31)
	if err !=  nil {
		t.Error("backfilled returned error for flat line",err)
	}
	if dist != 140 {
		t.Error("backfilled, returned wrong prediction for flat line", dist)
	}
}	

func TestPredictSlope(t *testing.T) {
	bf,_ := newBestFit(SecondsInDay,10)
	bf.m=-1
	bf.c=4
	clear,err := bf.predict(4,1)
	if err !=  nil {
		t.Error("prediced returned error for sloping line",err)
	}
	if clear != 3 {
		t.Error("predicted returned wrong prediction for sloping line", clear)
	}
	clear,err = bf.predict(5,1)
	if err ==  nil {
		t.Error("predicted end date for line sloping below zero",clear)
	}
}

func TestBackfilledSlope(t *testing.T) {
	bf,_ := newBestFit(SecondsInDay,10)
	bf.m=-1
	bf.c=4
	d,err := bf.backfilled(1,3)
	if err !=  nil {
		t.Error("backfilled returned error for sloping line",err)
	}
	if d != 4 {
		t.Error("backfilled returned wrong prediction for sloping line", d)
	}
	d,err = bf.backfilled(4,5)
	if err ==  nil {
		t.Error("predicted end date for line sloping below zero",d)
	}
}

func TestPredictLongSlope(t *testing.T) {
	bf,_ := newBestFit(SecondsInDay,10)
	bf.m=-0.01
	bf.c=100
	clear,err := bf.predict(1000,1)
	if err !=  nil {
		t.Error("prediced returned error for sloping line",err)
	}
	if clear != 12 {
		t.Error("predicted returned wrong prediction for sloping line", clear)
	}
	clear,err = bf.predict(1000,1000)
	if err !=  nil {
		t.Error("prediced returned error for sloping line",err)
	}
	if clear != 1012 {
		t.Error("predicted returned wrong prediction for sloping line", clear)
	}
	clear,err = bf.predict(1000,5000)
	if err !=  nil {
		t.Error("prediced returned error for sloping line",err)
	}
	if clear != 5021 {
		t.Error("predicted returned wrong prediction for sloping line", clear)
	}
}


