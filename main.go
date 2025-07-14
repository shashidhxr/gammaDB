package main

import (
	"./server"
	"./db"
)

func main() {
	var database = db.New()

	var svr = server.New(database)

	if err := svr.Start("9090"); err != nil {			// gg
		panic(err)
	}
}