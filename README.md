# Bitcask Implementation in Go

A Go implementation of the Bitcask key-value store based on the [Bitcask paper](https://riak.com/assets/bitcask-intro.pdf).

## Current Implementation Status

- ✅ Basic database creation and file management
- ✅ Single-writer locking mechanism using `db.lock`
- ✅ Active file creation and management
- ✅ Entry format implementation with CRC32 checksums
- ✅ Basic Put operation

## Project Structure

bitcask-go/
├── cmd/
│   └── bitcaskd/
│       └── main.go     # Entry point
├── internal/
│   └── bitcask/
│       ├── bitcask.go  # Core implementation
│       └── entry.go    # Data structures
└── tests/

## Entry Format

Each key-value entry is stored with the following format:
+-------+------------+---------+-----------+---------+-----------+
|  CRC  | Timestamp | KeySize | ValueSize |   Key   |   Value   |
+-------+------------+---------+-----------+---------+-----------+
|  4B   |    8B     |   4B    |    4B     | KeySize | ValueSize|


## Running Tests

```bash
cd internal/bitcask
go test -v

```

## License
MIT
