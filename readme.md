# RPC Broadcast Chat (Go)

This project modifies a Go `net/rpc` chat system to support real-time broadcasting with Go concurrency.

## Features
- Multiple clients connect to a single RPC server.
- When a client joins, all other clients are notified: `User [ID] joined`.
- When a client sends a message, it is broadcast to all other clients (no self-echo).
- Uses goroutines and channels for concurrent send/receive.
- Shared client list is synchronized using a `Mutex`.
- Leave notification when a client exits: `User [ID] left`.

## Files
- `server.go` : RPC chat server (broadcast event loop + client registry)
- `client.go` : RPC client (input loop + long polling receive loop)

## Requirements
- Go (any recent version should work)

## How To Run

### 1) Start the server
```bash
go run server.go

The server listens on:

    127.0.0.1:12346

2) Run clients (in separate terminals)

Terminal 1:

go run client.go

Terminal 2:

go run client.go

Enter a unique ID for each client when prompted.
Quick Test

    Start 2 clients with different IDs.

    Type a message in Client A -> it should appear in Client B.

    Type a message in Client B -> it should appear in Client A.

    Sender should NOT see its own message as an incoming message (no self-echo).

    When a client joins/leaves, other clients should see a system notification.
