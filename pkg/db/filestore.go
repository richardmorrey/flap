package db

import (
	//"errors"
	//"bytes"
	"cloud.google.com/go/firestore"
)


type FilestoreIterator struct {
	iterator  firestore.DocumentIterator
}

// Thin wrapper on FilestoreDB method
func (self *FilestoreIterator) Next() (bool) {
	return ENOTIMPLEMENTED
}

// Thin wrapper on FilestoreDB method
func (self *FilestoreIterator) Key() ([]byte) {
	return ENOTIMPLEMENTED
}

// Thin wrapper on FilestoreDB method
func (self *FilestoreIterator) Value(s Serialize) {
	return ENOTIMPLEMENTED
}

// Thin wrapper on FilestoreIDB method
func (self *FilestoreIterator) Error() error {
	return ENOTIMPLEMENTED
}

// Thin wrapper on FilestoreDB method
func (self *FilestoreIterator) Release() error {
	return ENOTIMPLEMENTED
}

type FilestoreSnapshot struct {
	snapshot *firestore.Query
}

type FilestoreTable struct {
	snapshot *firestore.Query
}

// Release is thin wrapper on FilestoreDB method
func (self *FilestoreSnapshot) Release() error {
	return ENOTIMPLEMENTED
}

// Get is thin wrapper on FilestoreDB.Get
func (self *FilestoreSnapshot) Get(key []byte,s Serialize) error {
	return ENOTIMPLEMENTED
}

// NewIterator creates a thin wrapper around leveldb.Iterator
// It is effectively the factory function for the FilestoreIterator
// struct.
func (self *FilestoreSnapshot) NewIterator(prefix []byte) (Iterator,error) {
	return ENOTIMPLEMENTED
}

// Thin wrapper on FilestoreDB batch writer
type FilestoreBatchWrite struct {
	batchSize int
}

// Put is thin wrapper on leveldb Batch put 
func (self *FilestoreBatchWrite) Put(key []byte, s Serialize) error {
	return ENOTIMPLEMENTED
}

// Delete is thin wrapper on leveldb Batch put
func (self* FilestoreBatchWrite) Delete(key [] byte) error {
	return ENOTIMPLEMENTED
}

// Release forces write of any remaining data in the current batch
func (self* FilestoreBatchWrite) Release() error {
	return ENOTIMPLEMENTED
}

// write provides convenient way to write in batches of fixed size
func (self* FilestoreBatchWrite) write(flush bool) error {
	return ENOTIMPLEMENTED
}

// newFilestoreTable creates a new FilestoreTable struct, to support Table
// operations for a FilestoreDB implementation.
// As each FilestoreDB instance only supports one key-vale store it is
// in fact a wrapper for an entire FilestoreDB instance. 
/*func newFilestoreTable(db *leveldb.DB) *FilestoreTable {
	return nil
}

type FilestoreTable struct
{
	db *leveldb.DB
}
*/

// Get is thin wrapper on FilestoreDB.Get
func (self *FilestoreTable) Get(key []byte,s Serialize) error {
	return ENOTIMPLEMENTED
}

// Put is thin wrapper on FilestoreDB.Put
func (self *FilestoreTable) Put(key []byte, s Serialize) error {
	return ENOTIMPLEMENTED
}

// Delete is thin wrapper on FilestoreDB.Delete
func (self *FilestoreTable) Delete(key []byte) error {
	return ENOTIMPLEMENTED
}

// NewIterator creates a thin wrapper around leveldb.Iterator
// It is effectively the factory function for the FilestoreIterator
// struct.
func (self *FilestoreTable) NewIterator(prefix []byte) (Iterator,error) {
	return nil, ENOTIMPLEMENTED
}

// TakeSnapshot creates a thin wrapper around leveldb.Iterator
// It is effectively the factory function for the FilestoreSnapshot
func (self *FilestoreTable) TakeSnapshot() (Snapshot,error) {
	return nil, ENOTIMPLEMENTED
}

// MakeBatch creates a new FilestoreDB batch object for batch writes
func (self* FilestoreTable) MakeBatch(batchSize int) (BatchWrite,error) {
	return nil,ENOTIMPLEMENTED
}

func (self *FilestoreTable) close() {
  // ENOTIMPLEMENTED
}

type FilestoreDB struct
{
	tables map[string]*FilestoreTable
}

// NewFilestoreDB creates a new FilestoreDB struct to support
// Database operations. It manages a map of zero or more FilestoreTable
// instances, one for each Table be used by the wider system.
func NewFilestoreDB(path string) *FilestoreDB {
	db := new(FilestoreDB)
	db.tables = make(map[string]*FilestoreTable)
	return db
}

// OpenTable attempts to open a FilestoreTable instance to manage an existing Firestore
// Collection for the specified table name. If there is already an open instance it returns that.
// Otherwise it creates a new one.
func (self *FilestoreDB) OpenTable(name string) (Table,error) {
	if self.tables[name] != nil {
		return   self.tables[name],nil
	}
	return nil,ENOTIMPLEMENTED
}

// CloseTable closes any existing FilestoreTable instance for the specified table name.
// If an instance is found it and associated Collection handle are deleted.
// Otherwise an error is returned. The contents of the table are retained in Filestore.
func (self *FilestoreDB) CloseTable(name string) error {
	return ENOTIMPLEMENTED
}

// DropTable deletes all contents of the specified table. The associated Filestore Collection
// is deleted
func (self *FilestoreDB) DropTable(name string) error {
	return ENOTIMPLEMENTED
}

// Release closes all current table instances. To ensure resource clean-up it must be
// called once the FilestoreDB instance is finished with.
func (self *FilestoreDB) Release() error {
	return nil, ENOTIMPLEMENTED
}

// CreateTable creates a new key/value table. In FilestoreDB this is creating a new Filestore Collection.
// The Collection is given the name of the requested table, meaning the table name must be consistent
// with Collection name formatting rules for Google Cloud Filestore
func (self *FilestoreDB) CreateTable(name string) (Table,error) {

	return nil,ENOTIMPLEMENTED
}

