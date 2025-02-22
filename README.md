# Bitcask in Go: An Ongoing Implementation

This repository contains an ongoing implementation of a Bitcask key-value store in Go, following the principles outlined in the original Bitcask paper.

## Current Status

This project is currently in active development. The core data structure and append-only writing functionality are being implemented.

**What's Implemented:**

* **`BitcaskFile` Structure:** Defines the record format for data storage, including CRC checksum, timestamp, key and value sizes, and the key-value pair itself.
* **`CRCChecksum()`:** Calculates and sets the CRC32 checksum for a record, ensuring data integrity.
* **`CreateAppend()`:** Serializes a key-value pair into a byte slice, including metadata, for append-only writing to the data file.
* **Basic Data Serialization:** The data is being serialized to a byte array.
* **Basic Project Structure:** The project is structured with internal packages for database logic.

**What's Missing:**

* **File I/O:** Actual reading and writing to disk.
* **Indexing:** Implementation of the in-memory key directory for fast lookups.
* **Locking:** Mechanisms to prevent concurrent access conflicts.
* **Merging:** Log compaction and merging of data files.
* **Reading:** Functionality to retrieve values based on keys.
* **Error Handling:** Robust error handling throughout the codebase.
* **Testing:** Comprehensive unit and integration tests.
* **Open File and Create DB functions:** The main.go file is only printing to the console.


## Getting Started

1.  **Clone the repository:**

    ```bash
    git clone [repository URL]
    cd bitcask
    ```

2.  **Run the application:**

    ```bash
    go run main.go
    ```

    Currently, this only displays a menu of unimplemented options.

## Current Code Snippets

```go
package db

import (
        "encoding/binary"
        "fmt"
        "hash/crc32"
        "time"
)

// ... (BitcaskFile, LockFile, Indexing, Bitcask structs) ...

func (b *BitcaskFile) CRCChecksum() error {
        // ... (CRC calculation) ...
}

func (b *BitcaskFile) CreateAppend(k []byte, val []byte) []byte {
        // ... (Serialization of record data) ...
}

func (b *BitcaskFile) Write(k []byte, val []byte) {
        // ... (To be implemented) ...
}

