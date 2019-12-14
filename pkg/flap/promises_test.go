package flap

import (
	"testing"
	"reflect"
	//"fmt"
)

type testpredictor struct {
	clearAfter epochDays
	pv predictVersion
}

func (self *testpredictor) add(dist Kilometres) error {
	return ENOTIMPLEMENTED
}

func (self *testpredictor) predict(dist Kilometres, start epochDays) (epochDays,error) {
	return start+self.clearAfter,nil
}

func (self *testpredictor) version() predictVersion {
	return self.pv
}

func (self *testpredictor) backfilled(d1 epochDays,d2 epochDays) (Kilometres,error) {
	return 0, ENOTIMPLEMENTED
}

func TestProposeInvalid(t *testing.T) {
	var ps Promises
	var tp testpredictor
	_,err:= ps.Propose(0,1,1,0,nil)
	if err == nil {
		t.Error("Proposed a clearance with no predictor")
	}
	_,err= ps.Propose(1,1,1,0,&tp)
	if err == nil {
		t.Error("Proposed a clearance with no predictor")
	}
	_,err= ps.Propose(0,1,0,0,&tp)
	if err == nil {
		t.Error("Proposed a clearance with no predictor")
	}
	_,err= ps.Propose(0,1,1,0,&tp)
	if err == EINVALIDARGUMENT {
		t.Error("Propose rejected valid arguments")
	}
}

func TestProposeFull(t *testing.T) {
	var ps Promises
	var tp testpredictor
	for i:=0; i < MaxPromises; i++ {
		ps.entries[i] = Promise{TripStart:epochDays(i+10).toEpochTime()}
	}
	_,err := ps.Propose(epochDays(12).toEpochTime(),epochDays(13).toEpochTime(),10,epochDays(10).toEpochTime(),&tp)
	if err != ENOROOMFORMOREPROMISES {
		t.Error("Propose not erroring when there are no spare promises")
	} 
}

func TestFirstPromise(t *testing.T) {
	var ps Promises
	var tp testpredictor
	tp.clearAfter=14
	psold := ps
	p := Promise{TripStart:epochDays(2).toEpochTime(),TripEnd:epochDays(3).toEpochTime(),Distance:10,Clearance:epochDays(17).toEpochTime()}
	proposal,err := ps.Propose(p.TripStart,p.TripEnd,p.Distance,epochDays(1).toEpochTime(),&tp)
	if err != nil {
		t.Error("Failed to propose a simple promise",err)
		return
	}
	if !reflect.DeepEqual(ps,psold) {
		t.Error("Propose has changed state of Promises", ps)
	}
	if proposal.entries[0] != p {
		t.Error("Propose didn't deliver expected proposal",proposal.entries[0])
	}
}

func TestNonOverlappingPromises(t *testing.T) {
	tp := testpredictor{clearAfter:3}
	var ps,psExpected Promises
	var proposal *Promises
	var err error
	for i := MaxPromises-1; i >=0 ; i-- {
		psExpected.entries[i]=Promise{TripStart:epochDays(10*(MaxPromises-i)).toEpochTime(),
					      TripEnd:epochDays(10*(MaxPromises-i)+6).toEpochTime(),
					      Distance:10,
					      Clearance:epochDays(10*(MaxPromises-i)+9).toEpochTime()}
		proposal,err = ps.Propose(psExpected.entries[i].TripStart,psExpected.entries[i].TripEnd,10,epochDays(1).toEpochTime(),&tp)
		if  err != nil {
			t.Error("Propose failed on non-overlapping promise",err)
			return
		}
		if !reflect.DeepEqual(proposal.entries[0], psExpected.entries[i])  {
			t.Error("Proposal doesnt have expected entry",proposal.entries[0])
		}
		ps.Make(proposal,&tp)
	}
	if !reflect.DeepEqual(proposal, psExpected) {
		t.Error("Full non-overlapping proposal doesn't have expected value",proposal)
	}
}


