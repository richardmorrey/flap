package db

import (
	//"fmt"
	"bytes"
	"cloud.google.com/go/datastore"
	"context"
	"google.golang.org/api/iterator"
)
const dataStoreMaxBatch = 500
// buildDatastoreQuery emulates querying with a prefix by building a
// quuery for all keys greater than the given prefix but less than
// the given prefix with the larged possbile unicode character 
// appended
func buildDatastoreQuery(kind string, prefix string) *datastore.Query {
	
	var q *datastore.Query

	if prefix != "" {
		keystart := datastore.NameKey(kind, prefix, nil)
		keyendstr := prefix+"\ufffd"
		keyend := datastore.NameKey(kind, keyendstr, nil)
		q = datastore.NewQuery(kind).Filter("__key__  >=", keystart).
		Filter("__key__  <", keyend)
	} else {
		q = datastore.NewQuery(kind)
	}
	return q
}

// NewIterator creates a thin wrapper around datastore.Iterator
// It is effectively the factory function for the DatastoreIterator
// struct.
func (self *DatastoreTable) NewIterator(prefix string) (Iterator,error) {
	iter := new(DatastoreIterator)
	iter.q = buildDatastoreQuery(self.kind,prefix)
	iter.i = self.client.Run(self.ctx, iter.q)
	iter.e = new(DatastoreEntity)
	return iter,nil
}

type DatastoreIterator struct {
	i  *datastore.Iterator
	q  *datastore.Query
	e  *DatastoreEntity
	k  *datastore.Key
	err  error
}

// Thin wrapper on DatastoreDB method
func (self *DatastoreIterator) Next() (bool) {
    if self.i != nil {
	var err error
	self.k, err = self.i.Next(self.e)
	if err != iterator.Done {
        	return true
	}
    }
    return false
}

// Thin wrapper on DatastoreDB method
func (self *DatastoreIterator) Key() (string) {
	return self.k.Name
}

// Thin wrapper on DatastoreDB method
func (self *DatastoreIterator) Value(s Serialize) {
	buff := bytes.NewBuffer(self.e.Blob)
	self.err = s.From(buff)
}

// Thin wrapper on DatastoreIDB method
func (self *DatastoreIterator) Error() error {
	return self.err
}

// Thin wrapper on DatastoreDB method
func (self *DatastoreIterator) Release() error {
	return ENOTIMPLEMENTED
}

type DatastoreSnapshot struct {
	tx *datastore.Transaction
	table *DatastoreTable
}

type DatastoreTable struct {
	ctx 	context.Context
	client 	*datastore.Client
	kind 	string
}

// Release is thin wrapper on DatastoreDB method
func (self *DatastoreSnapshot) Release() error {
	self.tx = nil
	return nil
}

// Get retrieves entity using the transaction
func (self *DatastoreSnapshot) Get(key string,s Serialize) error {
	
	// Build key
	k := datastore.NameKey(self.table.kind, key, nil)
	e := new(DatastoreEntity)

	// Get value
	if err := self.tx.Get(k, e); err != nil {
		return err
	}

	// Deserialize
	buff := bytes.NewBuffer(e.Blob)
	return s.From(buff)
}

// NewIterator creates a thin wrapper around leveldb.Iterator
// It is effectively the factory function for the DatastoreIterator
// struct	.
func (self *DatastoreSnapshot) NewIterator(prefix string) (Iterator,error) {
	iter := new(DatastoreIterator)
	q := buildDatastoreQuery(self.table.kind,prefix)
	iter.q = q.Transaction(self.tx)
	iter.i = self.table.client.Run(self.table.ctx, iter.q)
	iter.e = new(DatastoreEntity)
	return iter,nil
}

// Thin wrapper on DatastoreDB batch writer
type DatastoreBatchWrite struct {
	batchSize int
	table *DatastoreTable
	keys []*datastore.Key
	entities []*DatastoreEntity
}

// Put is thin wrapper on leveldb Batch put 
func (self *DatastoreBatchWrite) Put(key string, s Serialize) error {
	
	// Generate key and entity to store
	k := datastore.NameKey(self.table.kind, key, nil)
	var buff bytes.Buffer
	err := s.To(&buff)
	if err != nil {
		return err
	}

	// Add to batch and attempt write
	self.keys = append(self.keys,k)
	e := new(DatastoreEntity)
	e.Blob = buff.Bytes()
	self.entities = append(self.entities,e)
	return self.write(false)
}

// Delete is thin wrapper on leveldb Batch put
func (self* DatastoreBatchWrite) Delete(key string) error {
	return ENOTIMPLEMENTED
}

// Release forces write of any remaining data in the current batch
func (self* DatastoreBatchWrite) Release() error {
	return self.write(true)
}

