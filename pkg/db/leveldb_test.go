package db

import (
	"testing"
	"os"
	"reflect"
	"path/filepath"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)
var LEVELDBFOLDER="leveldbtest"

func teardown(db *LevelDB) {
	db.Release()
	os.RemoveAll(LEVELDBFOLDER)
}

func TestNewLevelDB(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	if  db.Path != LEVELDBFOLDER {
		t.Error("Path != \"leveldbtest\"", db)
	}
}

func TestCreateTable(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	_,err := db.CreateTable("bands")
	if err != nil {
		t.Error("Create failed",err)
	}
	_,err = db.CreateTable("bands")
	if err == nil {
		t.Error("Creation of table that already exists succeeded")
	}
}

func TestOpenTable(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	_,err := db.OpenTable("albums")
	if err == nil {
		t.Error( "Opened non-existent table")
	}

	_,err = db.CreateTable("albums")
	if err != nil {
		t.Error("Create failed",err)
	}

	table,err := db.OpenTable("albums")
	if err != nil {
		t.Error("Failed to open table",db)
	}
	if table == nil {
		t.Error("Returned nil for Table",db)
	}
	table2,err := db.OpenTable("albums")
	if err != nil {
		t.Error("Failed to open table",db)
	}
	if table2 == nil {
		t.Error("Returned nil for Table",db)
	}
	if table != table2 {
		t.Error("Returned two different structs for the same table",db)
	}
}

func TestPutGet(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	table,_ := db.CreateTable("songs")
	err := table.Put([]byte("The Mountain Goats"), []byte("Waylon Jennings Live"))
	if err  != nil {
		t.Error("Failed to put entry",err)
	}
	value, err := table.Get([]byte("The Mountain Goats"))
	if  err != nil {
		t.Error("Failed to get entry", err)
	}
	if string(value) != "Waylon Jennings Live" {
		t.Error("Failed to get entry", string(value))
	}
}

func TestDropTable(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	table,_ := db.CreateTable("songs")
	err := table.Put([]byte("The Mountain Goats"), []byte("Waylon Jennings Live"))
	db.CloseTable("songs")

	err = db.DropTable("labels")
	if err == nil {
		t.Error("Dropped non-existent table")
	}

	err = db.DropTable("songs")
	if err != nil {
		t.Error("Failed to drop table",err)
	}

	tablepath := filepath.Join(LEVELDBFOLDER,"songs")
	_, err = os.Stat(LEVELDBFOLDER)
	if err != nil {
		t.Error("Drop table deleted more than it should",err)
	}
	o := &opt.Options{
		ErrorIfMissing:true,
	}
	_, err = leveldb.OpenFile(tablepath,o)
	if err == nil {
		t.Error("Dropped table is still there")
	}

}

func TestDelete(t *testing.T) {
	db  := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	table,_ := db.CreateTable("songs")
		err := table.Put([]byte("The Mountain Goats"), []byte("Waylon Jennings Live"))
	if err  != nil {
		t.Error("Failed to put entry",err)
	}
	value, err := table.Get([]byte("The Mountain Goats"))
	if  err != nil {
		t.Error("Failed to get entry", err)
	}
	if string(value) != "Waylon Jennings Live" {
		t.Error("Failed to get entry", string(value))
	}
	err = table.Delete([]byte("The Mountain Goats"))
	if err != nil {
		t.Error("Failed to delete entry", err)
	}
	value, err = table.Get([]byte("The Mountain Goats"))
	if  err == nil {
		t.Error("Succeded in geting deleted entry", err)
	}
	if string(value) != "" {
		t.Error("Value returned for a deleted entry", string(value))
	}
}

func TestIterate(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	table,_ := db.CreateTable("songs")
	songlist:= map[string]string{
		"The Kinks": "Sitting in My Hotel",
		"Sacred Paws": "Wet Graffiti",
		"The Go-betweens": "Born to a Family",
	}
	for artist, song := range(songlist) {
		table.Put([]byte(artist), []byte(song))
	}
	songlistretrieved := map[string]string{}
	iterator,err := table.NewIterator(nil)
	if err != nil {
		t.Error("Failed to create Iterator", err)
	}
	for iterator.Next() {
		songlistretrieved[string(iterator.Key())] = string(iterator.Value())
	}
	if !reflect.DeepEqual(songlistretrieved,songlist) {
		t.Error("Retrieved song list doesnt match", songlist, songlistretrieved)
	}
}

func TestIterateSnapshot(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	table,_ := db.CreateTable("songs")
	songlist:= map[string]string{
		"The Kinks": "Sitting in My Hotel",
		"Sacred Paws": "Wet Graffiti",
		"The Go-betweens": "Born to a Family",
	}
	for artist, song := range(songlist) {
		table.Put([]byte(artist), []byte(song))
	}
	songlistretrieved := map[string]string{}
	ss,err := table.TakeSnapshot()
	if err != nil {
		t.Error("Failed to create snapshot")
	}
	iterator,err := ss.NewIterator(nil)
	if err != nil {
		t.Error("Failed to create iterator from snapshot", err)
	}
	for iterator.Next() {
		songlistretrieved[string(iterator.Key())] = string(iterator.Value())
	}
	if !reflect.DeepEqual(songlistretrieved,songlist) {
		t.Error("Retrieved song list doesnt match", songlist, songlistretrieved)
	}
}

func TestIteratePrefix(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	table,_ := db.CreateTable("songs")
	songlist:= map[string]string{
		"The Kinks": "Sitting in My Hotel",
		"Sacred Paws": "Wet Graffiti",
		"The Go-betweens": "Born to a Family",
	}
	for artist, song := range(songlist) {
		table.Put([]byte(artist), []byte(song))
	}
	songlistretrieved := map[string]string{}
	iterator,err := table.NewIterator([]byte("The"))
	if err != nil {
		t.Error("Failed to create Iterator", err)
	}
	for iterator.Next() {
		songlistretrieved[string(iterator.Key())] = string(iterator.Value())
	}
	if reflect.DeepEqual(songlistretrieved,songlist) {
		t.Error("Retrieved song list matches", songlist, songlistretrieved)
	}
	delete(songlist,"Sacred Paws")
	if !reflect.DeepEqual(songlistretrieved,songlist) {
		t.Error("Retrieved song lists dont match", songlist, songlistretrieved)
	}
}

func TestBatchWrite(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	table,_ := db.CreateTable("songs")
	songlist:= map[string]string{
		"The Kinks": "Sitting in My Hotel",
		"Sacred Paws": "Wet Graffiti",
		"The Go-betweens": "Born to a Family",
	}
	bw,err := table.MakeBatch(2)
	if err != nil {
		t.Error("Failed to create BatchWrite")
	}
	for artist, song := range(songlist) {
		err := bw.Put([]byte(artist), []byte(song))
		if err != nil {
			t.Error("BatchWrite put failed with error",err)
		}
	}
	bw.Release()
	songlistretrieved := map[string]string{}
	iterator,err := table.NewIterator(nil)
	for iterator.Next() {
		songlistretrieved[string(iterator.Key())] = string(iterator.Value())
	}
	if !reflect.DeepEqual(songlistretrieved,songlist) {
		t.Error("BatchWrite didnt write full song list", songlist, songlistretrieved)
	}
}


