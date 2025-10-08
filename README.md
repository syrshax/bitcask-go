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

## Build

To build the executable, ensure you have **Go (version 1.18 or higher)** installed.

1.  **Clone the repository:**
    ```bash
    git clone <your-repo-url>
    cd go-cask
    ```
2.  **Build the executable:**
    Use the `go build` command in the root directory. This will create a runnable binary named `go-cask` (or `go-cask.exe` on Windows).
    ```bash
    go build -o go-cask
    ```
---

## Usage
Once the executable is built, you can interact with the key-value store using the following commands. The data directory is created at `./cask-data` by default.

### Storing Data (`--put`)

Use the `--put` command followed by the `<key>` and `<value>` you want to store.

```bash
# Syntax: ./go-cask --put <key> <value>
$ ./go-cask --put username alice
OK
 Key:username  Val:alice

$ ./go-cask --put city "San Francisco"
OK
 Key:city  Val:San Francisco
```

### Getting Data (`--get`)
Use the `--get`command followed by the `<key>` to retrieve the stored value.

```bash
# Syntax: ./go-cask --get <key>
$ ./go-cask --get username
alice

# If the key is not found:
$ ./go-cask --get age
Error: Key not found: age
```
