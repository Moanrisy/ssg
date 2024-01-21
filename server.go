package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/moanrisy/ssg/shared"
)

type player int

const (
	PLAYER1 player = iota
	PLAYER2
)

var GameState struct {
	playerReady [2]bool
}

type PlayerState struct {
	idFromAddr   string
	input        []string
	playerNumber player
}

var (
	upgrader = websocket.Upgrader{}
	countMu  sync.Mutex
	count    int
	Players  [2]PlayerState
)

func welcomeMessage(c *websocket.Conn) {
	var messageFrom string
	switch c.RemoteAddr().String() {
	case Players[PLAYER1].idFromAddr:
		messageFrom = "Player 1"
	case Players[PLAYER2].idFromAddr:
		messageFrom = "Player 2"
	}
	welcomeMessage := "Welcome, you are " + messageFrom
	c.WriteMessage(websocket.TextMessage, []byte(welcomeMessage))
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	fmt.Println("Client connected", c.RemoteAddr())
	fmt.Println()
	countMu.Lock()
	if !GameState.playerReady[PLAYER1] {
		GameState.playerReady[PLAYER1] = true
		Players[PLAYER1].playerNumber = PLAYER1
		Players[PLAYER1].idFromAddr = c.RemoteAddr().String()
	} else if !GameState.playerReady[PLAYER2] {
		GameState.playerReady[PLAYER2] = true
		Players[PLAYER2].playerNumber = PLAYER2
		Players[PLAYER2].idFromAddr = c.RemoteAddr().String()
	} else {
		fmt.Print("Maximum 2 players allowed!")
	}
	countMu.Unlock()

	welcomeMessage(c)

	for {

		var message shared.Message

		c.ReadJSON(&message)
		switch message.Type {
		case shared.MESSAGE:
			fmt.Printf("Message from client: %v", message.Content)
		case shared.INPUT:

			switch c.RemoteAddr().String() {
			case Players[PLAYER1].idFromAddr:
				fmt.Printf("From player: %v\n", Players[PLAYER1].playerNumber+1)
				fmt.Printf("Input choosen: %v\n", message.Content)
				countMu.Lock()
				Players[PLAYER1].input = append(Players[PLAYER1].input, message.Content)
				count++
				fmt.Println("Input received:", count)
				countMu.Unlock()
			case Players[PLAYER2].idFromAddr:
				fmt.Printf("From player: %v\n", Players[PLAYER2].playerNumber+1)
				fmt.Printf("Input choosen: %v\n", message.Content)
				countMu.Lock()
				Players[PLAYER2].input = append(Players[PLAYER2].input, message.Content)
				count++
				fmt.Println("Input received:", count)
				countMu.Unlock()
			}

		}

		fmt.Println()

		if len(Players[PLAYER1].input) == 5 {
			fmt.Println("Player", Players[PLAYER1].playerNumber+1, "choosen input is: ")
			for _, v := range Players[PLAYER1].input {
				fmt.Print(v)
			}
			fmt.Println()
		}

		if len(Players[PLAYER2].input) == 5 {
			fmt.Println("Player", Players[PLAYER2].playerNumber+1, "choosen input is: ")
			for _, v := range Players[PLAYER2].input {
				fmt.Print(v)
			}
			fmt.Println()
		}
	}
}

func main() {
	fmt.Println("Server started")

	http.HandleFunc("/ws", echo)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
