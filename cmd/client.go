package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
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

	closeChan := make(chan struct{})

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
				// log.Println(err)
				close(closeChan)
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
			var input string
			// Loop until selected input is valid number
			for {
				inputReader := bufio.NewReader(os.Stdin)
				input, err = inputReader.ReadString('\n')
				if err != nil {
					fmt.Println("Error reading input:", err)
					return
				}

				input = strings.TrimSpace(input)
				number, err := strconv.Atoi(input)
				// Check if input is not number
				if err != nil {
					fmt.Println("Invalid input. Please enter a valid number.")
					continue
				}

				// Check if the number is within the specified range
				if number >= 0 && number <= 123 {
					result := strconv.Itoa(number)
					fmt.Printf("You entered: %s\n", result)
					break
				} else {
					fmt.Println("Number out of range. Please enter a number between 0 and 123.")
				}
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
			close(closeChan)
			return

		case <-closeChan:
			fmt.Println("Connection closed by server")
			fmt.Println("Thank you for playing")
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
