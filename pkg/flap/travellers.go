package flap

import (
	"github.com/richardmorrey/flap/pkg/db"
	"crypto/sha1"
	"encoding/binary"
	"bytes"
	"errors"
	"encoding/hex"
)

var ETABLENOTOPEN = errors.New("Table not open")
var EPASSPORTSTRINGWRONGLENGTH = errors.New("Passport string is wrong length")

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

func (self *Passport) ToString() string {
	return string(self.Number[:]) + " " + string(self.Issuer[:])
}

func (self *Passport) FromString(s string) error {
	if  len(s) != 13 {
		return EPASSPORTSTRINGWRONGLENGTH
	}
	copy(self.Number[:], s[:9])
	copy(self.Issuer[:], s[10:])
	return nil
}


type Traveller struct {
	version     uint8
	Created	    EpochTime
	passport    Passport
	tripHistory TripHistory
	Transactions Transactions
	Promises    Promises
	Kept	    Promise
	Balance	    Kilometres
}

type ClearanceReason		uint8
const (
	CRGrounded	ClearanceReason = 0x00
	CRMidTrip 	ClearanceReason = 0x01  
	CRKeptPromise	ClearanceReason = 0x02
	CRInCredit	ClearanceReason = 0x03
)

// Cleared returns true if the traveller is cleared
// to travel at the specified date/time
// Also indicates reason for clearance
func (self *Traveller) Cleared(now EpochTime) ClearanceReason {

	// If we have a.Kept then update its Clearance date, which might
	// have changed due to "stacking"
	if self.Kept.Clearance != 0 {
		self.Kept.Clearance,_ = self.Promises.match(self.Kept)
	}

	// Cleared to travel if in a trip, past the clearance date or with a distance account in credit
	cr := CRGrounded
	if (self.tripHistory.MidTrip()) {
		cr = CRMidTrip
	} else if (self.Kept.Clearance > 0 && self.Kept.Clearance <= now) {
		cr = CRKeptPromise 
	} else if self.Balance >=0 {
		cr = CRInCredit
	}
	return cr
}

