package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/moanrisy/ssg/shared"
)

var upgrader = websocket.Upgrader{}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	fmt.Println("Client connected", c.RemoteAddr())
	fmt.Println()

	c.WriteMessage(websocket.TextMessage, []byte("Hello client!"))

	count := 0
	for {

		var message shared.Message

		c.ReadJSON(&message)
		switch message.Type {
		case shared.MESSAGE:
			fmt.Printf("Message from client: %v", message.Content)
		case shared.INPUT:
			fmt.Printf("From client: %v", message.Content)
			count++
			fmt.Println("Input received:", count)
		}

		fmt.Println()

	}
}

func main() {
	fmt.Println("Server started")

	http.HandleFunc("/ws", echo)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
