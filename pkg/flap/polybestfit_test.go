package flap

import (
	"testing"
	"reflect"
	"bytes"
)

func TestPolyZeroMaxpoints(t *testing.T) {
	_,err := newPolyBestFit(PromisesConfig{MaxPoints:1})
	if err != EMAXPOINTSBELOWTWO {
		t.Error("Allowing a maxpoints value of less than 2")
	}
}

func TestPolyAddOne(t *testing.T) {
	p,err := newPolyBestFit(PromisesConfig{MaxPoints:10,Degree:1})
	if err != nil {
		t.Error("newPolyBestFit returned error",err)
	}
	p.add(1,10)
	if err != nil {
		t.Error("Faiing to add the first point,err")
	}
	if len(p.consts) != 0 {
		t.Error("Calculating consts for a single point line",p.consts)
	}
}

func TestPolyHorzontalLine(t *testing.T) {
	p,_ := newPolyBestFit(PromisesConfig{MaxPoints:10,Degree:1})
	p.add(1,10)
	p.add(2,10)
	if len(p.consts) != 2 {
		t.Error("Not enough points for a 1 degree regression",p.consts)
	}
	if to3DecimalPlaces(p.consts[1]) !=0 {
		t.Error("Calculated non-zero gradient for horizontal line", p.consts[0])
	}
	if to3DecimalPlaces(p.consts[0]) != 10 {
		t.Error("Failed to calculate c correctly for horizontal line", p.consts[1])
	}
}

func TestPolyDegree2(t *testing.T) {
	ys := []Kilometres{1, 6, 17, 34, 57, 86, 121, 162, 209, 262, 321}
	p,_ := newPolyBestFit(PromisesConfig{MaxPoints:11,Degree:2})
	for _,y := range ys {
		p.add(10,y)
	}
	if len(p.consts) != 3 {
		t.Error("Not enough points for a 3 degree regression",p.consts)
	}
	if to3DecimalPlaces(p.consts[0]) !=1 {
		t.Error("Calculated incorrect first const for polynomial line", p.consts[0])
	}
	if to3DecimalPlaces(p.consts[1]) !=2 {
		t.Error("Calculated incorrect second const for polynomial line", p.consts[1])
	}
	if to3DecimalPlaces(p.consts[2]) !=3 {
		t.Error("Calculated incorrect third const for polynomial line", p.consts[2])
	}
}

func TestPolyPredictY(t *testing.T) {
	ys := []Kilometres{1, 6, 17, 34, 57, 86, 121, 162, 209, 262, 321}
	p,_ := newPolyBestFit(PromisesConfig{MaxPoints:11,Degree:2})
	for _,y := range ys {
		p.add(10,y)
	}
	y,err := p.predictY(8)
	if err != nil {
		t.Error("predictY returned unexpected error")
	}
	if to3DecimalPlaces(y) != 209 {
		t.Error("predictY failed to return correct prediction for day 8", y)
	}
}

func TestPolyPredictYNoPoints(t *testing.T) {
	p,_ := newPolyBestFit(PromisesConfig{MaxPoints:10,Degree:1})
	_,err := p.predictY(8)
	if err != ENOVALIDPREDICTION {
		t.Error("predictY didnt return error with no data points",err)
	}
}

func TestPolyPredictYOnePoint(t *testing.T) {
	p,_ := newPolyBestFit(PromisesConfig{MaxPoints:10,Degree:1})
	p.add(1,13)
	_,err := p.predictY(500)
	if err != ENOVALIDPREDICTION{
		t.Error("predictY didnt return error with too few data points",err)
	}
}

func TestPolyLongHorizontal(t *testing.T) {
	p,_ := newPolyBestFit(PromisesConfig{MaxPoints:10,Degree:1})
	for x:=1; x<1000;x++ {
		p.add(epochDays(x),999)
	}
	if to3DecimalPlaces(p.consts[1]) !=0 {
		t.Error("Calculated non-zero gradient for horizontal line", p.consts[1])
	}
	if to3DecimalPlaces(p.consts[0]) != 999 {
		t.Error("Failed to calculate c correctly for horizontla line", p.consts[0])
	}
}

func TestPolyAscending(t *testing.T) {
	p,_ := newPolyBestFit(PromisesConfig{MaxPoints:1000,Degree:1})
	x := epochDays(1)
	for y:=Kilometres(37); y<1000;y+=5 {
		p.add(x,y)
		x++
	}
	if to3DecimalPlaces(p.consts[1]) !=5 {
		t.Error("Calculated incorrect gradient for ascending line", p.consts[1])
	}
	if to3DecimalPlaces(p.consts[0]) != 32 {
		t.Error("Calcualted incorrect constant for ascending line", p.consts[0])
	}
}

