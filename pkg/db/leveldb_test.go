package db

import (
	"testing"
	"os"
	"reflect"
	"path/filepath"
	"bytes"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)
var LEVELDBFOLDER="leveldbtest"

type Song struct {
	title string
}

func (self *Song) To(buff *bytes.Buffer) error {
	buff.WriteString(self.title)
	return nil
}

func (self *Song) From(buff *bytes.Buffer) error {
	self.title,_ = buff.ReadString(0)
	return nil
}

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
	sIn := Song{title:"Waylon Jennings Live"}
	err := table.Put([]byte("The Mountain Goats"), &sIn)
	if err  != nil {
		t.Error("Failed to put entry",err)
	}
	var sOut Song
	err = table.Get([]byte("The Mountain Goats"),&sOut)
	if  err != nil {
		t.Error("Failed to get entry", err)
	}
	if sOut.title != "Waylon Jennings Live" {
		t.Error("Failed to get entry", sOut.title)
	}
}

func TestDropTable(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	table,_ := db.CreateTable("songs")
	sIn := Song{title:"Waylon Jennings Live"}
	table.Put([]byte("The Mountain Goats"), &sIn)
	db.CloseTable("songs")

	err := db.DropTable("labels")
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
	sIn := Song{title:"Waylon Jennings Live"}
	err := table.Put([]byte("The Mountain Goats"), &sIn)
	if err  != nil {
		t.Error("Failed to put entry",err)
	}
	var sOut Song
	err = table.Get([]byte("The Mountain Goats"),&sOut)
	if  err != nil {
		t.Error("Failed to get entry", err)
	}
	if sOut.title != "Waylon Jennings Live" {
		t.Error("Failed to get entry", sOut.title)
	}
	err = table.Delete([]byte("The Mountain Goats"))
	if err != nil {
		t.Error("Failed to delete entry", err)
	}
	sOut.title=""
	err = table.Get([]byte("The Mountain Goats"),&sOut)
	if  err == nil {
		t.Error("Succeded in geting deleted entry", err)
	}
	if sOut.title != "" {
		t.Error("Value returned for a deleted entry", sOut.title)
	}
}

func TestIterate(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	table,_ := db.CreateTable("songs")
	songlist:= map[string]Song{
		"The Kinks": Song{title:"Sitting in My Hotel"},
		"Sacred Paws": Song{title:"Wet Graffiti"},
		"The Go-betweens": Song{title:"Born to a Family"},
	}
	for artist, song := range(songlist) {
		table.Put([]byte(artist), &song)
	}
	songlistretrieved := map[string]Song{}
	iterator,err := table.NewIterator(nil)
	if err != nil {
		t.Error("Failed to create Iterator", err)
	}
	var sOut Song
	for iterator.Next() {
		iterator.Value(&sOut)
		songlistretrieved[string(iterator.Key())] = sOut
	}
	if !reflect.DeepEqual(songlistretrieved,songlist) {
		t.Error("Retrieved song list doesnt match", songlist, songlistretrieved)
	}
}

func TestIterateSnapshot(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	table,_ := db.CreateTable("songs")
	songlist:= map[string]Song{
		"The Kinks": Song{title:"Sitting in My Hotel"},
		"Sacred Paws": Song{title:"Wet Graffiti"},
		"The Go-betweens": Song{title:"Born to a Family"},
	}
	for artist, song := range(songlist) {
		table.Put([]byte(artist), &song)
	}

	songlistretrieved := map[string]Song{}
	ss,err := table.TakeSnapshot()
	if err != nil {
		t.Error("Failed to create snapshot")
	}
	iterator,err := ss.NewIterator(nil)
	if err != nil {
		t.Error("Failed to create iterator from snapshot", err)
	}
	var sOut Song
	for iterator.Next() {
		iterator.Value(&sOut)
		songlistretrieved[string(iterator.Key())] = sOut
	}
	if !reflect.DeepEqual(songlistretrieved,songlist) {
		t.Error("Retrieved song list doesnt match", songlist, songlistretrieved)
	}
}

func TestIteratePrefix(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	table,_ := db.CreateTable("songs")
	songlist:= map[string]Song{
		"The Kinks": Song{title:"Sitting in My Hotel"},
		"Sacred Paws": Song{title:"Wet Graffiti"},
		"The Go-betweens": Song{title:"Born to a Family"},
	}
	for artist, song := range(songlist) {
		table.Put([]byte(artist), &song)
	}

	songlistretrieved := map[string]Song{}
	iterator,err := table.NewIterator([]byte("The"))
	if err != nil {
		t.Error("Failed to create Iterator", err)
	}
	var sOut Song
	for iterator.Next() {
		iterator.Value(&sOut)
		songlistretrieved[string(iterator.Key())] = sOut
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
	songlist:= map[string]Song{
		"The Kinks": Song{title:"Sitting in My Hotel"},
		"Sacred Paws": Song{title:"Wet Graffiti"},
		"The Go-betweens": Song{title:"Born to a Family"},
	}
	bw,err := table.MakeBatch(2)
	if err != nil {
		t.Error("Failed to create BatchWrite")
	}
	for artist, song := range(songlist) {
		table.Put([]byte(artist), &song)
		if err != nil {
			t.Error("BatchWrite put failed with error",err)
		}
	}
	bw.Release()
	songlistretrieved := map[string]Song{}
	iterator,err := table.NewIterator(nil)
	var sOut Song
	for iterator.Next() {
		iterator.Value(&sOut)
		songlistretrieved[string(iterator.Key())] = sOut
	}
	if !reflect.DeepEqual(songlistretrieved,songlist) {
		t.Error("BatchWrite didnt write full song list", songlist, songlistretrieved)
	}
}


