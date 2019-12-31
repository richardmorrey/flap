package flap

import (
	"testing"
	"reflect"
	"errors"
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
	backfilledDist Kilometres
	pv predictVersion
	ba backfilledArgs
	pa predictArgs
}

func (self *testpredictor) add(x epochDays,y Kilometres) {
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
	return self.backfilledDist,nil
}

func TestProposeInvalid(t *testing.T) {
	var ps Promises
	var tp testpredictor
	_,err:= ps.propose(0,1,1,0,nil)
	if err == nil {
		t.Error("Proposed a clearance with no predictor")
	}
	_,err= ps.propose(1,1,1,0,&tp)
	if err == nil {
		t.Error("Proposed a clearance with equal trip start and end")
	}
	_,err= ps.propose(0,1,0,0,&tp)
	if err == nil {
		t.Error("Proposed a clearance with no distance")
	}
	_,err= ps.propose(0,1,1,0,&tp)
	if err == nil {
		t.Error("Proposed a clearance date with trip start equalling current date")
	}
	_,err= ps.propose(1,2,1,0,&tp)
	if err == EINVALIDARGUMENT {
		t.Error("Propose rejected valid arguments")
	}
}

func TestProposeFull(t *testing.T) {
	var ps Promises
	var tp testpredictor
	fillpromises(&ps)
	_,err := ps.propose(epochDays(8).toEpochTime(),epochDays(9).toEpochTime(),10,epochDays(1).toEpochTime(),&tp)
	if err != ENOROOMFORMOREPROMISES {
		t.Error("Propose not erroring when there are no spare promises")
	} 
}

func TestFirstPromise(t *testing.T) {
	var ps Promises
	var tp testpredictor
	tp.clearRate=1
	psold := ps
	p := Promise{TripStart:epochDays(2).toEpochTime(),TripEnd:epochDays(3).toEpochTime(),Distance:2,Clearance:epochDays(5).toEpochTime()}
	proposal,err := ps.propose(p.TripStart,p.TripEnd,p.Distance,epochDays(1).toEpochTime(),&tp)
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
					      Clearance:epochDays(10*(MaxPromises-i)+8).toEpochTime()}
		proposal,err = ps.propose(psExpected.entries[i].TripStart,psExpected.entries[i].TripEnd,2,epochDays(1).toEpochTime(),&tp)
		if  err != nil {
			t.Error("Propose failed on non-overlapping promise",err)
			return
		}
		if !reflect.DeepEqual(proposal.entries[0], psExpected.entries[i])  {
			t.Error("Proposal doesnt have expected entry",proposal.entries[0])
		}
		ps.make(proposal,&tp)
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
				      Clearance:epochDays(10*(MaxPromises-i)+8).toEpochTime()}
	}
}

func TestOverlappingTrip1(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:1}
	proposal,err := ps.propose(epochDays(47).toEpochTime(),epochDays(50).toEpochTime(),3, epochDays(20).toEpochTime(),&tp)
	if err != EOVERLAPSWITHNEXTPROMISE{
		t.Error("Propose accepted overlapping trip time", err, proposal)
	}
}

func TestOverlappingTrip2(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:1}
	proposal,err := ps.propose(epochDays(46).toEpochTime(),epochDays(49).toEpochTime(),3, epochDays(20).toEpochTime(),&tp)
	if err != EOVERLAPSWITHPREVPROMISE{
		t.Error("Propose accepted overlapping trip time", err, proposal)
	}
}

func TestOverlappingTrip3(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:1}
	proposal,err := ps.propose(epochDays(106).toEpochTime(),epochDays(109).toEpochTime(),3, epochDays(20).toEpochTime(),&tp)
	if err != EOVERLAPSWITHPREVPROMISE{
		t.Error("Propose accepted overlapping trip time", err, proposal)
	}
}

func TestOverlappingTrip4(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:1}
	proposal,err := ps.propose(epochDays(19).toEpochTime(),epochDays(20).toEpochTime(),3, epochDays(17).toEpochTime(),&tp)
	if err != EOVERLAPSWITHNEXTPROMISE{
		t.Error("Propose accepted overlapping trip time", err, proposal)
	}
}

