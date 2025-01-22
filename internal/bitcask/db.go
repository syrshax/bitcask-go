package bitcask

import (
	"fmt"
	"os"
	"path/filepath"
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
		return nil, err
	}

	return &Bitcask{
		directory: directory,
		lockFile:  lockFile,
	}, nil
}

func (b *Bitcask) Close() error {
	if b.lockFile != nil {
		return b.lockFile.Close()
	}
	return nil
}
