package flap

import (
	//"errors"
	"encoding/binary"
	"bytes"
	"sort"
)

type TransactionType		uint8
const (
	TTFlight	TransactionType = 0x00
	TTTaxiOverhead 	TransactionType = 0x01  
	TTDailyShare	TransactionType = 0x02
	TTBalanceAdjustment TransactionType = 0x03
)
type Transaction struct {
	Date EpochTime
	Distance Kilometres
	TT	TransactionType
}

// To implements db/Serialize
func (self *Transaction) To(buff *bytes.Buffer) error {
	return binary.Write(buff, binary.LittleEndian,self)
}

// From implemments db/Serialize
func (self *Transaction) From(buff *bytes.Buffer) error {
	return binary.Read(buff,binary.LittleEndian,self)
}

const MaxTransactions=100

type Transactions struct {
	entries			[MaxTransactions]Transaction
}

// add adds a single transaction to the top of the list, with the oldest transaction
// being dropped if the list is full.
func (self *Transactions) add(t Transaction) {
	copy(self.entries[1:], self.entries[0:])
	self.entries[0]=t
}

// To implements db/Serialize
func (self *Transactions) To(buff *bytes.Buffer) error {
	n := int32(sort.Search(MaxTransactions,  func(i int) bool {return self.entries[i].Date==0}))
	err := binary.Write(buff, binary.LittleEndian,&n)
	if err != nil {
		return err
	}
	
	for i:=int32(0); i < n; i++ {
		err = self.entries[i].To(buff)
		if err != nil {
			return err
		}
	}
	return err
}

// From implemments db/Serialize
func (self *Transactions) From(buff *bytes.Buffer) error {
	var n int32
	err := binary.Read(buff,binary.LittleEndian,&n)
	if err != nil {
		return err
	}
	for  i:=int32(0); i < n; i++ {
		err = self.entries[i].From(buff)
		if err != nil {
			return err
		}
	}
	return err
}

type TransactionsIterator struct {
	index int
	t *Transactions
}

func (self *TransactionsIterator) Next() (bool) {
	if self.index > 0 {
		self.index--
		return true
	}
	return false
}

func (self *TransactionsIterator) Value() Transaction {
	return self.t.entries[self.index]
}

func (self *TransactionsIterator) Error() error {
	return nil
}

// NewIterator provides iterator for iterating over all transactions from oldest to newest
// by transcation time
func (self *Transactions) NewIterator() *TransactionsIterator {
	iter := new(TransactionsIterator)
	iter.index = sort.Search(MaxTransactions,  func(i int) bool {return self.entries[i].Date==0}) 
	iter.t=self
	return iter
}

