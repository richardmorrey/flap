package model

import (
	"testing"
)


func TestFindNoWeights(t *testing.T) {
	var w Weights
	_,err := w.find(0)
	if err == nil {
		t.Error("find succeeded with no Weights")
	}
}

func TestFindOneWeight(t *testing.T) {
	var w Weights
	w.add(1)
	i,err := w.find(0)
	if err != nil {
		t.Error("find failed with one weight")
	}
	if i != 0 {
		t.Error("find returned wrong index for one weight")
	}
	i,err = w.find(1)
	if err != nil {
		t.Error("find failed with one weight")
	}
	if i != 0 {
		t.Error("find returned wrong index for one weight")
	}
	i,err = w.find(2)
	if err == nil {
		t.Error("find found 2 with one weight")
	}
}

func TestFindMoreWeights(t *testing.T) {
	var w Weights
	w.add(1)
	w.add(9)
	w.add(90)
	i,err := w.find(1)
	if err != nil {
		t.Error("find failed for 1")
	}
	if i != 0 {
		t.Error("find returned wrong index for 1")
	}
	i,err = w.find(1)
	if err != nil {
		t.Error("find failed for 1")
	}
	if i != 0 {
		t.Error("find returned wrong index for 1")
	}
	i,err = w.find(2)
	if err != nil {
		t.Error("find failed for 2")
	}
	if i != 1 {
		t.Error("find returned wrong index for 2")
	}
	i,err = w.find(10)
	if err != nil {
		t.Error("find failed for 10")
	}
	if i != 1 {
		t.Error("find returned wrong index for 10")
	}
	i,err = w.find(11)
	if err != nil {
		t.Error("find failed for 11")
	}
	if i != 2 {
		t.Error("find returned wrong index for 11",i,w.Scale)
	}
	i,err = w.find(100)
	if err != nil {
		t.Error("find failed for 100")
	}
	if i != 2 {
		t.Error("find returned wrong index for 100")
	}
	i,err = w.find(101)
	if err == nil {
		t.Error("find succeeded for 101")
	}
}
