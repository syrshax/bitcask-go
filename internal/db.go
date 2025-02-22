package db

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"time"
)

type BitcaskFile struct {
	crc      uint32
	tstamp   int32
	ksz      int32
	value_sz int32
	key      []byte
	value    []byte
}

type LockFile struct{}

type Indexing struct{}

type Bitcask struct {
	file     BitcaskFile
	indexing Indexing
	lock     LockFile
}

func (b *BitcaskFile) CRCChecksum() error {
	crc := crc32.NewIEEE()
	_, err := crc.Write(b.key)
	if err != nil {
		return err
	}
	_, err = crc.Write(b.value)
	if err != nil {
		return err
	}
	b.crc = crc.Sum32()
	return nil
}

func (b *BitcaskFile) CreateAppend(k []byte, val []byte) []byte {
	b.ksz = int32(len(k))
	b.value_sz = int32(len(val))
	b.tstamp = int32(time.Now().UnixMilli())
	b.key = k
	b.value = val
	b.CRCChecksum()

	buffer := make([]byte, 0)

	crcBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(crcBytes, b.crc)
	fmt.Println(crcBytes)

	crcBytes[0] = byte(b.crc)
	crcBytes[1] = byte(b.crc >> 8)
	crcBytes[2] = byte(b.crc >> 16)
	crcBytes[3] = byte(b.crc >> 24)
	buffer = append(buffer, crcBytes...)

	tstampBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(tstampBytes, uint32(b.tstamp))
	buffer = append(buffer, tstampBytes...)

	kszBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(kszBytes, uint32(b.ksz))
	buffer = append(buffer, kszBytes...)

	valueSzBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(valueSzBytes, uint32(b.value_sz))
	buffer = append(buffer, valueSzBytes...)

	buffer = append(buffer, b.key...)
	buffer = append(buffer, b.value...)
	fmt.Println(buffer)
	return buffer
}

func (b *BitcaskFile) Write(k []byte, val []byte) {
}
