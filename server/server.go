package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"xbegd/db"
)

type Server struct {
	db *db.DB
	nodeId string
	peers map[string]string
	peerMutex sync.RWMutex
}

func New(db *db.DB, nodeId string) *Server {
	return &Server {
		db: db,
		nodeId: nodeId,
		peers: make(map[string]string), 
	}
}

func (s *Server)handleConnection(conn net.Conn) {

	conn.SetDeadline(time.Now().Add(30 * time.Second))				// os.ErrDeadlineExceeded
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))				// r and w for threads 
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	
	defer conn.Close()

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
		conn.SetDeadline(time.Now().Add(30 * time.Second))
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