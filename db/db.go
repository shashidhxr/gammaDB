package db

import (
	"strings"
	"sync"
)

type DB struct {
	data map[string]string
	mu sync.RWMutex
}

func New() *DB {			// gg
	return &DB {
		data: make(map[string]string),
	}
}

func (db *DB) HandleCommand(command string) string {
	var parts []string = strings.Split(command, " ")

	if len(parts) < 1 {
		return "ERROR: Empty command"
	}

	var cmd = strings.ToUpper(parts[0])
	
	switch cmd {
		case "SET": 
			if(len(parts) != 3) {
				return "ERROR: use SET <KEY> <VALUE>"
			}
			db.mu.Lock()
			db.data[parts[1]] = parts[2]
			db.mu.Unlock()
			return "OK"
			
		case "GET":
			if(len(parts) != 2) {
				return "ERROR: use GET <KEY>"
			}
			db.mu.RLock()
			var val, exists = db.data[parts[1]]
			db.mu.RUnlock()

			if !exists {
				return "ERROR: Key not found"
			}
			return val

		case "DELETE":
			if(len(parts) != 2) {
				return "ERROR: use DELETE <KEY>"
			}
			db.mu.Lock()
			defer db.mu.Unlock()
			var _, exists = db.data[parts[1]]

			if !exists {
				return "WARNING: Key does not exist"
			}
			delete(db.data, parts[1])
			return "OK"
			
		default:
			return "Unknown commnad, Use HELP for manual"
		}
}