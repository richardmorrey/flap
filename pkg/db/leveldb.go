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

type Table interface
{
	Get([]byte) ([]byte, error)
	Put([]byte, []byte) error
	Delete([]byte) error
	NewIterator([]byte) (Iterator,error)
}

type Iterator interface {
	Next() (bool)
	Key() ([]byte)
	Value() ([]byte)
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
func (self *LevelIterator) Key() ([]byte) {
	return self.iterator.Key()
}

// Thin wrapper on LevelDB method
func (self *LevelIterator) Value() ([]byte) {
	return self.iterator.Value()
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
func (self *LevelTable) Get(key []byte) (value []byte,err error) {
	return self.db.Get(key,nil)
}

// Put is thin wrapper on LevelDB.Put
func (self *LevelTable) Put(key []byte, value []byte) error {
	return self.db.Put(key,value,nil)
}

// Delete is thin wrapper on LevelDB.Delete
func (self *LevelTable) Delete(key []byte) error {
	return self.db.Delete(key,nil)
}

// NewIterator creates a thin wrapper around leveldb.Iterator
// It is effectively the factory function for the LevelIterator
// struct.
func (self *LevelTable) NewIterator(prefix []byte) (Iterator,error) {
	iter := new(LevelIterator)
	if prefix != nil {
		iter.iterator=self.db.NewIterator(util.BytesPrefix(prefix),nil)
	} else {
		iter.iterator=self.db.NewIterator(nil,nil)
	}
	return iter,nil
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
