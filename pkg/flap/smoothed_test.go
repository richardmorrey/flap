package flap

import (
	"testing"
	"reflect"
	"bytes"
)

func TestSmoothSmallWindow(t *testing.T) {
	var s = SmoothYs{windowSize:1,maxYs:3}
	s.AddY(1)
	s.AddY(2)
	s.AddY(3)
	s.AddY(4)
	if !reflect.DeepEqual(s.ys,[]float64{2,3,4}) {
		t.Error("Ys not correct for smoothing window of 1",s.ys)
	}
}

func TestSmoothSetWindows(t *testing.T) {
	var s SmoothYs
	err:=s.SetWindows(1,0)
	if err != EMAXPOINTSBELOWTWO {
		t.Error("Accepted value when window sizes not set")
	}
	err=s.SetWindows(2,0)
	if s.maxYs!=2 || s.windowSize !=1 {
		t.Error("Failed to set default window size",s)
	}
	err=s.SetWindows(3,5)
	if s.maxYs !=3 || s.windowSize!=5 {
		t.Error("Failed to set default window size",s)
	}
}

func TestSmoothWindowSimple(t *testing.T) {
	var s = SmoothYs{windowSize:3,maxYs:5}
	for i:=10.0; i <=100; i+=10 {
		s.AddY(i)
	}
	if !reflect.DeepEqual(s.ys,[]float64{50,60,70,80,90}) {
		t.Error("Ys not correct for smoothing window of 1",s.ys)
	}
}

func TestSmoothFillingWindow(t *testing.T) {
	var s = SmoothYs{windowSize:3,maxYs:3}
	for i:=10.0; i <=30; i+=10 {
		s.AddY(i)
	}
	if !reflect.DeepEqual(s.ys,[]float64{10,15,20}) {
		t.Error("Ys not correct for filling window",s.ys)
	}
}

func TestSmoothedBestFitFromTo(t *testing.T) {

	var buff bytes.Buffer
	var s = SmoothYs{windowSize:3,maxYs:4}
	for i:=10.0; i <=30; i+=10 {
		s.AddY(i)
	}

	err := s.To(&buff)
	if err != nil {
		t.Error("To failed",err)
	}

	s2 := SmoothYs{windowSize:1,maxYs:2}
	err = s2.From(&buff)
	if err != nil {
		t.Error("From failed",err)
	}

	if !reflect.DeepEqual(s,s2) {
		t.Error("Deserialised doesnt equal serialized",s ,s2)
	}

}