func TestFitsNoOverlap1(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:0}
	proposal,err := ps.propose(epochDays(17).toEpochTime(),epochDays(17).toEpochTime()+1,1, epochDays(15).toEpochTime()+10,&tp)
	if err != nil {
		t.Error("Propose rejected valid proposal", err, proposal)
	}
}

func TestFitsNoOverlap2(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:1}
	proposal,err := ps.propose(epochDays(107).toEpochTime(),epochDays(107).toEpochTime()+1,1, epochDays(16).toEpochTime()+10,&tp)
	if err != nil {
		t.Error("Propose rejected valid proposal", err, proposal)
	}
}

func TestFitsNoOverlap3(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:0}
	proposal,err := ps.propose(epochDays(47).toEpochTime(),epochDays(47).toEpochTime()+1,1, epochDays(16).toEpochTime()+10,&tp)
	if err != nil {
		t.Error("Propose rejected valid proposal", err, proposal)
	}
}

func TestDoesntFit(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:2}
	proposal,err := ps.propose(epochDays(47).toEpochTime(),epochDays(47).toEpochTime()+1,1, epochDays(16).toEpochTime()+10,&tp)
	if err != EEXCEEDEDMAXSTACKSIZE  {
		t.Error("Propose accepts stacked proposal that doesnt fit",proposal)
	}
}

func TestFitsStacked(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:2}
	proposal,err := ps.propose(epochDays(88).toEpochTime(),epochDays(88).toEpochTime()+1,1, epochDays(16).toEpochTime()+10,&tp)
	if err != nil {
		t.Error("Propose doesnt accept valid stacked proposal",err,proposal)
	}
}

func TestStackTooLong(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	tp := testpredictor{clearRate:2}
	proposal,err := ps.propose(epochDays(78).toEpochTime(),epochDays(78).toEpochTime()+1,1, epochDays(16).toEpochTime()+10,&tp)
	if err != EEXCEEDEDMAXSTACKSIZE {
		t.Error("Propose accepts stacked proposal that doesnt fit",err,proposal)
	}
}
type errpredictor struct {
	err error
}
func (self *errpredictor) add(x epochDays, y Kilometres) {}
func (self *errpredictor) predict(dist Kilometres, start epochDays) (epochDays,error) { return 0, self.err }
func (self *errpredictor) version() predictVersion { return 0 }
func (self *errpredictor) backfilled(d1 epochDays,d2 epochDays) (Kilometres,error) { return 0, self.err }
func TestProposePredNotReady(t *testing.T) {
	var ps Promises
	var ep errpredictor
	ep.err=ENOTENOUGHDATAPOINTS
	p := Promise{TripStart:epochDays(2).toEpochTime(),TripEnd:epochDays(3).toEpochTime(),Distance:2,Clearance:epochDays(4).toEpochTime()}
	proposal,err := ps.propose(p.TripStart,p.TripEnd,p.Distance,epochDays(1).toEpochTime(),&ep)
	if err != nil {
		t.Error("Failed to propose a promise when predicitor isnt ready",err)
		return
	}
	if proposal.entries[0] != p {
		t.Error("Propose didn't deliver expected proposal when predictor isnt ready",proposal.entries[0])
	}
}

