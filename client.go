// client.go
package main

import (
	"bufio"
	"fmt"
	"net/rpc"
	"os"
	"strings"
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

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your ID: ")
	idRaw, _ := reader.ReadString('\n')
	id := strings.TrimSpace(idRaw)
	if id == "" {
		id = "anonymous"
	}

	client, err := rpc.Dial("tcp", "127.0.0.1:12346")
	if err != nil {
		fmt.Println("dial error:", err)
		return
	}
	defer client.Close()

	var j JoinReply
	if err := client.Call("Chat.Join", JoinArgs{ID: id}, &j); err != nil {
		fmt.Println("join error:", err)
		return
	}

	defer func() {
		var lr LeaveReply
		_ = client.Call("Chat.Leave", LeaveArgs{ID: id}, &lr)
	}()

	fmt.Println("Joined. Type messages. Type 'exit' to quit.")
	fmt.Println("Note: no self-echo (your messages won't be shown back to you).")

	done := make(chan struct{})

	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}

			var pr PollReply
			err := client.Call("Chat.Poll", PollArgs{ID: id, TimeoutMs: 25000, MaxBatch: 32}, &pr)
			if err != nil {
				fmt.Println("\nPoll error:", err)
				close(done)
				return
			}

			for _, m := range pr.Messages {
				if m.System {
					fmt.Printf("\n%s\n> ", m.Text)
				} else {
					fmt.Printf("\n[%s] %s: %s\n> ", m.Time.Format("15:04:05"), m.From, m.Text)
				}
			}
		}
	}()

	for {
		fmt.Print("> ")
		line, _ := reader.ReadString('\n')
		text := strings.TrimSpace(line)

		if text == "" {
			continue
		}
		if text == "exit" || text == "/exit" {
			fmt.Println("bye!")
			close(done)
			return
		}

		var sr SendReply
		if err := client.Call("Chat.Send", SendArgs{ID: id, Text: text}, &sr); err != nil {
			fmt.Println("send error:", err)
			continue
		}
	}
}
