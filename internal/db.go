package db

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"time"
)

const HEADER_SIZE int64 = 16

var ErrKeyNotFound = errors.New("bitcask: key not found")

type BitcaskFile struct {
	crc      uint32
	tstamp   int32
	ksz      int32
	value_sz int32
	key      []byte
	value    []byte
}

type LockFile struct{}

type IndexEntry struct {
	filename string
	offset   int64
	size     int64
}

type Bitcask struct {
	activeFile *os.File
	path       string
	Keydir     map[string]IndexEntry
	lock       LockFile
}

func NewBitcaskFile(k, val []byte) (*BitcaskFile, error) {
	b := &BitcaskFile{}

	b.ksz = int32(len(k))
	b.value_sz = int32(len(val))
	b.tstamp = int32(time.Now().UnixMilli())
	b.key = k
	b.value = val

	if err := b.CRCChecksum(); err != nil {
		return nil, err
	}
	return b, nil
}

func (bf *BitcaskFile) Encode() []byte {
	size := HEADER_SIZE + int64(bf.ksz) + int64(bf.value_sz)
	buf := make([]byte, size)

	binary.LittleEndian.PutUint32(buf[0:4], bf.crc)
	binary.LittleEndian.PutUint32(buf[4:8], uint32(bf.tstamp))
	binary.LittleEndian.PutUint32(buf[8:12], uint32(bf.ksz))
	binary.LittleEndian.PutUint32(buf[12:16], uint32(bf.value_sz))

	copy(buf[16:], bf.key)
	copy(buf[16+int64(bf.ksz):], bf.value)

	return buf
}

func (b *Bitcask) Get(key []byte) ([]byte, error) {
	entry, ok := b.Keydir[string(key)]
	if !ok {
		return nil, ErrKeyNotFound
	}

	value := make([]byte, entry.size)

	_, err := b.activeFile.ReadAt(value, entry.offset)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (b *Bitcask) Put(key, value []byte) error {
	writeOffset, err := b.activeFile.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("bitcask: failed to get write offset: %w", err)
	}

	bitcaskFile, err := NewBitcaskFile(key, value)
	if err != nil {
		return fmt.Errorf("bitcask: failed to get make the CRC: %w", err)
	}
	record := bitcaskFile.Encode()

	_, err = b.activeFile.Write(record)
	if err != nil {
		return fmt.Errorf("bitcask: failed to write record: %w", err)
	}

	newIndexEntry := IndexEntry{
		filename: b.activeFile.Name(),
		offset:   writeOffset + HEADER_SIZE + int64(bitcaskFile.ksz),
		size:     int64(bitcaskFile.value_sz),
	}

	b.Keydir[string(key)] = newIndexEntry
	return nil
}

func Open(d string) (*Bitcask, error) {
	err := os.MkdirAll(d, 0775)
	if err != nil {
		return nil, err
	}

	fileDir := filepath.Join(d, "0001.data")
	file, err := os.OpenFile(fileDir, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("bitcask: failed to open data file %s: %w", fileDir, err)
	}

	b := &Bitcask{
		activeFile: file,
		path:       fileDir,
		Keydir:     map[string]IndexEntry{},
		lock:       LockFile{},
	}

	err = b.loadIndex()
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Bitcask) Close() error {
	return b.activeFile.Close()
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

func (b *Bitcask) loadIndex() error {
	_, err := b.activeFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	buf := make([]byte, HEADER_SIZE)

	var idx int64 = 0
	keyDirMap := map[string]IndexEntry{}

	for {
		_, err := io.ReadFull(b.activeFile, buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		crc := binary.LittleEndian.Uint32(buf[0:4])
		ksz := binary.LittleEndian.Uint32(buf[8:12])
		value_sz := binary.LittleEndian.Uint32(buf[12:16])

		key := make([]byte, ksz)
		if _, err := io.ReadFull(b.activeFile, key); err != nil {
			return err
		}

		value := make([]byte, value_sz)
		if _, err := io.ReadFull(b.activeFile, value); err != nil {
			return err
		}

		checksum := crc32.NewIEEE()
		checksum.Write(key)
		checksum.Write(value)

		if checksum.Sum32() != crc {
			return errors.New("database corruption: checksum mismatch")
		}

		offset := idx + HEADER_SIZE + int64(ksz)

		entry := IndexEntry{
			filename: b.activeFile.Name(),
			offset:   int64(offset),
			size:     int64(value_sz),
		}

		idx = idx + HEADER_SIZE + int64(ksz) + int64(value_sz)
		keyDirMap[string(key)] = entry
	}

	b.Keydir = keyDirMap

	return nil
}
