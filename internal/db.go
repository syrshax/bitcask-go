package db

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const HEADER_SIZE int64 = 16
const MAX_FILE_SIZE int64 = 4096

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
	Keydir       map[string]IndexEntry
	activeFile   *os.File
	activeFileId int
	dir          string
	lock         LockFile
	maxFileSize  int64
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

	fileSize := len(value) + int(writeOffset)
	if int64(fileSize) > b.maxFileSize {
		log.Println("Max File Exceded! Rotation Should Happen")
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

	offset := writeOffset + HEADER_SIZE + int64(bitcaskFile.ksz)
	newIndexEntry := IndexEntry{
		filename: b.activeFile.Name(),
		offset:   offset,
		size:     int64(bitcaskFile.value_sz),
	}

	b.Keydir[string(key)] = newIndexEntry
	return nil
}

func Open(d string, maxFileSize int64) (*Bitcask, error) {
	err := os.MkdirAll(d, 0775)
	if err != nil {
		return nil, err
	}

	dirItems, err := os.ReadDir(d)
	if err != nil {
		return nil, err
	}

	maxId := 1
	for _, d := range dirItems {
		if strings.HasSuffix(d.Name(), ".data") {
			idStr := strings.TrimSuffix(d.Name(), ".data")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				continue
			}
			if id > maxId {
				maxId = id
			}
		}
	}

	fileName := fmt.Sprintf("%08d.data", maxId)
	fileDir := filepath.Join(d, fileName)
	file, err := os.OpenFile(fileDir, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, fmt.Errorf("bitcask: failed to open data file %s: %w", fileDir, err)
	}

	b := &Bitcask{
		activeFile:   file,
		activeFileId: maxId,
		maxFileSize:  maxFileSize,
		dir:          d,
		Keydir:       map[string]IndexEntry{},
		lock:         LockFile{},
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
	keyDirMap := make(map[string]IndexEntry)

	dirEntries, err := os.ReadDir(b.dir)
	if err != nil {
		return err
	}

	var dataFilePaths []string
	for _, entry := range dirEntries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".data") {
			dataFilePaths = append(dataFilePaths, entry.Name())
		}
	}

	sort.Slice(dataFilePaths, func(i, j int) bool {
		idA, _ := strconv.Atoi(strings.TrimSuffix(dataFilePaths[i], ".data"))
		idB, _ := strconv.Atoi(strings.TrimSuffix(dataFilePaths[j], ".data"))
		return idA < idB
	})

	for _, fileName := range dataFilePaths {
		filePath := filepath.Join(b.dir, fileName)

		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		var offset int64 = 0
		for {
			headerBuf := make([]byte, HEADER_SIZE)
			_, err := io.ReadFull(file, headerBuf)
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			crc := binary.LittleEndian.Uint32(headerBuf[0:4])
			ksz := binary.LittleEndian.Uint32(headerBuf[8:12])
			vsz := binary.LittleEndian.Uint32(headerBuf[12:16])

			key := make([]byte, ksz)
			if _, err := io.ReadFull(file, key); err != nil {
				return err
			}
			value := make([]byte, vsz)
			if _, err := io.ReadFull(file, value); err != nil {
				return err
			}

			checksum := crc32.NewIEEE()
			checksum.Write(key)
			checksum.Write(value)
			if checksum.Sum32() != crc {
				return errors.New("database corruption: checksum mismatch in file " + fileName)
			}

			entry := IndexEntry{
				filename: fileName,
				offset:   offset + HEADER_SIZE + int64(ksz),
				size:     int64(vsz),
			}
			keyDirMap[string(key)] = entry

			offset += HEADER_SIZE + int64(ksz) + int64(vsz)
		}
	}

	b.Keydir = keyDirMap
	return nil
}
