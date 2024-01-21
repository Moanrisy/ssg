package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/moanrisy/ssg/shared"
)

var (
	messageMutex sync.Mutex
	message      shared.Message
	inputChannel chan string
)

func main() {
	fmt.Println("Client started")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Connect to the WebSocket server
	addr := "ws://localhost:8081/ws"
	conn, _, err := websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	done := make(chan struct{})

	// Handle incoming messages from the server
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}
			fmt.Printf("Received message: %s\n", message)
		}
	}()

	// Send Hi message to the server
	var message shared.Message
	message.Content = string([]byte("Hi server ")) + "I'm from " + conn.LocalAddr().String() + "\n"
	message.Type = shared.MESSAGE
	conn.WriteJSON(&message)

	// Input need to use channel
	// because bufio will block Interrupt signal (C-c to Interrupt is blocked until enter pressed)
	inputChannel = make(chan string, 1)
	go func() {
		for {
			inputReader := bufio.NewReader(os.Stdin)
			input, err := inputReader.ReadString('\n')
			if err != nil {
				log.Println(err)
				return
			}
			inputChannel <- input
		}
	}()

	// Interrupt with C-c or send input to server
	for {
		select {
		case <-interrupt:
			fmt.Println("Interrupt received, closing connection...")
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println(err)
			}
			return

		case input := <-inputChannel:
			messageMutex.Lock()
			message.Content = input
			message.Type = shared.INPUT
			messageMutex.Unlock()

			err = conn.WriteJSON(&message)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}
