package flap

import (
	"github.com/richardmorrey/flap/pkg/db"
	"crypto/sha1"
	"encoding/binary"
	"bytes"
	"errors"
	"unsafe"
	//"fmt"
)

var ETABLENOTOPEN = errors.New("Table not open")

type PassportNumber [9]byte
type IssuingCountry [3]byte

type Passport struct {
	Number PassportNumber
	Issuer IssuingCountry
}

func NewPassport(number string,issuer string) Passport {
	var passport Passport
	copy(passport.Number[:],number)
	copy(passport.Issuer[:],issuer)
	return passport
}

type passportKey [20]byte

type Traveller struct {
	passport    Passport
	tripHistory TripHistory
	Promises    Promises
	cleared	    EpochTime
	balance	    Kilometres
}

// Cleared returns true if the traveller is cleared
// to travel at the specified date/time
func (self *Traveller) Cleared(now EpochTime) bool {	
	return self.tripHistory.MidTrip() || (self.cleared > 0 && self.cleared <= now) || self.balance >=0
}

// keep checks for a matching promise if we are mid-trip. If one is found
// the trip is ended and the clearance data is set to match that in the matched promise
func (self *Traveller) keep() bool {
	kept := false
	if self.MidTrip() {
		cd,err := self.Promises.keep(self.tripHistory.tripStartEndLength())
		if err == nil {
			err = self.EndTrip()
			if err == nil {
				logDebug("kept promise, clearday set to ",cd/SecondsInDay)
				self.cleared=cd
				kept = true
			}
		}
	}
	return kept
}

// Wrapper for TripHistory endTrip
func (self *Traveller) EndTrip() error {
	return self.tripHistory.EndTrip()
}

// Wrapper for TripHistory reopenTrip
func (self *Traveller) ReopenTrip() error {
	return self.tripHistory.ReopenTrip()
}

// Wrapper for TripHistory midTrip
func (self *Traveller) MidTrip() bool {
	return self.tripHistory.MidTrip()
}

// Wrapper for TripHistory.AsJSON
func (self *Traveller) AsJSON() string {
	return self.tripHistory.AsJSON()
}

// Wrapper for TripHistory.AsKML
func (self *Traveller) AsKML(a *Airports) string {
	return self.tripHistory.AsKML(a)
}

// submitFlight adds given flight to trip history and if traveller is cleared for travel.
// Also, If "debit" is true, the flight distance  is subtracted from the traveller's distance balance.
// If traveller is not cleared for travel no action is taken and an error is returned.
func (self *Traveller) submitFlight(flight *Flight,now EpochTime, debit bool) error {
	if !self.Cleared(now) {
		logDebug("balance:",self.balance,"clear day:",self.cleared/SecondsInDay)
		return EGROUNDED
	}
	err := self.tripHistory.AddFlight(flight)
	if err != nil {
		return err
	}
	//logDebug("triphistory:",self.tripHistory)
	if debit {
		self.balance -= flight.distance
	}
	// Make sure clearance promise only gets applied once
	self.cleared = 0
	return nil
}

// generateKey generates a unique key based on the contents of a
// Passport struct. as the SHA1 of fields in the passport structure.
// Note hash algorithm is use to ensure no hotspots when iterating over
//keys by prefix.
func (self *Passport) generateKey() (passportKey,error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian,self)
	if err != nil {
		return passportKey{},err
	}
	return sha1.Sum(buf.Bytes()), nil
}

type Travellers struct {
	table db.Table
}

// NewTravellers opens a interface for the Travellers table from the 
// given database. If the table doesnt exist it is created.
const travellersTableName = "travellers"
func NewTravellers(flapdb db.Database) *Travellers {
	travellers := new(Travellers)
	table,err := flapdb.OpenTable(travellersTableName)
	if err == db.ETABLENOTFOUND { 
		table,err = flapdb.CreateTable(travellersTableName)
	}
	if err != nil {
		return nil
	}
	travellers.table  = table
	return travellers
}

// Drops travellers table from given database
func dropTravellers(database db.Database) error {
	return database.DropTable(travellersTableName)
}

// GetTraveller finds and returns a record matching the
// give passport details in the current table.
func (self *Travellers) GetTraveller(passport Passport) (Traveller,error) {
	
	// Retrieve record
	if self.table == nil {
		return Traveller{},ETABLENOTOPEN
	}

	it,err := self.NewIterator(nil)
	defer it.Release()
	key,err := passport.generateKey()
	if err != nil {
		return Traveller{},err
	}
	blob, err := self.table.Get(key[:])
	if err != nil {
		return Traveller{},err
	}

	// Deserialize and return struct
	return (*(*Traveller)(unsafe.Pointer(&blob[0]))), err
}

// PutTraveller stores a record for the given Traveller in the
// current table. Any existing record is overwritten.
var sizeOfTraveller = unsafe.Sizeof(Traveller{})
func (self  *Travellers) PutTraveller(traveller Traveller) error {
	
	// Serialize record
	if self.table == nil {
		return ETABLENOTOPEN
	}
	data := (*(*[1<<31 - 1]byte)(unsafe.Pointer(&traveller)))[:sizeOfTraveller]

	// Put record
	key,err := traveller.passport.generateKey()
	if err != nil {
		return err
	}
	return self.table.Put(key[:], data);
}

type TravellersIterator struct {
	iterator db.Iterator
}

func (self *TravellersIterator) Next() (bool) {
	return self.iterator.Next() 
}

func (self *TravellersIterator) Value() Traveller {
	blob  := self.iterator.Value()
	return (*(*Traveller)(unsafe.Pointer(&blob[0])))
}

func (self *TravellersIterator) Error() error {
	return self.iterator.Error()
}

func (self *TravellersIterator) Release() error {
	self.iterator.Release()
	return self.iterator.Error()
}

func (self *Travellers) NewIterator(prefix []byte) (*TravellersIterator,error) {
	iter := new(TravellersIterator)
	var err error
	if prefix != nil {
		iter.iterator,err=self.table.NewIterator(prefix)
	} else {
		iter.iterator,err=self.table.NewIterator(nil)
	}
	return iter,err
}


