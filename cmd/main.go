package main

import (
	db "bitcask/internal"
	"errors"
	"fmt"
	"log"
	"os"
)

const dbPath = "./cask-data"

const maxFileSize = 16 * 1024 // 8 Kilobytes

func main() {
	database, err := db.Open(dbPath, maxFileSize)
	if err != nil {
		log.Fatalf("Error: Failed to open database: %v", err)
	}
	defer database.Close()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "--put":
		if len(os.Args) != 4 {
			fmt.Fprintln(os.Stderr, "Error: --put requires exactly a key and a value.")
			printUsage()
			os.Exit(1)
		}

		key := []byte(os.Args[2])
		value := []byte(os.Args[3])

		if err := database.Put(key, value); err != nil {
			log.Fatalf("Error: Failed to save data: %v", err)
		}
		fmt.Printf("OK\n Key:%s\tVal:%s\n", key, value)

	case "--get":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "Error: --get requires exactly one key.")
			printUsage()
			os.Exit(1)
		}

		key := []byte(os.Args[2])
		value, err := database.Get(key)

		if err != nil {
			if errors.Is(err, db.ErrKeyNotFound) {
				fmt.Fprintf(os.Stderr, "Error: Key not found: %s\n", key)
			} else {
				log.Fatalf("Error: Failed to retrieve data: %v", err)
			}
			os.Exit(1)
		}
		fmt.Println(string(value))

	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "\nUsage:")
	fmt.Fprintln(os.Stderr, "  bitcask-go --put <key> <value>")
	fmt.Fprintln(os.Stderr, "  bitcask-go --get <key>")
}
