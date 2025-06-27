package db_test

import (
	db "bitcask/internal"
	"bytes"
	"testing"
)

func TestPutAndGet(t *testing.T) {
	tmpDir := t.TempDir()
	bitcask, err := db.Open(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open: %v", err)
	}
	defer bitcask.Close()

	key1 := []byte("Cat")
	val1 := []byte("Black")

	err = bitcask.Put(key1, val1)
	if err != nil {
		t.Fatalf("Failed to put value: %v", val1)
	}

	retrive, err := bitcask.Get(key1)
	if err != nil {
		t.Fatalf("Failed to get: %v", err)
	}

	if !bytes.Equal(val1, retrive) {
		t.Errorf("Failed to retrieve values: %v, Expected: %v", val1, retrive)
	}

}

func TestReopenDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	bdb, err := db.Open(tmpDir)
	if err != nil {
		t.Fatalf("Failed to open: %v", err)
	}

	key := []byte("persistence")
	val := []byte("testing")

	bdb.Put(key, val)

	bdb.Close()

	bdb, err = db.Open(tmpDir)
	if err != nil {
		t.Fatal("error on opening")
	}
	defer bdb.Close()

	data, err := bdb.Get(key)
	if err != nil {
		t.Fatalf("failed to get key")
	}

	if !bytes.Equal(data, val) {
		t.Errorf("Failed to retrieve values: %v, Expected: %v", data, val)
	}
}
