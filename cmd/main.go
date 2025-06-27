package main

import (
	db "bitcask/internal"
	"fmt"
)

func main() {
	fmt.Printf(" -- Implementation of Bitcask paper --\n\n")
	db, _ := db.Open("database")
	key := []byte("foo")
	val := []byte("bar")
	db.Put(key, val)
	data, _ := db.Get(key)
	fmt.Printf("KEY: %v\tVALUE: %v\n", string(key), string(data))
}
