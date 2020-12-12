package db

import (
	"testing"
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

func TestDatastoreCreateTable(t *testing.T) {
	db := setupDatastore(t)
	if db == nil {
		return
	}

	defer teardownDatastore(db)
	dotestCreateTable(db,t)
}

func TestDatastorePutGet(t *testing.T) {
	db := setupDatastore(t)
	if db == nil {
		return
	}

	defer teardownDatastore(db)
	dotestPutGet(db,t)
}

func TestDatastoreDropTable(t *testing.T) {
	db := setupDatastore(t)
	if db == nil {
		return
	}

	defer teardownDatastore(db)
	dotestDropTable(db,t)
}

func TestDatastoreDelete(t *testing.T) {
	db := setupDatastore(t)
	if db == nil {
		return
	}

	defer teardownDatastore(db)
	dotestDelete(db,t) 
}

func TestDatastoreIterate(t *testing.T) {
	db := setupDatastore(t)
	if db == nil {
		return
	}

	defer teardownDatastore(db)
	dotestIterate(db,t)
}

func TestDatastoreIteratePrefixEmpty(t *testing.T) {
	db := setupDatastore(t)
	if db == nil {
		return
	}

	defer teardownDatastore(db)
	dotestIteratePrefixEmpty(db,t)
}

func TestDatastoreIterateSnapshot(t *testing.T) {
	db := setupDatastore(t)
	if db == nil {
		return
	}

	defer teardownDatastore(db)
	dotestIterateSnapshot(db,t)
}

func TestDatastoreIterateSnapshotPrefixEmpty(t *testing.T) {
	db := setupDatastore(t)
	if db == nil {
		return
	}

	defer teardownDatastore(db)
	dotestIterateSnapshotPrefixEmpty(db,t)
}

func TestDatastoreIterateSnapshotASCII(t *testing.T) {
	db := setupDatastore(t)
	if db == nil {
		return
	}

	defer teardownDatastore(db)
	dotestIterateSnapshotASCII(db,t)
}

func TestDatastoreIteratePrefix(t *testing.T) {
	db := setupDatastore(t)
	if db == nil {
		return
	}

	defer teardownDatastore(db)
	dotestIteratePrefix(db,t)
}

func TestDatastoreBatchWrite(t *testing.T) {
	db := setupDatastore(t)
	if db == nil {
		return
	}

	defer teardownDatastore(db)
	dotestBatchWrite(db,t)
}

