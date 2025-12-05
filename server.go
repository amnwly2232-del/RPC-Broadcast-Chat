// server.go
package main

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type Message struct {
	From   string
	Text   string
	Time   time.Time
	System bool
}

type JoinArgs struct{ ID string }
type JoinReply struct{ OK bool }

type LeaveArgs struct{ ID string }
type LeaveReply struct{ OK bool }

type SendArgs struct {
	ID   string
	Text string
}
type SendReply struct{ OK bool }

type PollArgs struct {
	ID        string
	TimeoutMs int
	MaxBatch  int
}
type PollReply struct {
	Messages []Message
}

type ClientState struct {
	id     string
	inbox  chan Message
	closed bool
}

type Chat struct {
	mu      sync.Mutex
	clients map[string]*ClientState
	events  chan Message
}

func NewChat() *Chat {
	c := &Chat{
		clients: make(map[string]*ClientState),
		events:  make(chan Message, 1024),
	}
	go c.runBroadcaster()
	return c
}

func (c *Chat) runBroadcaster() {
	for msg := range c.events {
		type dst struct {
			id    string
			inbox chan Message
		}
		var recips []dst

		c.mu.Lock()
		for id, st := range c.clients {
			if st.closed {
				continue
			}
			if id == msg.From {
				continue
			}
			recips = append(recips, dst{id: id, inbox: st.inbox})
		}
		c.mu.Unlock()

		for _, r := range recips {
			select {
			case r.inbox <- msg:
			default:
				log.Printf("[SERVER] client %q inbox full -> dropping msg\n", r.id)
			}
		}

		if msg.System {
			log.Printf("[SERVER] %s\n", msg.Text)
		} else {
			log.Printf("[SERVER %s] %s: %s\n", msg.Time.Format("15:04:05"), msg.From, msg.Text)
		}
	}
}

func (c *Chat) systemFrom(id, text string) {
	c.events <- Message{
		From:   id,
		Text:   text,
		Time:   time.Now(),
		System: true,
	}
}

func (c *Chat) Join(args JoinArgs, reply *JoinReply) error {
	id := args.ID
	if id == "" {
		return errors.New("empty id")
	}

	c.mu.Lock()
	if _, exists := c.clients[id]; exists {
		c.mu.Unlock()
		return errors.New("id already in use")
	}
	c.clients[id] = &ClientState{
		id:    id,
		inbox: make(chan Message, 64),
	}
	c.mu.Unlock()

	reply.OK = true
	c.systemFrom(id, fmt.Sprintf("User [%s] joined", id))
	return nil
}

func (c *Chat) Leave(args LeaveArgs, reply *LeaveReply) error {
	id := args.ID
	if id == "" {
		return errors.New("empty id")
	}

	c.mu.Lock()
	st, ok := c.clients[id]
	if ok && !st.closed {
		st.closed = true
		delete(c.clients, id)
		close(st.inbox)
	}
	c.mu.Unlock()

	reply.OK = ok
	if ok {
		c.systemFrom(id, fmt.Sprintf("User [%s] left", id))
	}
	return nil
}

func (c *Chat) Send(args SendArgs, reply *SendReply) error {
	if args.ID == "" {
		return errors.New("empty id")
	}
	if args.Text == "" {
		return errors.New("empty message")
	}

	c.mu.Lock()
	_, ok := c.clients[args.ID]
	c.mu.Unlock()
	if !ok {
		return errors.New("not joined")
	}

	reply.OK = true
	c.events <- Message{
		From:   args.ID,
		Text:   args.Text,
		Time:   time.Now(),
		System: false,
	}
	return nil
}

func (c *Chat) Poll(args PollArgs, reply *PollReply) error {
	if args.ID == "" {
		return errors.New("empty id")
	}
	if args.TimeoutMs <= 0 {
		args.TimeoutMs = 25000
	}
	if args.MaxBatch <= 0 {
		args.MaxBatch = 32
	}

	c.mu.Lock()
	st, ok := c.clients[args.ID]
	c.mu.Unlock()
	if !ok {
		return errors.New("not joined")
	}

	timeout := time.NewTimer(time.Duration(args.TimeoutMs) * time.Millisecond)
	defer timeout.Stop()

	select {
	case msg, ok := <-st.inbox:
		if !ok {
			return errors.New("client disconnected")
		}
		reply.Messages = append(reply.Messages, msg)

		for len(reply.Messages) < args.MaxBatch {
			select {
			case m2, ok := <-st.inbox:
				if !ok {
					return errors.New("client disconnected")
				}
				reply.Messages = append(reply.Messages, m2)
			default:
				return nil
			}
		}
		return nil

	case <-timeout.C:
		reply.Messages = nil
		return nil
	}
}

func main() {
	chat := NewChat()
	if err := rpc.Register(chat); err != nil {
		log.Fatal(err)
	}

	l, err := net.Listen("tcp", ":12346")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	log.Println("RPC broadcast chat server listening on :12346")
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("accept error:", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
