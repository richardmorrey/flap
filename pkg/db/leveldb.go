// Package provides DB implementations for use by the flap package. Eacn supported DB technology
// has wrapper structs supporting the three interfaces usable by the main flap package:
// Database, Table, and Iterator.
package db

import (
	"errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"path/filepath"
	"bytes"
	"os"
)
var ENOTIMPLEMENTED = errors.New("Not implemented")
var ETABLEALREADYEXISTS = errors.New("Table already exists")
var ETABLENOTFOUND  = errors.New("Table not found")
var EFAILED = errors.New("Operation failed")
var EINVALIDTABLENAME = errors.New("Invalid table name")

type Database interface
{
	OpenTable(string) (Table,error)
	CloseTable(string) error
	DropTable(string) error
	CreateTable(string) (Table,error)
}

type Reader interface 
{
	Get(string,Serialize) error
}

type Writer interface
{
 	Put(string,Serialize) error
	Delete(string) error
}

type Table interface
{
	Reader
	Writer
	NewIterator(string) (Iterator,error)
	TakeSnapshot() (Snapshot,error)
	MakeBatch(int) (BatchWrite,error)
}

type Snapshot interface
{
	Reader
	Release() error
	NewIterator(string) (Iterator,error)
}

type BatchWrite interface
{
	Writer
	Release() error
}

type Iterator interface {
	Next() (bool)
	Key() (string)
	Value(s Serialize)
	Error() (error)
	Release() error
}

type Serialize interface {
	To(*bytes.Buffer) error
	From(*bytes.Buffer) error
}

type LevelIterator struct {
	iterator iterator.Iterator
}

// Thin wrapper on LevelDB method
func (self *LevelIterator) Next() (bool) {
	return self.iterator.Next() 
}

// Thin wrapper on LevelDB method
func (self *LevelIterator) Key() (string) {
	return string(self.iterator.Key())
}

// Thin wrapper on LevelDB method
func (self *LevelIterator) Value(s Serialize) {
	buff := bytes.NewBuffer(self.iterator.Value())
	s.From(buff)
}

// Thin wrapper on LevelIDB method
func (self *LevelIterator) Error() error {
	return self.iterator.Error()
}

// Thin wrapper on LevelDB method
func (self *LevelIterator) Release() error {
	self.iterator.Release()
	return self.iterator.Error()
}

type LevelSnapshot struct {
	snapshot *leveldb.Snapshot
}

// Release is thin wrapper on LevelDB method
func (self *LevelSnapshot) Release() error {
	self.snapshot.Release()
	return nil
}

// Get is thin wrapper on LevelDB.Get
func (self *LevelSnapshot) Get(key string,s Serialize) error {
	blob, err := self.snapshot.Get([]byte(key),nil)
	if err != nil {
		return err
	}
	buff := bytes.NewBuffer(blob)
	return s.From(buff)
}

// NewIterator creates a thin wrapper around leveldb.Iterator
// It is effectively the factory function for the LevelIterator
// struct.
func (self *LevelSnapshot) NewIterator(prefix string) (Iterator,error) {
	iter := new(LevelIterator)
	if prefix != "" {
		iter.iterator=self.snapshot.NewIterator(util.BytesPrefix([]byte(prefix)),nil)
	} else {
		iter.iterator=self.snapshot.NewIterator(nil,nil)
	}
	return iter,nil
}

// Thin wrapper on LevelDB batch writer
type LevelBatchWrite struct {
	batch *leveldb.Batch
	table *LevelTable
	batchSize int
}

// Put is thin wrapper on leveldb Batch put 
func (self *LevelBatchWrite) Put(key string, s Serialize) error {
	var buff bytes.Buffer
	err := s.To(&buff)
	if err != nil {
		return err
	}
	self.batch.Put([]byte(key),buff.Bytes())
	return self.write(false)
}

// Delete is thin wrapper on leveldb Batch delete
func (self* LevelBatchWrite) Delete(key string) error {
	self.batch.Delete([]byte(key))
	return self.write(false)
}

// Release forces write of any remaining data in the current batch
func (self* LevelBatchWrite) Release() error {
	return self.write(true)
}

// write provides convenient way to write in batches of fixed size
func (self* LevelBatchWrite) write(flush bool) error {
	var err error
	if flush || ((self.batch.Len() % self.batchSize)==0) {
		err = self.table.db.Write(self.batch,nil)
		self.batch.Reset()
	}
	return err
}

// newLevelTable creates a new LevelTable struct, to support Table
// operations for a LevelDB implementation.
// As each LevelDB instance only supports one key-vale store it is
// in fact a wrapper for an entire LevelDB instance. 
func newLevelTable(db *leveldb.DB) *LevelTable {
	table := new(LevelTable)
	table.db= db
	return table
}

type LevelTable struct
{
	db *leveldb.DB
}

