package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
	"github.com/moanrisy/ssg/shared"
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

	// Send input to the server
	for {
		inputReader := bufio.NewReader(os.Stdin)
		input, _ := inputReader.ReadString('\n')

		message.Content = input
		message.Type = shared.INPUT

		err = conn.WriteJSON(&message)
		if err != nil {
			log.Println(err)
			return
		}
	}
}
