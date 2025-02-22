package db

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"testing"
)

func TestCRCChecksum(t *testing.T) {
	key := []byte("hello")
	value := []byte("world")

	expectedCRC := func() uint32 {
		crc := crc32.NewIEEE()
		crc.Write(key)
		crc.Write(value)
		return crc.Sum32()
	}()

	bf := BitcaskFile{
		key:   key,
		value: value,
	}

	err := bf.CRCChecksum()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bf.crc != expectedCRC {
		t.Errorf("CRC mismatch: got %x, expected %x", bf.crc, expectedCRC)
	}
}

func TestBitcaskFile_CreateAppend(t *testing.T) {
	key := []byte("testKey")
	value := []byte("testValue")

	bf := &BitcaskFile{}

	result := bf.CreateAppend(key, value)

	if len(result) == 0 {
		t.Error("Append produced an empty result")
	}

	expectedCRC := func() uint32 {
		crc := crc32.NewIEEE()
		crc.Write(key)
		crc.Write(value)
		return crc.Sum32()
	}()

	extractedCRC := binary.LittleEndian.Uint32(result[:4])

	if extractedCRC != expectedCRC {
		t.Errorf("CRC mismatch: got %d, expected %d", extractedCRC, expectedCRC)
	}

	if !bytes.Contains(result, key) {
		t.Error("Key not found in appended data")
	}
	if !bytes.Contains(result, value) {
		t.Error("Value not found in appended data")
	}

	if bf.ksz != int32(len(key)) {
		t.Errorf("key size mismatch: got %d, expected %d", bf.ksz, len(key))
	}

	if bf.value_sz != int32(len(value)) {
		t.Errorf("value size mismatch: got %d, expected %d", bf.value_sz, len(value))
	}

}
