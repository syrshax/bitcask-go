# Go-Cask: A Bitcask Implementation in Go
Go-Cask is an educational project to build a high-performance key-value store in Go, based on the principles of the Bitcask paper. It features an append-only log structure for writes and a fast, in-memory index for reads.

## Features Implemented:
- Put(key, value): Persists a key-value pair to disk.

- Get(key): Retrieves a value by its key.

- Append-Only Persistence: All data is written to a single, append-only data file, ensuring high write throughput.

- In-Memory Keydir: All keys and the on-disk location of their values are stored in a hash map for extremely fast lookups.


## Design
Go-Cask follows the core design of Bitcask, separating data storage from the index.

### On-Disk Format
Each record written to the data file is a binary-encoded entry with the following structure:

| Field      | Size (bytes) | Description                               |
| :--------- | :----------- | :---------------------------------------- |
| **CRC** | 4            | 32-bit CRC checksum of the key and value  |
| **Timestamp**| 4            | Unix timestamp of the write operation     |
| **Key Size** | 4            | The size of the key in bytes              |
| **Value Size**| 4            | The size of the value in bytes            |
| **Key** | Variable     | The key data                              |
| **Value** | Variable     | The value data                            |


### In-Memory Index (Keydir)
To provide fast reads, Go-Cask holds a map[string]IndexEntry in memory. This map, the "Keydir," acts as a pointer to the location of the latest value for every key on disk.

An **IndexEntry** contains:

- filename: The data file where the value is stored.

- offset: The byte offset pointing to the start of the value.

- size: The size of the value in bytes.

This means a Get operation is just one fast hash map lookup followed by a single disk seek and read.
