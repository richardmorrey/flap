package flap

import (
	"testing"
	"math"
	"reflect"
)

func TestEmptyLine(t *testing.T) {
	bf,_ := newBestFit(10)
	err := bf.calculateLine(1)
	if err != ENOTENOUGHDATAPOINTS {
		t.Error("Calculated a line with no points")
	}
}

func TestZeroMaxpoints(t *testing.T) {
	_,err := newBestFit(1)
	if err != EMAXPOINTSBELOWTWO {
		t.Error("Allowing a maxpoints value of less than 2")
	}
}

func TestMaxpoints(t* testing.T) {
	bf,_ := newBestFit(2)
	bf.add(1,1)
	bf.add(2,2)
	bf.add(3,3)

	if !reflect.DeepEqual(bf.ys,[]Kilometres{2,3}) {
		t.Error("maxpoints not being enforced",bf.ys)
	}
}

func TestOnePointLine(t *testing.T) {
	bf,_ := newBestFit(10)
	bf.add(1,0)
	err := bf.calculateLine(1)
	if err != ENOTENOUGHDATAPOINTS {
		t.Error("Calculated a line with no points")
	}
}

func TestHorzontalLine(t *testing.T) {
	bf,_ := newBestFit(10)
	bf.add(1,10)
	bf.add(2,10)
	if bf.m !=0 {
		t.Error("Calculated non-zero gradient for horizontal line", bf.m)
	}
	if bf.c != 10 {
		t.Error("Failed to calculate c correctly for horizontal line", bf.c)
	}
}

func TestLongHorizontal(t *testing.T) {
	bf,_ := newBestFit(10)
	for x:=1; x<1000;x++ {
		bf.add(epochDays(x),999)
	}
	if bf.m !=0 {
		t.Error("Calculated non-zero gradient for horizontal line", bf.m)
	}
	if bf.c != 999 {
		t.Error("Failed to calculate c correctly for horizontla line", bf.c)
	}
}

func TestAscending(t *testing.T) {
	bf,_ := newBestFit(1000)
	x := epochDays(1)
	for y:=Kilometres(37); y<1000;y+=5 {
		bf.add(x,y)
		x++
	}
	if bf.m !=5 {
		t.Error("Calculated incorrect gradient for ascending line", bf.m)
	}
	if bf.c != 32 {
		t.Error("Calcualted incorrect constant for ascending line", bf.c)
	}
}

func TestDescending(t *testing.T) {
	bf,_ := newBestFit(1000)
	x := epochDays(1)
	for y:=Kilometres(567); y>0;y-=5 {
		bf.add(x,y)
		x++
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
	bf,_ := newBestFit(10)
	x := epochDays(1)
	for y:=Kilometres(500); y>0;y-=50 {
		if y % 100 ==0 {
			bf.add(x,y+10)
		} else {
			bf.add(x,y-10)
		}
		x++
	}

	// Expected results from http://www.endmemo.com/statistics/lr.php
	if to3DecimalPlaces(bf.m) !=-50.606 {
		t.Error("Calculated incorrect gradient for wobbly line", bf.m)
	}
	if to3DecimalPlaces(bf.c) != 553.333 {
		t.Error("Calculated incorrect constant for wobbly line", bf.c)
	}
}

func TestPredictFlat(t *testing.T) {
	bf,_ := newBestFit(10)
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
	bf,_ := newBestFit(10)
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
	dist,err = bf.backfilled(17,17)
	if err !=  nil {
		t.Error("backfilled returned error for flat line",err)
	}
	if dist != 0 {
		t.Error("backfilled, returned wrong prediction for same start/end", dist)
	}

}	

func TestPredictSlope(t *testing.T) {
	bf,_ := newBestFit(10)
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
	bf,_ := newBestFit(10)
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
	bf,_ := newBestFit(10)
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

func TestVersion(t *testing.T) {
	bf,_ := newBestFit(1000)
	x := epochDays(1)
	for y:=Kilometres(100); y>80;y-=5 {
		bf.add(x,y)
		x++
	}
	pv := bf.version()
	for y:=Kilometres(80); y>50;y-=5 {
		bf.add(x,y)
		x++
	}
	if pv != bf.version() {
		t.Error("version changed when m and c should have stayed the same")
	}
	for y:=Kilometres(50); y>0;y-=10 {
		bf.add(x,y)
		x++
	}
	if bf.version() == pv {
		t.Error("version didnt change when m and c should have changed",bf.version())
	}
}
