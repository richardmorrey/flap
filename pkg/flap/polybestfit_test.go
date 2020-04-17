package flap

import (
	"testing"
	//"reflect"
)

func TestPolyZeroMaxpoints(t *testing.T) {
	_,err := newPolyBestFit(PromisesConfig{MaxPoints:1})
	if err != EMAXPOINTSBELOWTWO {
		t.Error("Allowing a maxpoints value of less than 2")
	}
}

func TestPolyAddOne(t *testing.T) {
	p,err := newPolyBestFit(PromisesConfig{MaxPoints:10,Degree:1})
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

/*
func TestLongHorizontal(t *testing.T) {
	bf,_ := newPolyBestFit(PromisesConfig{MaxPoints:10,Degree:1})
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
	bf,_ := newBestFit(PromisesConfig{MaxPoints:1000})
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
	bf,_ := newBestFit(PromisesConfig{MaxPoints:1000})
	x := epochDays(1)
	for y:=Kilometres(567); y>0;y-=5 {
		bf.add(x,y)
		x++
	}
	if bf.m !=-5 {
		t.Error("Calculated incorrect gradient for descending line", bf.m)
	}
}
*/