func TestProposePredError(t *testing.T) {
	var ps Promises
	var ep errpredictor
	var ETEST = errors.New("Test Error")
	ep.err=ETEST
	p := Promise{TripStart:epochDays(2).toEpochTime(),TripEnd:epochDays(3).toEpochTime(),Distance:2,Clearance:epochDays(4).toEpochTime()}
	_,err := ps.propose(p.TripStart,p.TripEnd,p.Distance,epochDays(1).toEpochTime(),&ep)
	if err != ETEST {
		t.Error("Failed to act on unexpected predictor error",err)
		return
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
	tp := testpredictor{clearRate:1,backfilledDist:3}
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
	if tp.pa.dist !=  17 {
		t.Error("updateStackEntry used wrong dist arg to predictor.predict",tp.pa.dist)
	}
	if tp.pa.start !=  epochDays(26) {
		t.Error("updateStackEntry used wrong d2 arg to predictor.predict",tp.pa.start)
	}
	if ps.entries[1].Clearance != epochDays(20).toEpochTime() {
		t.Error("updateStackEntry set wrong clearance date for the stacked flight",ps.entries[1].Clearance)
	}
	if ps.entries[0].Clearance != epochDays(43).toEpochTime() {
		t.Error("updatedStackEntry set wrong clearance date for the following flight",ps.entries[0].Clearance)
	}
	if ps.entries[1].index != 1 {
		t.Error("updateStackEntry set wrong stack index for the stacked flight",ps.entries[1].index)
	}
	if ps.entries[0].index != 0 {
		t.Error("updatedStackEntry set wrong stack index for the following flight",ps.entries[0].index)
	}
}

func TestUpdateStackEntryContinued(t* testing.T) {
	var ps Promises
	tp := testpredictor{clearRate:1,backfilledDist:3}
	ps.entries[2]=Promise{TripStart:epochDays(1).toEpochTime(),
				      TripEnd:epochDays(5).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(10).toEpochTime(),
				      index:2}
	ps.entries[1]=Promise{TripStart:epochDays(10).toEpochTime(),
				      TripEnd:epochDays(15).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(25).toEpochTime(),
			      	      carriedOver:5}
	ps.entries[0]=Promise{TripStart:epochDays(20).toEpochTime(),
				      TripEnd:epochDays(25).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(35).toEpochTime()}
	err := ps.updateStackEntry(1,&tp)
	if err != nil {
		t.Error("updateStackEntry returned error for simple case",err)
	}
	if ps.entries[2].index !=2  {
		t.Error("updateStackEntry didnt maintain correct index for existing stack entry", ps.entries[2].index)
	}
	if ps.entries[1].index !=3  {
		t.Error("updateStackEntry didnt set correct index for new stack entry", ps.entries[1].index)
	}
	if ps.entries[2].Clearance != epochDays(10).toEpochTime() {
		t.Error("updateStackEntry didnt retain Clearance date for existing stack entry",ps.entries[2].Clearance)
	}
	if ps.entries[1].Clearance != epochDays(20).toEpochTime() {
		t.Error("updateStackEntry didnt change Clearance date for existing stack entry",ps.entries[1].Clearance)
	}
	if ps.entries[0].Clearance != epochDays(48).toEpochTime() {
		t.Error("updateStackEntry didnt set correct Clearance date for unstacked entry",ps.entries[0].Clearance)
	}
}

func TestUpdateStackEntryFull(t* testing.T) {
	var ps Promises
	tp := testpredictor{clearRate:1,backfilledDist:3}
	ps.entries[2]=Promise{TripStart:epochDays(10).toEpochTime(),
				      TripEnd:epochDays(15).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(25).toEpochTime(),
			      	      index:3}
	ps.entries[1]=Promise{TripStart:epochDays(20).toEpochTime(),
				      TripEnd:epochDays(25).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(35).toEpochTime()}
	ps.entries[0]=Promise{TripStart:epochDays(20).toEpochTime(),
				      TripEnd:epochDays(25).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(35).toEpochTime()}
	err := ps.updateStackEntry(1,&tp)
	if err != EEXCEEDEDMAXSTACKSIZE {
		t.Error("updateStackEntry made stack too long",err)
	}
}

func TestRestack(t *testing.T) {
	var ps Promises
	tp := testpredictor{clearRate:1,backfilledDist:4}
	ps.entries[3]=Promise{TripStart:epochDays(1).toEpochTime(),
				      TripEnd:epochDays(5).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(16).toEpochTime()}
	ps.entries[2]=Promise{TripStart:epochDays(10).toEpochTime(),
				      TripEnd:epochDays(15).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(26).toEpochTime()}
	ps.entries[1]=Promise{TripStart:epochDays(20).toEpochTime(),
				      TripEnd:epochDays(25).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(36).toEpochTime()}
	ps.entries[0]=Promise{TripStart:epochDays(30).toEpochTime(),
				      TripEnd:epochDays(36).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(46).toEpochTime()}
	err := ps.restack(2,&tp)
	if (err != nil) {
		t.Error("failed to restack valid promises",err)
	}
	for i := 3; i >= 1; i-- {
		if ps.entries[i].index != stackIndex(4-i) {
			t.Error("restack set incorrect stack index for entry",i,ps.entries[i].index)
		}
		if ps.entries[i].Clearance != ps.entries[i-1].TripStart {
			t.Error("restack set learance date that doesnt match start date of next trip", i, ps.entries[i].Clearance,ps.entries[i-1].TripStart)
		}
	}
	if ps.entries[0].index !=0 {
		t.Error("restack set incorrect stack index for latest entry",ps.entries[0])
	}
	if ps.entries[0].Clearance != epochDays(65).toEpochTime() {
		t.Error("restack inal clearance data doesnt account for total carry over from previous stacked flights", ps.entries[0].Clearance)
	}
}

func TestRestackFull(t *testing.T) {
	var ps Promises
	tp := testpredictor{clearRate:1,backfilledDist:4}
	ps.entries[4]=Promise{TripStart:epochDays(1).toEpochTime(),
				      TripEnd:epochDays(5).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(16).toEpochTime()}
	ps.entries[3]=Promise{TripStart:epochDays(10).toEpochTime(),
				      TripEnd:epochDays(15).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(26).toEpochTime()}
	ps.entries[2]=Promise{TripStart:epochDays(20).toEpochTime(),
				      TripEnd:epochDays(25).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(36).toEpochTime()}
	ps.entries[1]=Promise{TripStart:epochDays(30).toEpochTime(),
				      TripEnd:epochDays(36).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(46).toEpochTime()}
	ps.entries[0]=Promise{TripStart:epochDays(40).toEpochTime(),
				      TripEnd:epochDays(36).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(56).toEpochTime()}
	err := ps.restack(3,&tp)
	if err != EEXCEEDEDMAXSTACKSIZE {
		t.Error("restack succeeded where no valid stacking available")
	}
}

func TestRestackOldest(t *testing.T) {
	var ps Promises
	tp := testpredictor{clearRate:1,backfilledDist:4}
	ps.entries[3]=Promise{TripStart:epochDays(1).toEpochTime(),
				      TripEnd:epochDays(5).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(16).toEpochTime()}
	ps.entries[2]=Promise{TripStart:epochDays(10).toEpochTime(),
				      TripEnd:epochDays(15).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(26).toEpochTime()}
	ps.entries[1]=Promise{TripStart:epochDays(20).toEpochTime(),
				      TripEnd:epochDays(25).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(36).toEpochTime()}
	ps.entries[0]=Promise{TripStart:epochDays(30).toEpochTime(),
				      TripEnd:epochDays(36).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(46).toEpochTime()}
	err := ps.restack(3,&tp)
	if (err != nil) {
		t.Error("failed to restack valid promises",err)
	}
	for i := 3; i >= 1; i-- {
		if ps.entries[i].index != stackIndex(4-i) {
			t.Error("restack set incorrect stack index for entry",i,ps.entries[i].index)
		}
		if ps.entries[i].Clearance != ps.entries[i-1].TripStart {
			t.Error("restack set clearance date that doesnt match start date of next trip", i, ps.entries[i].Clearance,ps.entries[i-1].TripStart)
		}
	}
	if ps.entries[0].index !=0 {
		t.Error("restack set incorrect stack index for latest entry",ps.entries[0])
	}
	if ps.entries[0].Clearance != epochDays(65).toEpochTime() {
		t.Error("restack clearance date doesn't account for total carry over from previous stacked flights", ps.entries[0].Clearance)
	}
}

func TestMakeInvalid(t *testing.T) {
	var ps Promises
	var pl Proposal
	tp := testpredictor{pv:1}
	fillpromises(&pl.Promises)
	psinit := ps
	err:= ps.make(&pl,&tp)
	if err != EPROPOSALEXPIRED {
		t.Error("Make accepts proposal made with older predictor")
	}
	if !reflect.DeepEqual(ps.entries,psinit.entries) {
		t.Error("Make changed its promises when presented with an invalid proposal",ps.entries) 
	}
}

func TestMakeValid(t *testing.T) {
	var ps Promises
	pl := Proposal{version:1}
	tp := testpredictor{pv:1}
	fillpromises(&pl.Promises)
	err:= ps.make(&pl,&tp)
	if err != nil {
		t.Error("Make rejects valid proposal")
	}
	if !reflect.DeepEqual(ps.entries,pl.entries) {
		t.Error("Make didn't adopt the promises of a valid proposal",ps.entries) 
	}
}

func TestKeepExact(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	c,err:= ps.keep(epochDays(50).toEpochTime(),epochDays(56).toEpochTime(),2)
	if err != nil {
		t.Error("keep can't find valid promise",err)
	}
	if c != epochDays(58).toEpochTime() {
		t.Error("keep returns incorrect clearance date for valid promise",c)
	}
}

func TestKeepExactiOldest(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	c,err:= ps.keep(epochDays(10).toEpochTime(),epochDays(16).toEpochTime(),2)
	if err != nil {
		t.Error("keep can't find valid promise",err)
	}
	if c != epochDays(18).toEpochTime() {
		t.Error("keep returns incorrect clearance date for valid promise",c)
	}
}

func TestKeepExactNewest(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	st:= epochDays((MaxPromises-1)*10)
	c,err:= ps.keep(st.toEpochTime(),(st+6).toEpochTime(),2)
	if err != nil {
		t.Error("keep can't find valid promise",err)
	}
	if c != (st+8).toEpochTime() {
		t.Error("keep returns incorrect clearance date for valid promise",c)
	}
}

func TestKeepLaterStart(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	c,err:= ps.keep(epochDays(50).toEpochTime()+1,epochDays(56).toEpochTime(),2)
	if err != nil {
		t.Error("keep can't find valid promise",err)
	}
	if c != epochDays(58).toEpochTime() {
		t.Error("keep returns incorrect clearance date for valid promise",c)
	}
}

func TestKeepEarlierEnd(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	c,err:= ps.keep(epochDays(50).toEpochTime(),epochDays(56).toEpochTime()-1,2)
	if err != nil {
		t.Error("keep can't find valid promise",err)
	}
	if c != epochDays(58).toEpochTime() {
		t.Error("keep returns incorrect clearance date for valid promise",c)
	}
}

func TestKeepEarilerStart(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	_,err:= ps.keep(epochDays(50).toEpochTime()-1,epochDays(56).toEpochTime(),2)
	if err != EPROMISEDOESNTMATCH {
		t.Error("keep matched promise with earlier start time",err)
	}
}

func TestKeepLaterEnd(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	_,err:= ps.keep(epochDays(50).toEpochTime(),epochDays(56).toEpochTime()+1,2)
	if err != EPROMISEDOESNTMATCH {
		t.Error("keep matched promise with later end time",err)
	}
}

func TestKeepWrongDistance(t *testing.T) {
	var ps Promises
	fillpromises(&ps)
	_,err:= ps.keep(epochDays(50).toEpochTime(),epochDays(56).toEpochTime(),3)
	if err != EPROMISEDOESNTMATCH {
		t.Error("keep matched promise with differnt distance",err)
	}
}

func TestIterateEmpty(t *testing.T) {
	var ps Promises 
	it := ps.NewIterator()
	if it.Next() {
		t.Error("Next returns true for empty Promises struct")
	}
}

func TestIterateFull(t *testing.T) {
	var ps Promises 
	fillpromises(&ps)
	it := ps.NewIterator()
	i:=0
	for it.Next() {
		i++
		if !reflect.DeepEqual(it.Value(),ps.entries[MaxPromises-i])  {
			t.Error("Next returns wrong answer for full promises  struct",it.Value())
		}
	}
	if i !=MaxPromises {
		t.Error("Value failed to iterate over all the values of a full promises struct",i)
	}
}

func TestIterateOne(t *testing.T) {
	var ps Promises
	ps.entries[0]=Promise{TripStart:epochDays(40).toEpochTime(),
				      TripEnd:epochDays(36).toEpochTime(),
				      Distance:10,
				      Clearance:epochDays(56).toEpochTime()}
	it := ps.NewIterator()
	if !it.Next() {
		t.Error("Next failed to iterate over a single value")
	}
	if !reflect.DeepEqual(it.Value(),ps.entries[0]) {
		t.Error("Value returned wrong value for singlei value promises")
	}
	if it.Next() {
		t.Error("Next returns true for second invocation against single entry promises")
	}
}
