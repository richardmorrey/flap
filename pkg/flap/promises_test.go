package flap

import (
	"testing"
	"reflect"
	//"fmt"
)

type testpredictor struct {
	clearRate epochDays
	stackedLeft Kilometres
	pv predictVersion
}

func (self *testpredictor) add(dist Kilometres) error {
	return ENOTIMPLEMENTED
}

func (self *testpredictor) predict(dist Kilometres, start epochDays) (epochDays,error) {
	return start+self.clearRate*epochDays(dist),nil
}

func (self *testpredictor) version() predictVersion {
	return self.pv
}

func (self *testpredictor) backfilled(d1 epochDays,d2 epochDays) (Kilometres,error) {
	return self.stackedLeft,nil
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
		t.Error("Proposed a clearance with equal trip start and end")
	}
	_,err= ps.Propose(0,1,0,0,&tp)
	if err == nil {
		t.Error("Proposed a clearance with no distance")
	}
	_,err= ps.Propose(0,1,1,0,&tp)
	if err == nil {
		t.Error("Proposed a clearance date with trip start equalling current date")
	}
	_,err= ps.Propose(1,2,1,0,&tp)
	if err == EINVALIDARGUMENT {
		t.Error("Propose rejected valid arguments")
	}
}

func TestProposeFull(t *testing.T) {
	var ps Promises
	var tp testpredictor
	fillpromises(&ps)
	_,err := ps.Propose(epochDays(8).toEpochTime(),epochDays(9).toEpochTime(),10,epochDays(1).toEpochTime(),&tp)
	if err != ENOROOMFORMOREPROMISES {
		t.Error("Propose not erroring when there are no spare promises")
	} 
}

func TestFirstPromise(t *testing.T) {
	var ps Promises
	var tp testpredictor
	tp.clearRate=1
	psold := ps
	p := Promise{TripStart:epochDays(2).toEpochTime(),TripEnd:epochDays(3).toEpochTime(),Distance:3,Clearance:epochDays(17).toEpochTime()}
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
	tp := testpredictor{clearRate:1}
	var ps,psExpected Promises
	var proposal *Proposal
	var err error
	for i := MaxPromises-1; i >=0 ; i-- {
		psExpected.entries[i]=Promise{TripStart:epochDays(10*(MaxPromises-i)).toEpochTime(),
					      TripEnd:epochDays(10*(MaxPromises-i)+6).toEpochTime(),
					      Distance:3,
					      Clearance:epochDays(10*(MaxPromises-i)+9).toEpochTime()}
		proposal,err = ps.Propose(psExpected.entries[i].TripStart,psExpected.entries[i].TripEnd,3,epochDays(1).toEpochTime(),&tp)
		if  err != nil {
			t.Error("Propose failed on non-overlapping promise",err)
			return
		}
		if !reflect.DeepEqual(proposal.entries[0], psExpected.entries[i])  {
			t.Error("Proposal doesnt have expected entry",proposal.entries[0])
		}
		ps.Make(proposal,&tp)
	}
	if !reflect.DeepEqual(proposal.Promises, psExpected) {
		t.Error("Full non-overlapping proposal doesn't have expected value",proposal.Promises)
	}
	if !reflect.DeepEqual(ps, psExpected) {
		t.Error("Full non-overlapping promises doesn't have expected value",ps)
	}
}

func fillpromises(ps *Promises) {
	for i := MaxPromises-1; i >=0 ; i-- {
		ps.entries[i]=Promise{TripStart:epochDays(10*(MaxPromises-i)).toEpochTime(),
				      TripEnd:epochDays(10*(MaxPromises-i)+6).toEpochTime(),
				      Distance:3,
				      Clearance:epochDays(10*(MaxPromises-i)+1).toEpochTime()}
	}
}

func TestOverlappingTrip1(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:1}
	proposal,err := ps.Propose(epochDays(47).toEpochTime(),epochDays(50).toEpochTime(),3, epochDays(20).toEpochTime(),&tp)
	if err != EOVERLAPSWITHNEXTPROMISE{
		t.Error("Propose accepted overlapping trip time", err, proposal)
	}
}

func TestOverlappingTrip2(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:1}
	proposal,err := ps.Propose(epochDays(46).toEpochTime(),epochDays(49).toEpochTime(),3, epochDays(20).toEpochTime(),&tp)
	if err != EOVERLAPSWITHPREVPROMISE{
		t.Error("Propose accepted overlapping trip time", err, proposal)
	}
}

func TestOverlappingTrip3(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:1}
	proposal,err := ps.Propose(epochDays(106).toEpochTime(),epochDays(109).toEpochTime(),3, epochDays(20).toEpochTime(),&tp)
	if err != EOVERLAPSWITHPREVPROMISE{
		t.Error("Propose accepted overlapping trip time", err, proposal)
	}
}

func TestOverlappingTrip4(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:1}
	proposal,err := ps.Propose(epochDays(19).toEpochTime(),epochDays(20).toEpochTime(),3, epochDays(17).toEpochTime(),&tp)
	if err != EOVERLAPSWITHNEXTPROMISE{
		t.Error("Propose accepted overlapping trip time", err, proposal)
	}
}

func TestFitsNoOverlap1(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:1}
	proposal,err := ps.Propose(epochDays(17).toEpochTime(),epochDays(17).toEpochTime()+1,3, epochDays(16).toEpochTime()+10,&tp)
	if err != nil {
		t.Error("Propose rejected valid proposal", err, proposal)
	}
}

func TestFitsNoOverlap2(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:1}
	proposal,err := ps.Propose(epochDays(107).toEpochTime(),epochDays(107).toEpochTime()+1,1, epochDays(16).toEpochTime()+10,&tp)
	if err != nil {
		t.Error("Propose rejected valid proposal", err, proposal)
	}
}

func TestFitsNoOverlap3(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:1}
	proposal,err := ps.Propose(epochDays(47).toEpochTime(),epochDays(47).toEpochTime()+1,3, epochDays(16).toEpochTime()+10,&tp)
	if err != nil {
		t.Error("Propose rejected valid proposal", err, proposal)
	}
}

