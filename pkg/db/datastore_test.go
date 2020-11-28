package db

import (
	"testing"
	//"os"
	//"reflect"
	//"path/filepath"
	//"bytes"
)

func setupDatastore(t *testing.T) *DatastoreDB {
	db := NewDatastoreDB("flaptest")
	if db == nil {
		t.Error("Failed to create db object")
	}
	return db
}

func teardownDatastore(db *DatastoreDB) {
	db.Release()
}

func TestDatastorePutGet(t *testing.T) {
	db := setupDatastore(t)
	if db == nil {
		return
	}
	defer teardownDatastore(db)
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

