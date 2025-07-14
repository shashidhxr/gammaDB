package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"

	"xbegd/db"
)

type Server struct {
	db *db.DB
}

func New(db *db.DB) *Server {
	return &Server{	db: db}
}

func (s *Server)handleConnection(conn net.Conn) {
	defer conn.Close()

	// todo - conn deadline(timeout)

	var reader = bufio.NewReader(conn)
	for {
		var msg, err = reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client Disconnected")
			return
		}

		msg = strings.TrimSpace(msg)
		fmt.Printf("Recieved: %s\n", msg)

		// conn.Write([]byte("ECHO: " + msg + "\n"))
		var response = s.db.HandleCommand(msg)
		conn.Write([]byte(response + "\n"))
	}
}

func (s *Server) Start(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("Server started on port %s\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			continue
		}
		go s.handleConnection(conn)
	}
}