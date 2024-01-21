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
	upgrader  = websocket.Upgrader{}
	countMu   sync.Mutex
	count     int
	Players   [2]PlayerState
	gameState = NewGameState()
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
		gameState.playerWebsocketConn[gameState.playerTurn].WriteMessage(websocket.TextMessage, []byte("Player 2 connected to the game.\n\nGame started.\n===================\nIt's your turn, please type input between 0-123"))
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

func isNumberPicked(message shared.Message) bool {
	for _, v := range Players[PLAYER1].input {
		if message.Content == v {
			return true
		}
	}
	for _, v := range Players[PLAYER2].input {
		if message.Content == v {
			return true
		}
	}
	return false
}

func formatPickedNumbersAsReference(playerOne bool, playerTwo bool) string {
	start := 0
	end := 0
	initSeparator := ""
	separator := ""
	lineSeparator := 0

	if playerOne && playerTwo {
		start = 0
		end = 1
		separator = "\t|\t"
		initSeparator = "|\t"
		lineSeparator = 105
	} else if playerOne {
		start = 0
		end = 0
		separator = " | "
		initSeparator = " "
		lineSeparator = 45
	} else if playerTwo {
		start = 1
		end = 1
		separator = " | "
		initSeparator = " "
		lineSeparator = 45
	}

	var result strings.Builder
	line := strings.Repeat("=", lineSeparator)
	fmt.Fprintln(&result, line)

	for i := start; i <= end; i++ {
		fmt.Fprintf(&result, "Player %d chosen input is:%v", Players[i].playerNumber+1, initSeparator)
		for _, v := range Players[i].input {
			fmt.Fprintf(&result, "%v%v", v, separator)
		}
		fmt.Fprintln(&result)
		fmt.Fprintln(&result, line)
	}

	return result.String()
}

func gameCompleted() {
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
				if isNumberPicked(message) {
					c.WriteMessage(websocket.TextMessage, []byte("Number already picked, please pick other number"))
					break
				}
				playerSendInput(gameState.playerTurn, message)
				gameState.playerWebsocketConn[gameState.previousPlayerTurn].WriteMessage(websocket.TextMessage, []byte("\nIt's your turn, please type input between 0-123"))
				c.WriteMessage(websocket.TextMessage, []byte("Waiting other player turn..."))
				gameState.previousPlayerTurn = gameState.playerTurn
			}
		}

		fmt.Println()

		// Print every input from player
		if len(Players[PLAYER1].input) > 0 {
			shared.ClearTerminal()
			formattedString := formatPickedNumbersAsReference(true, true)
			fmt.Println(formattedString)
		}

		// Print at the end of the session to each player
		if len(Players[PLAYER2].input) == 5 {
			numbersPickedByPlayerOne := formatPickedNumbersAsReference(true, false)
			numbersPickedByPlayerTwo := formatPickedNumbersAsReference(false, true)
			gameState.playerWebsocketConn[0].WriteMessage(websocket.TextMessage, []byte(numbersPickedByPlayerOne))
			gameState.playerWebsocketConn[1].WriteMessage(websocket.TextMessage, []byte(numbersPickedByPlayerTwo))
			gameCompleted()
		}
	}
}

func main() {
	fmt.Println("Server started")

	http.HandleFunc("/ws", echo)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
