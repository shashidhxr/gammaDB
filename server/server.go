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

	defer conn.Close()					// ?
	
	conn.SetDeadline(time.Now().Add(30 * time.Second))				// os.ErrDeadlineExceeded
	
	var reader = bufio.NewReader(conn)
	for {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))				// r and w for threads 
		
		
		var msg, err = reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Client connection disconnected: %v\n", err)
			return
		}
		
		msg = strings.TrimSpace(msg)
		fmt.Printf("Recieved: %s\n", msg)
		
		if strings.HasPrefix(msg, "PING ") {
			var response = "PONG " + s.nodeId + "\n"
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
			
			var _, err = conn.Write([]byte(response))
			if err != nil {
				fmt.Printf("Failed to send PONG: %v\n", err)
				return
			}
			
			conn.SetDeadline(time.Now().Add(30 * time.Second))		// reset deadline after success
			continue
		}
		
		
		var response string
		
		if strings.HasPrefix(msg, "SET ") {	
			if strings.Contains(msg, "/*replicated*/") {
				msg = strings.ReplaceAll(msg, " /*replicated*/", "")
				response = s.db.HandleCommand(msg)
				fmt.Printf("[%s] Executed replicated cmd: %s\n", s.nodeId, msg)
			} else {						// og cmd
				fmt.Printf("[%s] Replicating cmd: %s\n", s.nodeId, msg)
				if s.replicateCommand(msg) {
					response = s.db.HandleCommand(msg)
					fmt.Printf("[%s] Cmd replicated and Executed succesfully\n", s.nodeId)
				} else {
					response = "ERROR: Failed to Replicate"
					fmt.Printf("[%s] Replication failed\n", s.nodeId)
				}
			}
			// s.replicateToPeers(msg)
		} else {
			response = s.db.HandleCommand(msg)
		}
		
		// conn.Write([]byte("ECHO: " + msg + "\n"))

		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		var _, err2 = conn.Write([]byte(response + "\n"))			// sending response	
		if err2 != nil {
			fmt.Printf("Failed to send response: %v\n", err2)
		}
		
		conn.SetDeadline(time.Now().Add(30 * time.Second))			// reser timeout after success
	}
}

func (s *Server) Start(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	defer listener.Close()			// ?

	fmt.Printf("[%s] Server started on port %s\n", s.nodeId, port)

	for {
		var conn, err = listener.Accept()
		if err != nil {
			fmt.Printf("Accept error: %v\n", err)
			continue
		}
		fmt.Printf("[%s] New connectino accepted\n", s.nodeId)
		go s.handleConnection(conn)
	}
}

func (s *Server) StartHeartbeat() {
	go func() {
		time.Sleep(2 * time.Second)		// ?
		
		var ticker = time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			s.pingPeers()
		}
	}()
}

func (s *Server) pingPeers() {
	s.peerMutex.RLock()
	var peers = make(map[string]string)		// ?

	for id, addr := range s.peers {
		peers[id] = addr
	}

	s.peerMutex.RUnlock()		// defer?

	for id, addr := range s.peers {
		go func(peerId string, peerAddr string) {
			var conn, err = net.DialTimeout("tcp", peerAddr, 2 * time.Second)
			
			if err != nil {
				fmt.Printf("[%s] Node %s (%s) unreachable: %v\n", s.nodeId, peerId, peerAddr, err)
				return
			}
			defer conn.Close()
			
			// Send PING msg
			pingMsg := fmt.Sprintf("PING %s\n", s.nodeId)
			conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
			_, err = conn.Write([]byte(pingMsg))
			if err != nil {
				fmt.Printf("[%s] Failed to send PING to %s: %v\n", s.nodeId, id, err)
				return
			}

			// Wait for PONg
			conn.SetReadDeadline(time.Now().Add(2 * time.Second))
			reader := bufio.NewReader(conn)
			response, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("[%s] Node %s (%s) not responding: %v\n", s.nodeId, peerId, peerAddr, err)
				return
			}

			response = strings.TrimSpace(response)

			if strings.HasPrefix(response, "PONG") {
				fmt.Printf("[%s] node %s(%s) is alive\n", s.nodeId, peerId, peerAddr)
			} else {
				fmt.Printf("[%s] Did not recieve PONG from %s: %s\n", s.nodeId, peerId, response)
			}
		}(id, addr)
	}
}

func (s *Server) replicateCommand(cmd string) bool {
    s.peerMutex.RLock() // Read lock for peers map

	peers := make(map[string]string)
	for id, addr := range s.peers {
		peers[id] = addr
	}
	
    defer s.peerMutex.RUnlock()

	if len(peers) == 0 {
		fmt.Printf("[%s] No peers to replicate to\n", s.nodeId)
		return true
	}


   	var wg sync.WaitGroup
	var mu sync.Mutex

	var successCount = 0
	var totalPeers = len(peers)

    for id, addr := range s.peers {
        wg.Add(1)
        go func(peerId string, peerAddr string) {
            defer wg.Done()
            
            // Connect to peer with timeout
            conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
            if err != nil {
				fmt.Printf("[%s] Failed to connect to peer %s for replication: %v\n", s.nodeId, id, err)
                return
            }
            defer conn.Close()
            			
            // Send command with replication marker
			var replCommand = fmt.Sprintf("%s /*replicated*/\n", cmd)
			conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
			var _, err2 = conn.Write([]byte(replCommand))
			if err2 != nil {
				fmt.Printf("[%s] Failed to send replication to %s: %v\n", s.nodeId, peerId, err2)
				return
			}
			
			
            // Wait for acknowledgment	
			conn.SetReadDeadline(time.Now().Add(3 * time.Second))
            reader := bufio.NewReader(conn)
            ack, err := reader.ReadString('\n')
			
            if err != nil || !strings.HasPrefix(ack, "OK") {
				fmt.Printf("[%s] Failed to receive ack from %s: %v\n", s.nodeId, peerId, err)
				return
            }

			ack = strings.TrimSpace(ack)
			if ack == "OK" {
				mu.Lock()
				successCount++
				mu.Unlock()

				fmt.Printf("[%s] successfully replicated to %s\n", s.nodeId, peerId)
			} else {
				fmt.Printf("[%s] replication to %s failed: %s\n", s.nodeId, id, ack)
			}
        }(id, addr)
    }

    wg.Wait() // Wait for all replications to complete
    
	var success = successCount == totalPeers
	fmt.Printf("[%s] replication result: %d/%d peers successful\n", s.nodeId, successCount, totalPeers)
	return success
}

// func (s *Server) replicateToPeers(cmd string) {
// 	s.peerMutex.RLock()
// 	defer s.peerMutex.RLock()

// 	for _, addr := range s.peers {
// 		go func(addr string) {
// 			var conn, err = net.Dial("tcp", addr)
// 			if err != nil {
// 				return
// 			}
// 			defer conn.Close()
// 			fmt.Fprintf(conn, "%s /*replicated*/\n", cmd)
// 		}(addr)
// 	}
// }

func (s *Server) AddPeer(nodeId string, addr string) {
	s.peerMutex.Lock()
	defer  s.peerMutex.Unlock()
	s.peers[nodeId] = addr

	fmt.Printf("[%s] Added peer: %s - %s\n", s.nodeId, nodeId, addr)
}