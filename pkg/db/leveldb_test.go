package db

import (
	"testing"
	"os"
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
	dotestCreateTable(db,t)
}

func TestOpenTable(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	dotestOpenTable(db,t)
}

func TestPutGet(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	dotestPutGet(db,t)
}

func TestDropTable(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	dotestDropTable(db,t)
	tablepath := filepath.Join(LEVELDBFOLDER,"songs")
	_, err := os.Stat(LEVELDBFOLDER)
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
	dotestDelete(db,t) 
}

func TestIterate(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	dotestIterate(db,t)
}

func TestIterateSnapshot(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	dotestIterateSnapshot(db,t)
}

func TestIterateSnapshotPrefixEmpty(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	dotestIterateSnapshotPrefixEmpty(db,t)
}
func TestIterateSnapshotASCII(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	dotestIterateSnapshotASCII(db,t)
}

func TestIteratePrefix(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	dotestIteratePrefix(db,t)
}

func TestIteratePrefixEmpty(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	dotestIteratePrefixEmpty(db,t)
}

func TestBatchWrite(t *testing.T) {
	db := NewLevelDB(LEVELDBFOLDER)
	defer teardown(db)
	dotestBatchWrite(db,t)
}

