package bitcask

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestBitcaskInstanceCreation(t *testing.T) {
	testDir, err := os.MkdirTemp("", "bitcask-test-*")
	if err != nil {
		t.Fatalf("Couldn't create temp directory, %v", err)
	}
	defer os.RemoveAll(testDir)

	db, err := Open(testDir, OpenOptions{})
	if err != nil {
		t.Fatalf("Failed to open Bitcask, %v", err)
	}
	if _, err := os.Stat(testDir + "/db.lock"); err != nil {
		t.Errorf("Lock file coulnt be created, %v", err)
	}

	if err := db.Close(); err != nil {
		t.Errorf("Error closing the DB instance, %v:", err)
	}
}

func TestCreateActiveFile(t *testing.T) {
	// Create temp directory
	testDir, err := os.MkdirTemp("", "bitcask-test-*")
	if err != nil {
		t.Fatalf("Couldn't create temp directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Create active file
	file, fileID, err := CreateActiveFile(testDir)
	if err != nil {
		t.Fatalf("Failed to create active file: %v", err)
	}
	defer file.Close()

	// Test 1: Verify file exists with correct name pattern
	expectedPath := filepath.Join(testDir, fmt.Sprintf("%d.active.log", fileID))
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("Active file was not created at expected path: %s", expectedPath)
	}

	// Test 2: Verify file is writable
	testData := []byte("test data")
	_, err = file.Write(testData)
	if err != nil {
		t.Errorf("Could not write to active file: %v", err)
	}

	// Test 3: Verify fileID (timestamp) is recent
	now := uint64(time.Now().UnixNano())
	if fileID > now {
		t.Errorf("FileID is in the future. FileID: %d, Now: %d", fileID, now)
	}
	if now-fileID > uint64(time.Second) {
		t.Errorf("FileID is too old. FileID: %d, Now: %d", fileID, now)
	}

	// Test 4: Verify file permissions (accounting for umask)
	info, err := os.Stat(expectedPath)
	if err != nil {
		t.Fatalf("Could not stat file: %v", err)
	}
	if !info.Mode().IsRegular() {
		t.Error("Not a regular file")
	}
	perm := info.Mode().Perm()
	if perm&0600 == 0 { // Check owner read/write
		t.Error("File should be readable/writable by owner")
	}
	if perm&0060 == 0 { // Check group read/write
		t.Error("File should be readable/writable by group")
	}
}
