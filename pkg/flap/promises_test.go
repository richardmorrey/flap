package flap

import (
	"testing"
	//"math"
	//"reflect"
)

type testpredictor struct {
	clearAfter epochDays
}

func (self *testpredictor) add(dist Kilometres) error {
	return ENOTIMPLEMENTED
}

func (self *testpredictor) predict(dist Kilometres, start epochDays) (epochDays,error) {
	return start+self.clearAfter,nil
}

func (self *testpredictor) version() predictVersion {
	return 0 
}

func TestProposeInvalid(t *testing.T) {
	var ps Promises
	var tp testpredictor
	_,err:= ps.Propose(0,1,1,nil)
	if err == nil {
		t.Error("Proposed a clearance with no predictor")
	}
	_,err= ps.Propose(1,1,1,&tp)
	if err == nil {
		t.Error("Proposed a clearance with no predictor")
	}
	_,err= ps.Propose(0,1,0,&tp)
	if err == nil {
		t.Error("Proposed a clearance with no predictor")
	}
	_,err= ps.Propose(0,1,1,&tp)
	if err == EINVALIDARGUMENT {
		t.Error("Propose rejected valid arguments")
	}
}

