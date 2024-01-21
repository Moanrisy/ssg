package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/moanrisy/ssg/shared"
)

type player int

const (
	PLAYER1 player = iota
	PLAYER2
)

type GameState struct {
	playerWebsocketConn [2]*websocket.Conn
	playerTurn          player
	previousPlayerTurn  player
	playerReady         [2]bool
}

func NewGameState() *GameState {
	return &GameState{
		playerWebsocketConn: [2]*websocket.Conn{nil, nil},
		playerReady:         [2]bool{false, false},
		playerTurn:          PLAYER1,
		previousPlayerTurn:  PLAYER2,
	}
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

func registerPlayer(c *websocket.Conn) {
	fmt.Println("Client connected", c.RemoteAddr())
	countMu.Lock()
	if !gameState.playerReady[PLAYER1] {
		gameState.playerReady[PLAYER1] = true
		Players[PLAYER1].playerNumber = PLAYER1
		Players[PLAYER1].idFromAddr = c.RemoteAddr().String()
		gameState.playerWebsocketConn[PLAYER1] = c
	} else if !gameState.playerReady[PLAYER2] {
		gameState.playerReady[PLAYER2] = true
		Players[PLAYER2].playerNumber = PLAYER2
		Players[PLAYER2].idFromAddr = c.RemoteAddr().String()
		gameState.playerWebsocketConn[PLAYER2] = c
	} else {
		fmt.Print("Maximum 2 players allowed!")
	}
	countMu.Unlock()
}

func welcomeMessage(c *websocket.Conn) {
	var messageFrom string
	var player player
	switch c.RemoteAddr().String() {
	case Players[PLAYER1].idFromAddr:
		messageFrom = "Player 1"
		player = PLAYER1
	case Players[PLAYER2].idFromAddr:
		messageFrom = "Player 2"
		player = PLAYER2
	}
	welcomeMessage := "Welcome, you are " + messageFrom
	c.WriteMessage(websocket.TextMessage, []byte(welcomeMessage))

	if player == PLAYER1 {
		// additionalMessageAfterPlayer1Connected
		gameState.playerWebsocketConn[gameState.playerTurn].WriteMessage(websocket.TextMessage, []byte("Waiting player 2 to join..."))
	}

	if player == PLAYER2 {
		// additionalMessageAfterPlayer2Connected
		// ToPlayer1
		gameState.playerWebsocketConn[gameState.playerTurn].WriteMessage(websocket.TextMessage, []byte("Player 2 connected to the game.\nGame started.\n===================\nIt's your turn, please type input between 0-123"))
		// ToPlayer2
		c.WriteMessage(websocket.TextMessage, []byte("\nGame started.\n===================\nWaiting other player turn..."))
	}
}

func playerSendInput(playerTurn player, message shared.Message) {
	// disable cebug log
	// fmt.Printf("From player %v\t", Players[playerTurn].playerNumber+1)
	// fmt.Printf("Input choosen: %v\n", message.Content)
	countMu.Lock()
	Players[playerTurn].input = append(Players[playerTurn].input, message.Content)
	count++
	// fmt.Println("Input received:", count)
	countMu.Unlock()
}

func printPickedNumbersAsReference() {
	line := strings.Repeat("=", 100)
	fmt.Println(line)

	fmt.Println("Player", Players[PLAYER1].playerNumber+1, "choosen input is: ")
	for _, v := range Players[PLAYER1].input {
		fmt.Print(v)
		fmt.Print("\t|\t")
	}
	fmt.Println()
	fmt.Println(line)
	fmt.Println("Player", Players[PLAYER2].playerNumber+1, "choosen input is: ")
	for _, v := range Players[PLAYER2].input {
		fmt.Print(v)
		fmt.Print("\t|\t")
	}
	fmt.Println()
	fmt.Println(line)
}

func printGameCompleted() {
	if len(Players[PLAYER2].input) == 5 {
		fmt.Println()
		err := gameState.playerWebsocketConn[0].Close()
		if err != nil {
			fmt.Println("Error closing player 1 websocket connection")
		}
		gameState.playerWebsocketConn[1].Close()
		if err != nil {
			fmt.Println("Error closing player 2 websocket connection")
		}
		fmt.Println("Game completed")
		os.Exit(0)
	}
}

var gameState = NewGameState()

func echo(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	registerPlayer(c)
	welcomeMessage(c)

	for {

		var message shared.Message

		err := c.ReadJSON(&message)
		if err != nil {
			// fmt.Println("Client disconnected", c.RemoteAddr())
			break
		}

		switch message.Type {
		case shared.MESSAGE:
			fmt.Printf("Message from client: %v", message.Content)
		case shared.INPUT:
			switch c.RemoteAddr().String() {
			case Players[PLAYER1].idFromAddr:
				gameState.playerTurn = PLAYER1
			case Players[PLAYER2].idFromAddr:
				gameState.playerTurn = PLAYER2
			}

			if gameState.previousPlayerTurn == gameState.playerTurn {
				c.WriteMessage(websocket.TextMessage, []byte("It's not your turn!"))
			} else {
				playerSendInput(gameState.playerTurn, message)
				gameState.playerWebsocketConn[gameState.previousPlayerTurn].WriteMessage(websocket.TextMessage, []byte("It's your turn, please type input between 0-123"))
				c.WriteMessage(websocket.TextMessage, []byte("Waiting other player turn..."))
				gameState.previousPlayerTurn = gameState.playerTurn
			}
		}

		fmt.Println()

		if len(Players[PLAYER1].input) > 0 {
			shared.ClearTerminal()
			printPickedNumbersAsReference()
			printGameCompleted()
		}
	}
}

func main() {
	fmt.Println("Server started")

	http.HandleFunc("/ws", echo)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
