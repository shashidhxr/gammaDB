# GammaDB – A Distributed In-Memory Key-Value Store

![GammaDB Architecture](./assets/architecture.png)

---

## Features

- **Distributed Storage**: Data partitioned across multiple nodes  
- **Replication**: Ensures data redundancy and availability  
- **Multi-Client Support**: Handles multiple concurrent connections
- **Heartbeat Protocol**: Detects node failures and maintains cluster health  

---

## Getting Started

### Prerequisites
- Go 1.20+
- Git

### Installation

```bash
git clone https://github.com/shashidhxr/gammaDB.git
cd gammDB
go build

# Start a 2 nodes with id "node1" and "node2" in 2 terminals
./gammaDB node1
./gammaDB node2
```
# To send client requests, from different terminals(mimicing multiple clients)
```bash
telnet localhost:9090
    # use db commands

telnet localhost: 9091
    # use db commands
``` 

#### DB Commands
SET <key> <value>
GET <key>
DELETE <key>


### Development Roadmap
- Core key-value storage - done
- TCP interfaces - done
- Multi-client connection - done
- Multi-node connection - done
- Replication - done
- Raft consensus protocol (WIP)
- Data sharding
- TLS encryption
- Prometheus metrics