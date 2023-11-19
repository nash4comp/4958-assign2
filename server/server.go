package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"      // For clearScreen() by OS
	"os/exec" // For clearScreen() by OS
	"runtime"
	"strings"
	"sync"
)

// Define communication protocol between client and server
const (
	NickChangeCommand = "/NICK "
	MessageCommand    = "/MSG "
	ListCommand       = "/LIST"
	BroadcastCommand  = "/BC "
)

var (
	clients     = make(map[net.Conn]string) // Client connection mapping
	clientsLock sync.Mutex                  // Mutex for clients map
	newline     = "\n\r"
)

func main() {
	clearScreen()
	listener, err := net.Listen("tcp", "localhost:6666")
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}

	// Close listener when main() ends
	defer listener.Close()

	fmt.Println("Chat server is running on 6666...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept client connection:", err)
			continue
		}

		// Client connection is handled in a separate goroutine
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	var nickname string
	// The variable for checking if client has set nickname
	hasNickname := false

	// Sending welcome message to client when connected successfully
	welcomeMessage := "Welcome to the chat server!\n\rUse /NICK <nickname> to set your nickname.\n\r"
	_, err := conn.Write([]byte(welcomeMessage))
	if err != nil {
		fmt.Println("Failed to send welcome message to client: ", err)
		return
	}

	for {
		// Receive message from client
		input, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected:", err)
			// Process when client disconnects
			handleClientDisconnect(conn, nickname)
			return
		}

		// Remove newline character from input
		message := strings.TrimSpace(input)

		if strings.HasPrefix(message, NickChangeCommand) {
			// Set nickname
			newNickname := strings.TrimPrefix(message, NickChangeCommand)
			if isNicknameAvailable(newNickname) {
				nickname = newNickname

				// Mutex lock for clients map for thread safety
				clientsLock.Lock()
				// Store client's nickname
				clients[conn] = nickname
				// Mutex unlock for clients map
				clientsLock.Unlock()
				// Sending nickname confirmation message to client
				response := "Your nickname is now set to: " + nickname + newline
				hasNickname = true
				_, err := conn.Write([]byte(response))
				if err != nil {
					fmt.Println("Failed to send nickname confirmation to client:", err)
					return
				}
			} else {
				// Sending nickname in use message to client when nickname is already in use
				response := "Nickname already in use. Choose a different nickname." + newline
				_, err := conn.Write([]byte(response))
				if err != nil {
					fmt.Println("Failed to send nickname in use message to client: ", err)
					return
				}
			}
		} else if message == ListCommand && hasNickname {
			// Print client list and nickname
			// Print only nickname of clients who are already registered
			clientList := getClientList()
			response := "Client List: " + clientList + newline
			_, err := conn.Write([]byte(response))
			if err != nil {
				fmt.Println("Failed to send client list to client: ", err)
				return
			}
		} else if strings.HasPrefix(message, BroadcastCommand) && hasNickname {
			// Broadcasting the message
			broadcastMessage(strings.TrimPrefix(message, BroadcastCommand), nickname, conn)
		} else if strings.HasPrefix(message, MessageCommand) && hasNickname {
			parts := strings.SplitN(strings.TrimPrefix(message, MessageCommand), " ", 2)
			// Checking the number of arguments
			if len(parts) != 2 {
				response := "Usage: /MSG <nickname> <message>" + newline
				_, err := conn.Write([]byte(response))
				if err != nil {
					fmt.Println("Failed to send usage message to client: ", err)
					return
				}
				continue
			}

			recipientNickname := parts[0]
			messageToSend := parts[1]

			// Sending a private message to a specific nickname
			sendPrivateMessage(recipientNickname, messageToSend, nickname, conn)
		} else if hasNickname {
			commandList := NickChangeCommand + "<nickname>\n\r" + MessageCommand + "<nickname> <message>\n\r" + ListCommand + "\n\r" + BroadcastCommand + "<message>"
			response := "Please enter the correct command.\n\r" + commandList + newline
			conn.Write([]byte(response))
		} else {
			response := "You need to set your nickname at first.\n\rUse /NICK <nickname> to set your nickname." + newline
			conn.Write([]byte(response))
		}
	}
}

// Checking the nickname
func isNicknameAvailable(nickname string) bool {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	for _, existingNickname := range clients {
		if existingNickname == nickname {
			return false
		}
	}
	return true
}

func handleClientDisconnect(conn net.Conn, nickname string) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	// Delete the client information when the client disconnects
	delete(clients, conn)
	fmt.Println("Client disconnected:", nickname)
}

// Retrieving the nickname of clients
func getClientList() string {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	clientList := ""
	for _, clientNickname := range clients {
		clientList += clientNickname + " "
	}
	return clientList
}

func clearScreen() {
	switch runtime.GOOS {
	case "darwin", "linux":
		cmd := exec.Command("clear") // Clear screen for macOS and Linux
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls") // Clear screen for Windows
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		fmt.Println("Clear screen is not supported on this OS.")
	}
}

// Broadcasting the message to all clients
func broadcastMessage(message, senderNickname string, senderConn net.Conn) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	for clientConn, clientNickname := range clients {
		if clientConn != nil && clientNickname != senderNickname {
			_, err := clientConn.Write([]byte(senderNickname + ": " + message + newline))
			if err != nil {
				fmt.Println("Failed to send message to client: ", err)
			}
		} else if clientConn == senderConn {
			// Sending the same message to him/herself
			_, err := clientConn.Write([]byte("You: " + message + newline))
			if err != nil {
				fmt.Println("Failed to send message to client: ", err)
			}
		}
	}
}

func sendPrivateMessage(recipientNickname, message, senderNickname string, senderConn net.Conn) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	for clientConn, clientNickname := range clients {
		if clientConn != nil && clientNickname == recipientNickname {
			_, err := clientConn.Write([]byte(senderNickname + " (private): " + message + newline))
			if err != nil {
				fmt.Println("Failed to send private message to client: ", err)
			}
			return
		}
	}

	// Showing warning message when the nickname is not existed
	response := "User " + recipientNickname + " not found or offline." + newline
	_, err := senderConn.Write([]byte(response))
	if err != nil {
		fmt.Println("Failed to send private message to client: ", err)
	}
}
