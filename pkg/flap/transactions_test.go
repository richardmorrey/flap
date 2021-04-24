package flap

import (
	"testing"
	"reflect"
	"bytes"
)

func TestAddOne(t *testing.T) {
	var ts Transactions
	ts.add(Transaction{SecondsInDay,1234,TTTaxiOverhead})
	
	if !reflect.DeepEqual(ts.entries[0],Transaction{SecondsInDay,1234,TTTaxiOverhead}) {
		t.Error("Unexpected latest transaction",ts.entries[0])
	}

	if !reflect.DeepEqual(ts.entries[1],Transaction{}) {
		t.Error("Unexpected second transaction",ts.entries[1])
	}
}

func TestSerializeNone(t *testing.T) {
	var ts Transactions
	var buff bytes.Buffer
	err := ts.To(&buff)
	if err != nil {
		t.Error("To failed with error ", err)
	}
	var ts2 Transactions
	err = ts2.From(&buff)
	if err != nil {
		t.Error("From failed with error ",err)
	}
	if !reflect.DeepEqual(ts,ts2) {
		t.Error("Deserialized didnt match serialized",ts2)
	}
}

func TestSerializeOne(t *testing.T) {
	var ts Transactions
	ts.add(Transaction{SecondsInDay,1234,TTTaxiOverhead})
	var buff bytes.Buffer
	err := ts.To(&buff)
	if err != nil {
		t.Error("To failed with error ", err)
	}
	var ts2 Transactions
	err = ts2.From(&buff)
	if err != nil {
		t.Error("From failed with error ",err)
	}
	if !reflect.DeepEqual(ts,ts2) {
		t.Error("Deserialized didnt match serialized",ts2)
	}
}
func TestAddLots(t *testing.T) {
	var ts Transactions
	for i := 1; i <= MaxTransactions+1; i+=1 {
		ts.add(Transaction{EpochTime(SecondsInDay*i),Kilometres(i),TTDailyShare})
	}
	
	if !reflect.DeepEqual(ts.entries[0],Transaction{EpochTime(SecondsInDay*(MaxTransactions+1)),Kilometres(MaxTransactions+1),TTDailyShare}) {
		t.Error("Unexpected latest transaction",ts.entries[0])
	}

	if !reflect.DeepEqual(ts.entries[MaxTransactions-1],Transaction{EpochTime(SecondsInDay*2),Kilometres(2),TTDailyShare}) {
		t.Error("Unexpected oldest transaction",ts.entries[MaxTransactions-1])
	}
}

func TestSerailizeLots(t *testing.T) {
	var ts Transactions
	for i := 1; i <= MaxTransactions+1; i+=1 {
		ts.add(Transaction{EpochTime(SecondsInDay*i),Kilometres(i),TTDailyShare})
	}
	
	var buff bytes.Buffer
	err := ts.To(&buff)
	if err != nil {
		t.Error("To failed with error ", err)
	}
	
	var ts2 Transactions
	err = ts2.From(&buff)
	if err != nil {
		t.Error("From failed with error ",err)
	}
	if !reflect.DeepEqual(ts,ts2) {
		t.Error("Deserialized didnt match serialized",ts2)
	}

}

func TestTransactionsIterateEmpty(t *testing.T) {
	var ts Transactions
	it := ts.NewIterator()
	if it.Next() {
		t.Error("Next returns true for empty Transactions struct")
	}
}

func TestTransactionsIterateFull(t *testing.T) {
	var ts Transactions
	for i := 1; i <= MaxTransactions+1; i+=1 {
		ts.add(Transaction{EpochTime(SecondsInDay*i),Kilometres(i),TTDailyShare})
	}
	

	it := ts.NewIterator()
	i:=0
	for it.Next() {
		i++
		if !reflect.DeepEqual(it.Value(),ts.entries[MaxTransactions-i])  {
			t.Error("Next returns wrong answer for full transactions struct",it.Value())
		}
	}
	if i !=MaxTransactions {
		t.Error("Value failed to iterate over all the values of a full transactions struct",i)
	}
}

