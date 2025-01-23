package bitcask

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const MaxFileSize = 4 * 1024 * 1024 //thats 4MB as paper tells

type Bitcask struct {
	directory  string
	lockFile   *os.File
	activeFile *os.File
	fileID     uint64
}

type OpenOptions struct {
	ReadOnly bool
}

// EntryLog represents the on-disk format of a key-value pair
//
// Format according to paper:
//
// +-------+------------+---------+-----------+---------+-----------+
//
// |  CRC  | Timestamp | KeySize | ValueSize |   Key   |   Value   |
//
// +-------+------------+---------+-----------+---------+-----------+
//
// |  4B   |    8B     |   4B    |    4B     | KeySize | ValueSize|
type EntryLog struct {
	crc    uint32
	tstamp uint64
	ksz    uint32
	vsz    uint32
	key    []byte
	value  []byte
}

func (e *EntryLog) calculateCrc() uint32 {
	crc := crc32.NewIEEE()
	binary.Write(crc, binary.LittleEndian, e.tstamp)
	binary.Write(crc, binary.LittleEndian, e.ksz)
	binary.Write(crc, binary.LittleEndian, e.vsz)
	crc.Write(e.key)
	crc.Write(e.value)

	return crc.Sum32()
}

func (e *EntryLog) Encode() []byte {
	bufferSize := 4 + 8 + 4 + 4 + e.ksz + e.vsz
	buffer := make([]byte, bufferSize)

	pointer := 0
	binary.LittleEndian.PutUint32(buffer[pointer:], e.crc)
	pointer += 4
	binary.LittleEndian.PutUint64(buffer[pointer:], e.tstamp)
	pointer += 8
	binary.LittleEndian.PutUint32(buffer[pointer:], (e.ksz))
	pointer += 4
	binary.LittleEndian.PutUint32(buffer[pointer:], (e.vsz))
	pointer += 4
	copy(buffer[pointer:], e.key)
	pointer += len(e.key)
	copy(buffer[pointer:], e.value)

	return buffer

}

func (b *Bitcask) Put(key, value []byte) error {
	if len(key) == 0 {
		return fmt.Errorf("key cannot be empty")
	}
	if len(value) == 0 {
		return fmt.Errorf("value cannot be empty")
	}

	entry := NewEntryLog(key, value)
	encodedEntryLog := entry.Encode()

	_, err := b.activeFile.Write(encodedEntryLog)
	if err != nil {
		return fmt.Errorf("failed to write entry: %w", err)
	}

	err = b.activeFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	return nil
}

func NewEntryLog(key, value []byte) *EntryLog {
	e := &EntryLog{
		tstamp: uint64(time.Now().UnixNano()),
		ksz:    uint32(len(key)),
		vsz:    uint32(len(value)),
		key:    key,
		value:  value,
	}
	e.crc = e.calculateCrc()
	return e
}

func CreateActiveFile(directory string) (*os.File, uint64, error) {
	fileID := uint64(time.Now().UnixNano())

	filename := filepath.Join(directory, fmt.Sprintf("%d.active.log", fileID))

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, 0, err
	}

	return file, fileID, nil
}

func Open(directory string, options OpenOptions) (*Bitcask, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, err
	}

	lockPath := filepath.Join(directory, "db.lock")
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("couldn't create lock file: %w", err)
	}

	err = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		lockFile.Close()
		return nil, fmt.Errorf("database is locked by another process: %w", err)
	}

	activeFile, fileID, err := CreateActiveFile(directory)
	if err != nil {
		syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
		lockFile.Close()
		return nil, err
	}

	return &Bitcask{
		directory:  directory,
		lockFile:   lockFile,
		activeFile: activeFile,
		fileID:     fileID,
	}, nil
}

func (b *Bitcask) Close() error {
	if b.lockFile != nil {
		return b.lockFile.Close()
	}
	return nil
}
