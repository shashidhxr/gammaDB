# GammaDB – A Distributed In-Memory Key-Value Store

![GammaDB Architecture](./assets/architecture.png)

---

## Features



- **Distributed Storage**: Data partitioned across multiple nodes  
- **Replication**: Ensures data redundancy and availability  
- **Heartbeat Protocol**: Detects node failures and maintains cluster health  
- **Multi-Client Support**: Handles concurrent connections efficiently  

---

## Getting Started

### Prerequisites
- Go 1.20+
- Git

### Installation

```bash
git clone https://github.com/yourusername/gammaDB.git
cd gammDB
go build

# Start a 3-node cluster
./gammaDB -id node1 -port 9090
./gammaDB -id node2 -port 9091
./gammaDB -id node3 -port 9092
```

### API Documentation

#### HTTP API     (to be implemented)
GET    /v1/key/{key} 
POST   /v1/key/{key}
DELETE /v1/key/{key}

#### TCP Protocol
SET <key> <value>
GET <key>
DELETE <key>


### Development Roadmap
- Core key-value storage - done
- TCP/HTTP interfaces - done
- Multi-client connection - done
- Multi-node connection - done
- Replication - done
- Raft consensus protocol (WIP)
- Data sharding
- TLS encryption
- Prometheus metrics