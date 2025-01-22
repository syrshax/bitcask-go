package main

import (
	"database/sql"
	"fmt"
)

type db struct {
	conn *sql.Conn
}

func main() {
	fmt.Printf("Opening a new instance of db chose whatever option you want: \n 1- Create DB \n 2- Open file \n 3- Exit \n")
}