// keep checks for a matching promise if we are mid-trip. If one is found
// the trip is ended and the clearance date is set to match that in the matched promise
func (self *Traveller) keep() bool {
	kept := false
	if self.MidTrip() {
		p,err := self.Promises.keep(self.tripHistory.tripStartEndLength())
		if err == nil {
			err = self.EndTrip()
			if err == nil {
				logDebug("Kept promise for", self.passport.ToString(), ". Clearance set to",p.Clearance.ToTime())
				self.Kept=p
				kept = true
			} else  {
				logError(err)
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
// If a promises is being applied, then current balance is returned.
func (self *Traveller) submitFlight(flight *Flight,now EpochTime, taxiOH Kilometres, debit bool) (Kilometres,Kilometres,error) {

	//  Make sure we are cleared to travel
	cr := self.Cleared(now) 
	if cr == CRGrounded {
		logDebug("balance:",self.Balance,"cleared:",self.Kept.Clearance.ToTime())
		return 0,0,EGROUNDED
	}

	// Add flight to history
	err := self.tripHistory.AddFlight(flight)
	if err != nil {
		return 0,0,err
	}

	// Debit flight distance and any taxi overhead from balance
	bac := self.Balance
	pd := self.Kept.Distance
	if debit {
		self.transact(-flight.Distance,now,TTFlight)
		if (taxiOH != 0) {
			self.transact(-taxiOH,now,TTTaxiOverhead)
		}
	}

	// Reset any kept promise
	if cr == CRKeptPromise {
		self.Kept = Promise{}
		return bac,pd,nil
	}
	return 0,0,nil
}

// generateKey generates a unique key based on the contents of a
// Passport struct. as the SHA1 of fields in the passport structure.
// Note hash algorithm is use to ensure no hotspots when iterating over
//keys by prefix.
func (self *Passport) generateKey() (string,error) {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian,self)
	if err != nil {
		return "",err
	}
	sha1 := sha1.Sum(buf.Bytes())
	return hex.EncodeToString(sha1[:]), nil
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
	
	if self.table == nil {
		return Traveller{},ETABLENOTOPEN
	}
	var t Traveller
	err := t.get(passport,self.table)
	return t,err
}
func (self* Traveller) get(passport Passport, reader db.Reader) (error) {

	// Create key
	key,err := passport.generateKey()
	if err != nil {
		return err
	}

	// Retrieve value
	return reader.Get(key[:],self)
}

// PutTraveller stores a record for the given Traveller in the
// current table. Any existing record is overwritten.
func (self  *Travellers) PutTraveller(traveller Traveller) error {
	if self.table == nil {
		return ETABLENOTOPEN
	}
	return traveller.put(self.table)
}
func (self* Traveller) put(writer db.Writer) error {

	// Generate key
	key,err := self.passport.generateKey()
	if err != nil {
		return err
	}

	// Put record
	return writer.Put(key[:], self);
}

func (self *Traveller) To(buff *bytes.Buffer) error {
	err := binary.Write(buff,binary.LittleEndian,&(self.version))
	if err != nil {
		return logError(err)
	}
	err = binary.Write(buff,binary.LittleEndian,&(self.Created))
	if err != nil {
		return logError(err)
	}
	err = binary.Write(buff,binary.LittleEndian,&(self.passport))
	if err != nil {
		return logError(err)
	}
	err = self.tripHistory.To(buff)
	if err != nil {
		return logError(err)
	}
	err = self.Promises.To(buff)
	if err != nil {
		return logError(err)
	}	
	err = self.Transactions.To(buff)
	if err != nil {
		return logError(err)
	}

	err = self.Kept.To(buff)
	if err != nil {
		return logError(err)
	}
	return binary.Write(buff,binary.LittleEndian,&(self.Balance))
}

func (self *Traveller) From(buff *bytes.Buffer) error {	
	err := binary.Read(buff,binary.LittleEndian,&(self.version))
	if err != nil {
		return logError(err)
	}
	err = binary.Read(buff,binary.LittleEndian,&(self.Created))
	if err != nil {
		return logError(err)
	}
	err = binary.Read(buff,binary.LittleEndian,&(self.passport))
	if err != nil {
		return logError(err)
	}
	err = self.tripHistory.From(buff)
	if err != nil {
		return logError(err)
	}
	err = self.Promises.From(buff)
	if err != nil {
		return logError(err)
	}
	err = self.Transactions.From(buff)
	if err != nil {
		return logError(err)
	}
	err = self.Kept.From(buff)
	if err != nil {
		return logError(err)
	}
	return binary.Read(buff,binary.LittleEndian,&(self.Balance))
}

// transact carries out a balance adjustment, recording the transaction for posterity
func (self *Traveller) transact(amount Kilometres,now EpochTime, tt TransactionType) {
	self.Transactions.add(Transaction{now,amount, tt})
	self.Balance += amount
}

type TravellersIterator struct {
	iterator db.Iterator
}

func (self *TravellersIterator) Next() (bool) {
	return self.iterator.Next() 
}

func (self *TravellersIterator) Value() Traveller {
	var t Traveller
	self.iterator.Value(&t)
	return t
}

func (self *TravellersIterator) Error() error {
	return self.iterator.Error()
}

func (self *TravellersIterator) Release() error {
	self.iterator.Release()
	return self.iterator.Error()
}

func (self *Travellers) NewIterator(prefix string) (*TravellersIterator,error) {
	iter := new(TravellersIterator)
	var err error
	iter.iterator,err=self.table.NewIterator(prefix)
	return iter,err
}

type TravellersSnapshot struct {
	ss db.Snapshot
}

func (self *TravellersSnapshot) Get(pp Passport) (Traveller,error) {
	var t Traveller
	err := t.get(pp,self.ss)
	return t,err
}

func (self* TravellersSnapshot) Release() error {
	return self.ss.Release()
}

func (self *TravellersSnapshot) NewIterator(prefix string) (*TravellersIterator,error) {
	iter := new(TravellersIterator)
	var err error
	iter.iterator,err=self.ss.NewIterator(prefix)
	return iter,err
}

func (self *Travellers) TakeSnapshot() (*TravellersSnapshot,error) {
	snapshot := new(TravellersSnapshot)
	var err error
	snapshot.ss,err = self.table.TakeSnapshot()
	return snapshot,err
}

type TravellersBatchWrite struct {
	bw db.BatchWrite
}

func (self *TravellersBatchWrite) Put(traveller Traveller) error {
	return traveller.put(self.bw)
}

func (self TravellersBatchWrite) Release() error {
	return self.bw.Release()
}

func (self *Travellers) MakeBatch(size int) (*TravellersBatchWrite,error) {
	bw := new(TravellersBatchWrite)
	var err error
	bw.bw,err = self.table.MakeBatch(size)
	return bw,err
}