func TestPolyDescending(t *testing.T) {
	p,_ := newPolyBestFit(PromisesConfig{MaxPoints:1000,Degree:1})
	x := epochDays(1)
	for y:=Kilometres(567); y>0;y-=5 {
		p.add(x,y)
		x++
	}
	if to3DecimalPlaces(p.consts[1]) !=-5 {
		t.Error("Calculated incorrect gradient for descending line", p.consts[1])
	}
}

func TestPolyVersion(t *testing.T) {
	p,_ := newPolyBestFit(PromisesConfig{MaxPoints:1000,Degree:1})
	x := epochDays(1)
	for y:=Kilometres(100); y>80;y-=5 {
		p.add(x,y)
		x++
	}
	pv := p.version()
	for y:=Kilometres(80); y>50;y-=5 {
		p.add(x,y)
		x++
	}
	if pv != p.version() {
		t.Error("version changed when m and c should have stayed the same")
	}
	for y:=Kilometres(50); y>0;y-=10 {
		p.add(x,y)
		x++
	}
	if p.version() == pv {
		t.Error("version didnt change when m and c should have changed",p.version())
	}
}

func TestPolyPredictNoPoints(t *testing.T) {
	p,_ := newPolyBestFit(PromisesConfig{MaxPoints:10,Degree:1})
	_,err := p.predict(1000,10)
	if err != ENOTENOUGHDATAPOINTS {
 		t.Error("predict returned incorrect error  with no data points",err)
	}

}

func TestPolyPredictHorizontal(t *testing.T) {
	p,_ := newPolyBestFit(PromisesConfig{MaxPoints:10,Degree:1})
	p.add(1,10)
	p.add(2,10)
	ed,err := p.predict(1000,10)
	if err != nil {
 		t.Error("predict failed for horzontal line with two points")
	}
	if ed != 110 {
		t.Error("predict returned wrong day for horizontal line",ed)
	}
}

func TestPolyPredictSlope(t *testing.T) {
	p,_ := newPolyBestFit(PromisesConfig{MaxPoints:10,Degree:1})
	p.add(0,4)
	p.add(1,3)
	clear,err := p.predict(4,1)
	if err !=  nil {
		t.Error("prediced returned error for sloping line",err)
	}
	if clear != 3 {
		t.Error("predict returned wrong prediction for sloping line", clear)
	}
	clear,err = p.predict(9,1)
	if clear !=  4 {
		t.Error("predict not defaulting to simple algo for line going below zero",clear)
	}
}

func TestPolyBackfilledNoPoints(t *testing.T) {
	p,_ := newPolyBestFit(PromisesConfig{MaxPoints:10,Degree:1})
	_,err := p.backfilled(1000,10)
	if err != ENOTENOUGHDATAPOINTS {
 		t.Error("backfilled returned incorrect error  with no data points",err)
	}
}

func TestPolyBackfilledSlope(t *testing.T) {
	p,_ := newPolyBestFit(PromisesConfig{MaxPoints:10,Degree:1})
	p.add(0,4)
	p.add(1,3)
	d,err := p.backfilled(1,3)
	if err !=  nil {
		t.Error("backfilled returned error for sloping line",err)
	}
	if d != 3 {
		t.Error("backfilled returned wrong prediction for sloping line", d)
	}
	d,err = p.backfilled(4,6)
	if err !=  nil {
		t.Error("backfilled failed for line sloping below zero",d)
	}
	if d != 6 {
		t.Error("backfilled returned unexpected value for line sloping below zero",d)
	}
}

func TestPolyBestFitFromTo(t *testing.T) {

	var buff bytes.Buffer
	bf,_ := newPolyBestFit(PromisesConfig{MaxPoints:1000,Degree:2})
	x := epochDays(1)
	for y:=Kilometres(100); y>80;y-=5 {
		bf.add(x,y)
		x++
	}

	err := bf.To(&buff)
	if err != nil {
		t.Error("To failed",err)
	}

	bf2,_ := newPolyBestFit(PromisesConfig{MaxPoints:10,Degree:1}) 
	err = bf2.From(&buff)
	if err != nil {
		t.Error("From failed",err)
	}

	if !reflect.DeepEqual(bf,bf2) {
		t.Error("Deserialised doesnt equal serialized", bf ,bf2)
	}

}