// Get is thin wrapper on LevelDB.Get
func (self *LevelTable) Get(key string,s Serialize) error {
	blob, err := self.db.Get([]byte(key),nil)
	if err != nil {
		return err
	}
	buff := bytes.NewBuffer(blob)
	return s.From(buff)
}

// Put is thin wrapper on LevelDB.Put
func (self *LevelTable) Put(key string, s Serialize) error {
	var buff bytes.Buffer
	err := s.To(&buff)
	if err != nil {
		return err
	}
	return self.db.Put([]byte(key),buff.Bytes(),nil)
}

// Delete is thin wrapper on LevelDB.Delete
func (self *LevelTable) Delete(key string) error {
	return self.db.Delete([]byte(key),nil)
}

// NewIterator creates a thin wrapper around leveldb.Iterator
// It is effectively the factory function for the LevelIterator
// struct.
func (self *LevelTable) NewIterator(prefix string) (Iterator,error) {
	iter := new(LevelIterator)
	if prefix != "" {
		iter.iterator=self.db.NewIterator(util.BytesPrefix([]byte(prefix)),nil)
	} else {
		iter.iterator=self.db.NewIterator(nil,nil)
	}
	return iter,nil
}

// TakeSnapshot creates a thin wrapper around leveldb.Iterator
// It is effectively the factory function for the LevelSnapshot
func (self *LevelTable) TakeSnapshot() (Snapshot,error) {
	ss := new(LevelSnapshot)
	var err error
	ss.snapshot,err = self.db.GetSnapshot()
	return ss,err
}

// MakeBatch creates a new LevelDB batch object for batch writes
func (self* LevelTable) MakeBatch(batchSize int) (BatchWrite,error) {
	lb := new(LevelBatchWrite)
	lb.batch = leveldb.MakeBatch(batchSize)
	lb.table=self
	lb.batchSize=batchSize
	return lb,nil
}

func (self *LevelTable) close() {
	self.db.Close()
}

type LevelDB struct
{
	tables map[string]*LevelTable
	Path string
}

// NewLevelIDB creates a new LevelDB struct to support
// Database operations. It manages a map of zero or more LevelTable
// instances, one for each Table be used by the wider system.
func NewLevelDB(path string) *LevelDB {
	db := new(LevelDB)
	db.Path=path
	db.tables = make(map[string]*LevelTable)
	return db
}

// OpenTable attempts to open a LevelTable instance to manage an existing levelDB instance
// for the specified table name.
// If there is already an open instance it returns that. Otherwise it creates
// a new one.
func (self *LevelDB) OpenTable(name string) (Table,error) {

	if self.tables[name] != nil {
		return   self.tables[name],nil
	}

	dbpath := filepath.Join(self.Path,name)
	o := &opt.Options{
		ErrorIfMissing:true,
	}
	db, err := leveldb.OpenFile(dbpath,o)
	switch (err) {
		case nil:
			self.tables[name] = newLevelTable(db)
			return self.tables[name],nil
		case os.ErrNotExist:
			return nil, ETABLENOTFOUND
		default:
			return nil, EFAILED
	}
}

// CloseTable closes any existing LevelTable instance for the specified table name.
// If an instance is found it and the leveldb instance it holds are deleted.
// Otherwise an error is returned. The contents of the table are retained
// on disk.
func (self *LevelDB) CloseTable(name string) error {
	table := self.tables[name]
	if table != nil {
		table.close()
		return nil
	}
	return ETABLENOTFOUND
}

// DropTable deletes all contents of the specified table, removing it entirely
// from disk. Its a bit brutal but it seems the cleanest way of doing this is
// to delete all the files in the table folder.
func (self *LevelDB) DropTable(name string) error {
	if name=="" {
		return EINVALIDTABLENAME
	}
	tablepath := filepath.Join(self.Path,name)
	src, err := os.Stat(tablepath)
	if err != nil {
		return ETABLENOTFOUND
	}
	if !src.IsDir() {
		return EINVALIDTABLENAME
	}
	return os.RemoveAll(tablepath)
}

// Release closes all current table instances. To ensure resource clean-up it must be
// called once the LevelDB instance is finished with.
func (self *LevelDB) Release() error {
	for name, _ := range self.tables {
		self.CloseTable(name)
	}
	return nil
}

// CreateTable creates a new key/value table. In LevelDB this is creating a new LevelDB database file.
// The database file is given the name of the requested table, meaning the table name must be consistent
// with the file path formatting rules of the target OS.
func (self *LevelDB) CreateTable(name string) (Table,error) {

	if self.tables[name] != nil {
		return  nil,ETABLEALREADYEXISTS
	}

	dbpath := filepath.Join(self.Path,name)
	o := &opt.Options{
			ErrorIfExist:true,
		}
	db, err := leveldb.OpenFile(dbpath, o)

	switch (err) {
		case nil:
			self.tables[name] = newLevelTable(db)
			return self.tables[name],nil
		case os.ErrExist:
			return nil, ETABLEALREADYEXISTS
		default:
			return nil, EFAILED
	}
}

