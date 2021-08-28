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
	"time"
)
var ESSNODATA 		= errors.New("The first reporting period is not yet completed")
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

// calculateMeanDaily extrapolates the current  mean daily  distance travelled by all travllers across
// the course of a year. It does this as follows:
// (a) Take the total distrance travelled over the last full reporting period (reportdaydelta) 
// (b) If available sdjust for monthly variation using monthly weights associated with first bot spec 
func (self * summaryStats) calculateMeanDaily(mp *ModelParams, now flap.EpochTime) (flap.Kilometres,error) {

	// (a) Retrieve data for last full row
	rdd := mp.ReportDayDelta
	var dayOffset int
	lastFullRow := len(self.Rows)-1
	if lastFullRow >= 0 && self.Rows[lastFullRow].Entries < int(rdd) {
		lastFullRow -= 1
		dayOffset = self.Rows[lastFullRow].Entries
	}
	if (lastFullRow <0) {
		return 0,ESSNODATA
	}
	distDaily := self.Rows[lastFullRow].Travelled

	// (b) Adjust for monthly variation if monthly weights are available ...
	if len(mp.BotSpecs) > 0 && mp.BotSpecs[0].MonthWeights != nil {
		
		// ... calculate mean weight ...
		var weightTotal weight
		for _, w := range mp.BotSpecs[0].MonthWeights {
			weightTotal += w
		}
		meanWeight := float64(weightTotal)/float64(len(mp.BotSpecs[0].MonthWeights))

		// ... establish weight for "now" ...
		dayToAdjustFor := time.Time(now.ToTime())
		dayToAdjustFor.AddDate(0,0,-dayOffset)
		factor := meanWeight/float64(mp.BotSpecs[0].MonthWeights[dayToAdjustFor.Month()-1])
		logInfo("meanWeight:",meanWeight,"month:",dayToAdjustFor.Month(),"factor:",factor)

		// ... adjust daily distance to reflect mean for the year
		distDaily = distDaily * factor
	}

	return flap.Kilometres(distDaily),nil
}

