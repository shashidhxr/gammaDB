package main

import (
	"fmt"
	"os"
	"time"

	"xbegd/db"
	"xbegd/server"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("USAGE: go run main.go <nodeid>")
		return
	}

	var nodeId = os.Args[1]
	var db = db.New(nodeId)
	var server = server.New(db, nodeId)

	var port string
	switch nodeId {
	case "node1":
		server.AddPeer("node2", "localhost:9091")
		port = "9090"
	case "node2":
		server.AddPeer("node1", "localhost:9090")
		port = "9091"
	default:
		fmt.Printf("Unknown node ID: %s. Use 'node1' or 'node2'\n", nodeId)
		return
	}

	go func() {
		err := server.Start(port)
		if err != nil {
			fmt.Printf("Server failed to start: %v\n", err)
			os.Exit(1)
		}
	}()

	time.Sleep(1 * time.Second)

	server.StartHeartbeat()

	// var db1 = db.New("node1");
	// var svr1 = server.New(db1, "node1")

	// svr1.AddPeer("node2", "localhost:9091")

	// go svr1.Start("9090")
	// svr1.StartHeartbeat()

	// var db2 = db.New("node2")
	// var svr2 = server.New(db2, "node2")
	// svr2.AddPeer("node1", "localhost:9090")
	// go svr2.Start("9091")
	// svr2.StartHeartbeat()

	// fmt.Printf("Node %s is running. \n", nodeId)

	select {}
}
