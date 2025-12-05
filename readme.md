# RPC Broadcast Chat (Go)

A simple real-time chat system built with Go **net/rpc**. Clients connect to an RPC server, and messages are **broadcast** to all other connected clients using Go **goroutines + channels**.

---

## ✅ Features
- RPC-based chat server using `net/rpc` over TCP
- Real-time broadcast (server broadcaster loop)
- Join / Leave notifications:
  - `User [ID] joined`
  - `User [ID] left`
- No self-echo (sender does **not** receive their own messages)
- Concurrency using goroutines and channels
- Thread-safe client registry using `sync.Mutex`

---

## 📁 Project Structure
- `server.go` — RPC Server (client registry + broadcaster)
- `client.go` — RPC Client (send loop + long-poll receive loop)
- `README.md` — Instructions

---

## 🧰 Requirements
- Go (any recent version)

---

## 🚀 Run Instructions
```bash
# 1) Start the server
go run server.go

# Server runs on:
# 127.0.0.1:12346

# 2) Run clients (open 2 terminals)
# Terminal 1:
go run client.go

# Terminal 2:
go run client.go
