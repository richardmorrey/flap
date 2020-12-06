package model 

import (
	"encoding/gob"
	"bytes"
	"github.com/richardmorrey/flap/pkg/db"
)

type countryWeights struct {
	Countries []string
	Weights
}

// newCountryWeights constructs a new object to manage weights
// representing the travelling population size for each country.
// Used for setting the county of origin for traveller bots.
func newCountryWeights() *countryWeights {
	countryWeights := new(countryWeights)
	countryWeights.Countries = make([]string,0,100)
	return countryWeights
}

const cwFieldName="country weights"
// save saves the country weights to given table
func (self *countryWeights) save(table db.Table) error  {
	return table.Put(cwFieldName, self)
}

// load attempts to load the country weights from given table
func (self *countryWeights) load(table db.Table) error {
	return table.Get(cwFieldName, self)
}

// update  adds country held in countryState to the list of country weights
func (self *countryWeights) update(cs *countryState) error {
	self.Countries= append(self.Countries,string(cs.countryCode[:2]))
	w,err := cs.country.topWeight()
	if err != nil {
		return logError(err)
	}
	self.add(w)
	return nil
}

// To is part of db.Serialize
func (self* countryWeights) To(b *bytes.Buffer) error {
	enc := gob.NewEncoder(b) 
	return enc.Encode(self)
}

// From is part of db.Serialize
func (self* countryWeights) From(b *bytes.Buffer) error {
	dec := gob.NewDecoder(b)
	return dec.Decode(self)
}
