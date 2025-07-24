package main

import (
	"xbegd/server"
	"xbegd/db"
)

func main() {
	var db1 = db.New()

	var svr1 = server.New(db1, "node1")
	svr1.AddPeer("node2", "localhost:9091")
	go svr1.Start("9090")
	svr1.StartHeartbeat()

	select{}
}