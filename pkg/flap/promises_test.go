package flap

import (
	"testing"
	"reflect"
	//"fmt"
)

type backfilledArgs struct {
	d1 epochDays
	d2 epochDays
}

type predictArgs struct {
	dist Kilometres
	start epochDays
}

type testpredictor struct {
	clearRate epochDays
	stackedLeft Kilometres
	pv predictVersion
	ba backfilledArgs
	pa predictArgs
}

func (self *testpredictor) add(dist Kilometres) error {
	return ENOTIMPLEMENTED
}

func (self *testpredictor) predict(dist Kilometres, start epochDays) (epochDays,error) {
	self.pa.dist=dist
	self.pa.start=start
	return start+self.clearRate*epochDays(dist),nil
}

func (self *testpredictor) version() predictVersion {
	return self.pv
}

func (self *testpredictor) backfilled(d1 epochDays,d2 epochDays) (Kilometres,error) {
	self.ba.d1=d1
	self.ba.d2=d2
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
	p := Promise{TripStart:epochDays(2).toEpochTime(),TripEnd:epochDays(3).toEpochTime(),Distance:2,Clearance:epochDays(6).toEpochTime()}
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
					      Distance:2,
					      Clearance:epochDays(10*(MaxPromises-i)+9).toEpochTime()}
		proposal,err = ps.Propose(psExpected.entries[i].TripStart,psExpected.entries[i].TripEnd,2,epochDays(1).toEpochTime(),&tp)
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
				      Distance:2,
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
	tp := testpredictor{clearRate:0}
	proposal,err := ps.Propose(epochDays(17).toEpochTime(),epochDays(17).toEpochTime()+1,1, epochDays(15).toEpochTime()+10,&tp)
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
	tp := testpredictor{clearRate:0}
	proposal,err := ps.Propose(epochDays(47).toEpochTime(),epochDays(47).toEpochTime()+1,1, epochDays(16).toEpochTime()+10,&tp)
	if err != nil {
		t.Error("Propose rejected valid proposal", err, proposal)
	}
}

func TestUpdateStackEntryInvalid(t *testing.T) {
	var ps Promises
	tp := testpredictor{clearRate:1}
	
	err := ps.updateStackEntry(0,&tp)
	if err != EINVALIDARGUMENT {
		t.Error("updateStackEntry accepted promise with no successors")
	}
	err = ps.updateStackEntry(MaxPromises,&tp)
	if err != EINVALIDARGUMENT {
		t.Error("updateStackEntry accepted out-of-range index")
	}
 	err = ps.updateStackEntry(1,nil)
	if err != EINVALIDARGUMENT {
		t.Error("updateStackEntry accepted nil predictor")
	}
}

func TestUpdateStackEntrySimple(t* testing.T) {
	var ps Promises
	tp := testpredictor{clearRate:1,stackedLeft:3}
	ps.entries[1]=Promise{TripStart:epochDays(10).toEpochTime(),
				      TripEnd:epochDays(15).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(25).toEpochTime()}
	ps.entries[0]=Promise{TripStart:epochDays(20).toEpochTime(),
				      TripEnd:epochDays(25).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(35).toEpochTime()}
	err := ps.updateStackEntry(1,&tp)
	if err != nil {
		t.Error("updateStackEntry returned error for simple case",err)
	}
	if tp.ba.d1 !=  epochDays(16) {
		t.Error("updateStackEntry used wrong d1 arg to predictor.backfilled",tp.ba.d1)
	}
	if tp.ba.d2 !=  epochDays(20) {
		t.Error("updateStackEntry used wrong d2 arg to predictor.backfilled",tp.ba.d1)
	}
	if tp.pa.dist !=  13 {
		t.Error("updateStackEntry used wrong d1 arg to predictor.backfilled",tp.pa.dist)
	}
	if tp.pa.start !=  epochDays(26) {
		t.Error("updateStackEntry used wrong d2 arg to predictor.backfilled",tp.pa.start)
	}
	if ps.entries[1].Clearance != epochDays(20).toEpochTime() {
		t.Error("updateStackEntry set wrong clearance date for the stacked flight",ps.entries[1].Clearance)
	}
	if ps.entries[0].Clearance != epochDays(39).toEpochTime() {
		t.Error("updatedStackEntry set wrong clearnace date for the following fihgt",ps.entries[0].Clearance)
	}
	if ps.entries[1].index != 1 {
		t.Error("updateStackEntry set wrong stack index for the stacked flight",ps.entries[1].index)
	}
	if ps.entries[0].index != 0 {
		t.Error("updatedStackEntry set wrong stack index for the following flight",ps.entries[0].index)
	}
}

