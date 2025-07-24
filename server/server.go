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

func (s *Server) handleConnection(conn net.Conn) {

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

		if strings.HasPrefix(msg, "PING ") {
			conn.Write([]byte("PONG\n"))
			conn.SetDeadline(time.Now().Add(30 * time.Second))
			continue
		}
		
		if strings.HasPrefix(msg, "SET ") {		// prob - replicated signature check
			s.replicateToPeers(msg)
		}
		
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

func (s *Server) StartHeartbeat() {
	go func() {
		var ticker = time.NewTicker(5 * time.Second)
		for range ticker.C {
			s.pingPeers()
		}
	}()
}

func (s *Server) pingPeers() {
	s.peerMutex.RLock()
	defer s.peerMutex.RUnlock()

	for id, addr := range s.peers {
		go func(id string, addr string) {
			var conn, err = net.DialTimeout("tcp", addr, 2 * time.Second)
			
			if err != nil {
				fmt.Printf("Node %s (%s) unreachable", id, addr)
				return
			}
			defer conn.Close()
			fmt.Fprintf(conn, "PING %s\n", s.nodeId)


			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			var _, err2 = bufio.NewReader(conn).ReadString('\n')

			if err2 != nil {
				fmt.Printf("Node %s (%s) is not responding", id, addr)
			}
		}(id, addr)
	}
}

func (s *Server) replicateToPeers(cmd string) {
	s.peerMutex.RLock()
	defer s.peerMutex.RLock()

	for _, addr := range s.peers {
		go func(addr string) {
			var conn, err = net.Dial("tcp", addr)
			if err != nil {
				return
			}
			defer conn.Close()
			fmt.Fprintf(conn, "%s /*replicated*/\n", cmd)
		}(addr)
	}
}

func (s *Server) AddPeer(nodeId string, addr string) {
	s.peerMutex.Lock()
	defer  s.peerMutex.Unlock()
	s.peers[nodeId] = addr
}