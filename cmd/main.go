package main

import (
	"fmt"

	db "bitcask/internal"
)

type Bitcask struct {
	file     *db.BitcaskFile
	indexing *db.Indexing
	lock     *db.LockFile
}

func main() {
	fmt.Printf(" -- Implementation of Bitcask paper -- ")
	fmt.Printf("Opening a new instance of db \n Choose whatever option you want: \n 1- Create DB \n 2- Open file \n 3- Exit \n")
}