// write provides convenient way to write in batches of fixed size
func (self* DatastoreBatchWrite) write(flush bool) error {
	if flush || len(self.keys) % self.batchSize == 0 {
		_, err := self.table.client.PutMulti(self.table.ctx, self.keys, self.entities)
		if err == nil {
			self.keys = self.keys[:0]
			self.entities = self.entities[:0]
		}
		return err
	}
	return nil
}

type DatastoreEntity struct {
	Blob []byte `datastore:",noindex"`
}

// Get is thin wrapper on DatastoreDB.Get
func (self *DatastoreTable) Get(key string,s Serialize) error {

	// Build key
	k := datastore.NameKey(self.kind, key, nil)
	e := new(DatastoreEntity)

	// Get value
	if err := self.client.Get(self.ctx, k, e); err != nil {
		return err
	}

	// Deserialize
	buff := bytes.NewBuffer(e.Blob)
	return s.From(buff)
}

// Put is thin wrapper on DatastoreDB.Put
func (self *DatastoreTable) Put(key string, s Serialize) error {
	
	// Build key
	k := datastore.NameKey(self.kind, key, nil)

	// Serialize value
	var buff bytes.Buffer
	err := s.To(&buff)
	if err != nil {
		return err
	}

	// Put value
	e := DatastoreEntity{buff.Bytes()}
 	_, err = self.client.Put(self.ctx, k, &e)
	return err
}

// Delete is thin wrapper on DatastoreDB.Delete
func (self *DatastoreTable) Delete(key string) error {
	
	// Build key
	k := datastore.NameKey(self.kind, key, nil)

	// Delete  value
	return self.client.Delete(self.ctx, k)
}

// TakeSnapshot creates a thin wrapper around leveldb.Iterator
// It is effectively the factory function for the DatastoreSnapshot
func (self *DatastoreTable) TakeSnapshot() (Snapshot,error) {
	ss := new(DatastoreSnapshot)
	tx, err := self.client.NewTransaction(self.ctx,datastore.ReadOnly)
	ss.table = self
	ss.tx = tx
	return ss, err
}

// MakeBatch creates a new DatastoreDB batch object for batch writes
func (self* DatastoreTable) MakeBatch(batchSize int) (BatchWrite,error) {
	b := new(DatastoreBatchWrite)
	b.batchSize = dataStoreMaxBatch 
	if batchSize < dataStoreMaxBatch  {
		b.batchSize = batchSize
	}
	b.table = self
	b.keys = make([]*datastore.Key,0,batchSize)
	b.entities = make([]*DatastoreEntity,0,batchSize)
	return b,nil
}

// close closes a table, of which there is no concept in datastore
func (self *DatastoreTable) close() {
}

type DatastoreDB struct
{
	ctx context.Context
	client *datastore.Client
	tables map[string]*DatastoreTable
}

// NewDatastoreDB creates a new DatastoreDB struct to support
// Database operations. It wraps a context for datastore
// under given project name
func NewDatastoreDB(projectname string) *DatastoreDB {
	db := new(DatastoreDB)
	db.tables = make(map[string]*DatastoreTable)
	db.ctx = context.Background()
	client, err := datastore.NewClient(db.ctx,projectname)
	if err != nil {
		return nil
	}
	db.client = client
	return db
}

// The concept of table doesnt exist in datastore. Instead table
// name is used as a "kind" and table objects simply need to hold
// that and the context to interact with the store.
func (self *DatastoreDB) OpenTable(name string) (Table,error) {
	if self.tables[name] == nil {
		t := new(DatastoreTable)
		t.client=self.client
		t.ctx=self.ctx
		t.kind=name
		self.tables[name] = t
	}
	return self.tables[name],nil
}

// The concept of table doesnt exist in datastore
func (self *DatastoreDB) CloseTable(name string) error {
	delete(self.tables,name)
	return nil
}

// DropTable deletes all contents of the specified table. In
// data store this equates to deleting everything under the entity
func (self *DatastoreDB) DropTable(name string) error {

	// Get keys for all entries
	q := datastore.NewQuery(name).Limit(dataStoreMaxBatch)

	deleted := false
	for ;; {
		keys, err := self.client.GetAll(self.ctx,q.KeysOnly(),nil) 
		if err != nil {
			return err
		}

		// Check for some entries
		if len(keys) == 0 {
			if !deleted {
				return ETABLENOTFOUND
			} else {
				return nil
			}
		}
		deleted = true
		
		// Delete all entries
		err = self.client.DeleteMulti(self.ctx,keys)
		if err != nil {
			return err
		}
	}
}

// Release closes all current table instances. To ensure resource clean-up it must be
// called once the DatastoreDB instance is finished with.
func (self *DatastoreDB) Release() error {
	return self.client.Close()
}

// CreateTable creates a new key/value table. In Datastore tables doent exist.
// Equivlent is grouping entities(records) under the same "kind"
func (self *DatastoreDB) CreateTable(name string) (Table,error) {
	if self.tables[name] != nil {
		return nil, ETABLEALREADYEXISTS
	}
	return self.OpenTable(name)
}
