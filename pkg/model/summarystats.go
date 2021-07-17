package model

import (
	"github.com/richardmorrey/flap/pkg/flap"
	"github.com/richardmorrey/flap/pkg/db"
	"path/filepath"
	"fmt"
	"os"
	"encoding/json"
	"encoding/gob"
	"gonum.org/v1/plot/plotter"
	"bytes"
	"errors"
)
var ESSNODATA 		= errors.New("The first reporting period is not yet completed")
var ESSNOTENOUGHDATA     = errors.New("There is not enough data to asses traveller over the proceeding calendar year")
var ESSNOTABLE = errors.New("Table not provided")

type summaryStatsRow 	struct {
	DailyTotal		float64	
	Travelled		float64
	Flights			float64
	Travellers		float64
	Grounded		float64
	Share			float64
	Date			flap.EpochTime
	Entries			int
}

type summaryStats struct {
	Rows []summaryStatsRow
}

// newRow starts a new record for counting stats
func (self *summaryStats) newRow() {
	self.Rows = append(self.Rows,summaryStatsRow{})
}

// add adds the provide numbers to summary for the latest day
func (self *summaryStats) update(increment summaryStatsRow, rdd flap.Days, t db.Table) {
	self.load(t)
	i := len(self.Rows) -1
	if self.Rows[i].Entries == int(rdd)  {
		i += 1 
		self.newRow()
	}
	self.Rows[i].DailyTotal +=increment.DailyTotal
	self.Rows[i].Travelled += increment.Travelled
	self.Rows[i].Flights += increment.Flights
	self.Rows[i].Travellers += increment.Travellers
	self.Rows[i].Grounded += increment.Grounded
	self.Rows[i].Share += increment.Share
	self.Rows[i].Date = increment.Date
	self.Rows[i].Entries += 1
	self.save(t)
}

func (self* summaryStats) To(b *bytes.Buffer) error {
	enc := gob.NewEncoder(b) 
	return enc.Encode(self)
}

func (self* summaryStats) From(b *bytes.Buffer) error {
	dec := gob.NewDecoder(b)
	return dec.Decode(self)
}

const summaryStatsRecordKey="summarystats"

// load loads engine state from given table
func (self *summaryStats) load(t db.Table) error {
	var err error
	if t !=nil {
		err =  t.Get(summaryStatsRecordKey,self)
	}
	if len(self.Rows) == 0 {
		self.newRow()
	}
	return err
}

// save saves engine state to given table
func (self *summaryStats)  save(t db.Table) error {
	if t==nil {
		return ESSNOTABLE
	}
	return t.Put(summaryStatsRecordKey,self)
}

// returns summary stats as a JSON string 
func (self * summaryStats) asJSON() string {
	jsonData, _ := json.MarshalIndent(self.Rows, "", "    ")
	return string(jsonData)
}

// compile prepares summary stats for reporting
func (self* summaryStats) compile(path string,rdd flap.Days) (plotter.XYs, plotter.XYs) {

	travelledPts := make(plotter.XYs, 0)
	allowancePts := make(plotter.XYs, 0)
	fn := filepath.Join(path,"summary.csv")
	fh,_ := os.Create(fn)
	if fh == nil {
		return nil,nil
	}

	
	// Summmary stats title
	fh.WriteString("Date,DailyTotal,Travelled,Travellers,Grounded,Share\n")

	day := rdd
	for _,row := range self.Rows { 

		// Skip incomplete rows
		if row.Entries < int(rdd) {
			continue
		}

		// Summary stats line
		line := fmt.Sprintf("%d,%.2f,%.2f,%d,%d,%.2f\n",day,
				flap.Kilometres(row.DailyTotal),flap.Kilometres(row.Travelled),
				uint64(row.Travellers),uint64(row.Grounded),row.Share)
		fh.WriteString(line)
		day += rdd
	
		// Update points for summary graph
		travelledPts = append(travelledPts,plotter.XY{X:float64(day),Y:row.Travelled})
		allowancePts = append(allowancePts,plotter.XY{X:float64(day),Y:row.DailyTotal})
	}
	return  travelledPts,allowancePts
}

func (self * summaryStats) distPerTraveller(i int) float64 {
	if self.Rows[i].Travellers == 0 {
		return 0
	}
	return float64(self.Rows[i].Travelled)/float64(self.Rows[i].Travellers)
}

// calculateMeanDaily extrapolates the current  mean distance travelled by each traveller across
// the course of a year. It does this as follows:
// (a) Take the current distrance travelled over the last full reporting perioed (reportdaydelta)
// (b) Calculate the % change between this and the equivlent reporting period a year ago.
// (c) Take sum of distance times this proportion for all reporting periods in the past year.
// (d) Divide by the total number of days covered by all the reporting periods taken at step c.
func (self * summaryStats) calculateMeanDaily(rdd flap.Days) (flap.Kilometres,error) {

	// (a) Retrieve data for last full row
	now := len(self.Rows)-1
	if now >= 0 && self.Rows[now].Entries < int(rdd) {
		now -= 1
	}
	if (now <0) {
		return 0,ESSNODATA
	}
	distNow := self.distPerTraveller(now)

	// (b) Calc % change from a year ago
	yearago := now - (365/int(rdd))
	if yearago < 0 {
		return 0,ESSNOTENOUGHDATA
	}
	distThen := self.distPerTraveller(yearago)
	if distThen == 0 {
		return 0,nil
	}

	// (c) Calc full year distance per traveller
	var yearDistNow float64
	change := distNow/distThen
	for i:=yearago; i < now; i++ {
		yearDistNow += change*self.distPerTraveller(i)
	}

	// (d) Divide by days covered to give daily mean per traveller
	return flap.Kilometres(yearDistNow/(float64(now-yearago))),nil
}

