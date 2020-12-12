package db

import (
	"testing"
	"reflect"
	"bytes"
)

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

type Character struct {
	char  string
}

func (self *Character) To(buff *bytes.Buffer) error {
	buff.WriteString(self.char)
	return nil
}

func (self *Character) From(buff *bytes.Buffer) error {
	self.char,_ = buff.ReadString(0)
	return nil
}

func dotestCreateTable(db Database,t *testing.T) {
	_,err := db.CreateTable("bands")
	if err != nil {
		t.Error("Create failed",err)
	}
	_,err = db.CreateTable("bands")
	if err == nil {
		t.Error("Creation of table that already exists succeeded")
	}
}

func dotestOpenTable(db Database,t *testing.T) {
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

func dotestPutGet(db Database,t *testing.T) {
	table,_ := db.CreateTable("songs")
	sIn := Song{title:"Waylon Jennings Live"}
	err := table.Put("The Mountain Goats", &sIn)
	if err  != nil {
		t.Error("Failed to put entry",err)
	}
	var sOut Song
	err = table.Get("The Mountain Goats",&sOut)
	if  err != nil {
		t.Error("Failed to get entry", err)
	}
	if sOut.title != "Waylon Jennings Live" {
		t.Error("Failed to get entry", sOut.title)
	}
}

func dotestDropTable(db Database,t *testing.T) {
	table,_ := db.CreateTable("songs")
	sIn := Song{title:"Waylon Jennings Live"}
	table.Put("The Mountain Goats", &sIn)
	db.CloseTable("songs")

	err := db.DropTable("labels")
	if err == nil {
		t.Error("Dropped non-existent table")
	}

	err = db.DropTable("songs")
	if err != nil {
		t.Error("Failed to drop table",err)
	}
}

func dotestDelete(db Database,t *testing.T) {
	table,_ := db.CreateTable("songs")
	sIn := Song{title:"Waylon Jennings Live"}
	err := table.Put("The Mountain Goats", &sIn)
	if err  != nil {
		t.Error("Failed to put entry",err)
	}
	var sOut Song
	err = table.Get("The Mountain Goats",&sOut)
	if  err != nil {
		t.Error("Failed to get entry", err)
	}
	if sOut.title != "Waylon Jennings Live" {
		t.Error("Failed to get entry", sOut.title)
	}
	err = table.Delete("The Mountain Goats")
	if err != nil {
		t.Error("Failed to delete entry", err)
	}
	sOut.title=""
	err = table.Get("The Mountain Goats",&sOut)
	if  err == nil {
		t.Error("Succeded in geting deleted entry", err)
	}
	if sOut.title != "" {
		t.Error("Value returned for a deleted entry", sOut.title)
	}
}

func dotestIterate(db Database,t *testing.T) {
	table,_ := db.CreateTable("songs")
	songlist:= map[string]Song{
		"The Kinks": Song{title:"Sitting in My Hotel"},
		"Sacred Paws": Song{title:"Wet Graffiti"},
		"The Go-betweens": Song{title:"Born to a Family"},
	}
	for artist, song := range(songlist) {
		table.Put(artist, &song)
	}
	songlistretrieved := map[string]Song{}
	iterator,err := table.NewIterator("")
	if err != nil {
		t.Error("Failed to create Iterator", err)
	}
	var sOut Song
	for iterator.Next() {
		iterator.Value(&sOut)
		songlistretrieved[iterator.Key()] = sOut
	}
	if !reflect.DeepEqual(songlistretrieved,songlist) {
		t.Error("Retrieved song list doesnt match", songlist, songlistretrieved)
	}
} 

func dotestIterateSnapshotASCII(db Database,t *testing.T) {
	table,_ := db.CreateTable("characters")
	for i :=0; i < 256; i++ {
		v := Character{char:string(rune(i))}
		table.Put(string(rune(i)), &v)
	}
	ss,err := table.TakeSnapshot()
	if err != nil {
		t.Error("Failed to create snapshot")
	}
	for i :=0; i < 256; i++ {
		iterator,err := ss.NewIterator(string(rune(i)))
		if err != nil {
			t.Error("Failed to create iterator from snapshot", err)
		}
		var sOut Character
		if !iterator.Next(){
			t.Error("Failed to iterate single value")
		}
		iterator.Value(&sOut)
		if sOut.char !=string(rune(i)){
			t.Error("Failed to get expected value",sOut,rune(i))
		}
		if iterator.Next() {
			t.Error("Succeeded in iterating past single value")
		}
		if iterator.Error() != nil {
			t.Error("Reporting error at end of successful iteration")
		}
	}
}

func dotestIterateSnapshot(db Database,t *testing.T) {
	table,_ := db.CreateTable("songs")
	songlist:= map[string]Song{
		"The Kinks": Song{title:"Sitting in My Hotel"},
		"Sacred Paws": Song{title:"Wet Graffiti"},
		"The Go-betweens": Song{title:"Born to a Family"},
	}
	for artist, song := range(songlist) {
		table.Put(artist, &song)
	}

	songlistretrieved := map[string]Song{}
	ss,err := table.TakeSnapshot()
	if err != nil {
		t.Error("Failed to create snapshot")
	}
	iterator,err := ss.NewIterator("")
	if err != nil {
		t.Error("Failed to create iterator from snapshot", err)
	}
	var sOut Song
	for iterator.Next() {
		iterator.Value(&sOut)
		songlistretrieved[iterator.Key()] = sOut
	}
	if iterator.Error() != nil {
		t.Error("Reporting error at end of successful iteration")
	}
	if !reflect.DeepEqual(songlistretrieved,songlist) {
		t.Error("Retrieved song list doesnt match", songlist, songlistretrieved)
	}
}

func dotestIterateSnapshotPrefixEmpty(db Database,t *testing.T) {
	table,_ := db.CreateTable("songs")
	songlist:= map[string]Song{
		"The Kinks": Song{title:"Sitting in My Hotel"},
		"Sacred Paws": Song{title:"Wet Graffiti"},
		"The Go-betweens": Song{title:"Born to a Family"},
	}
	for artist, song := range(songlist) {
		table.Put(artist, &song)
	}
	ss,err := table.TakeSnapshot()
	if err != nil {
		t.Error("Failed to create snapshot")
	}
	iterator,err := ss.NewIterator("A")
	if err != nil {
		t.Error("Failed to create Iterator", err)
	}
	if iterator.Next() {
		t.Error("Next returning true for emtpy iterator")
	}	
	if iterator.Error() != nil {
		t.Error("Reporting error at end of successful iteration")
	}

}

func dotestIteratePrefix(db Database,t *testing.T) {
	table,_ := db.CreateTable("songs")
	songlist:= map[string]Song{
		"The Kinks": Song{title:"Sitting in My Hotel"},
		"Sacred Paws": Song{title:"Wet Graffiti"},
		"The Go-betweens": Song{title:"Born to a Family"},
	}
	for artist, song := range(songlist) {
		table.Put(artist, &song)
	}

	songlistretrieved := map[string]Song{}
	iterator,err := table.NewIterator("The")
	if err != nil {
		t.Error("Failed to create Iterator", err)
	}
	var sOut Song
	for iterator.Next() {
		iterator.Value(&sOut)
		songlistretrieved[iterator.Key()] = sOut
	}
	if iterator.Error() != nil {
		t.Error("Reporting error at end of successful iteration")
	}

	if reflect.DeepEqual(songlistretrieved,songlist) {
		t.Error("Retrieved song list matches", songlist, songlistretrieved)
	}
	delete(songlist,"Sacred Paws")
	if !reflect.DeepEqual(songlistretrieved,songlist) {
		t.Error("Retrieved song lists dont match", songlist, songlistretrieved)
	}
}

func dotestIteratePrefixEmpty(db Database,t *testing.T) {
	table,_ := db.CreateTable("songs")
	songlist:= map[string]Song{
		"The Kinks": Song{title:"Sitting in My Hotel"},
		"Sacred Paws": Song{title:"Wet Graffiti"},
		"The Go-betweens": Song{title:"Born to a Family"},
	}
	for artist, song := range(songlist) {
		table.Put(artist, &song)
	}
	iterator,err := table.NewIterator("A")
	if err != nil {
		t.Error("Failed to create Iterator", err)
	}
	if iterator.Next() {
		t.Error("Next returning true for emtpy iterator")
	}
}

func dotestBatchWrite(db Database,t *testing.T) {
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
		table.Put(artist, &song)
		if err != nil {
			t.Error("BatchWrite put failed with error",err)
		}
	}
	bw.Release()
	songlistretrieved := map[string]Song{}
	iterator,err := table.NewIterator("")
	var sOut Song
	for iterator.Next() {
		iterator.Value(&sOut)
		songlistretrieved[iterator.Key()] = sOut
	}
	if !reflect.DeepEqual(songlistretrieved,songlist) {
		t.Error("BatchWrite didnt write full song list", songlist, songlistretrieved)
	}
}


