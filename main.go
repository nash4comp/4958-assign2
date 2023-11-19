package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

// 클라이언트와 서버 간의 통신 프로토콜 정의
const (
	NickChangeCommand = "/NICK "
	MessageCommand    = "/MSG "
)

var (
	clients     = make(map[net.Conn]string) // 연결된 클라이언트와 닉네임 매핑
	clientsLock sync.Mutex                  // 클라이언트 매핑을 위한 뮤텍스
	newline     = "\n\r"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:6666")
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
	defer listener.Close()

	fmt.Println("Chat server is running...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept client connection:", err)
			continue
		}

		// 클라이언트 연결을 별도의 고루틴으로 처리
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	var nickname string
	// var nicknameSet bool // 닉네임이 설정되었는지 여부를 나타내는 변수 추가

	// 클라이언트에게 연결 성공 메시지 안내 전송
	welcomeMessage := "Welcome to the chat server! Use /NICK <nickname> to set your nickname.\n\r"
	_, err := conn.Write([]byte(welcomeMessage))
	if err != nil {
		fmt.Println("Failed to send welcome message to client:", err)
		return
	}

	for {
		// 클라이언트로부터 메시지를 받음
		input, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected:", err)
			// 클라이언트가 연결을 종료한 경우 처리
			handleClientDisconnect(conn, nickname)
			return
		}

		message := strings.TrimSpace(input)

		if strings.HasPrefix(message, NickChangeCommand) {
			// 닉네임 설정 시도
			newNickname := strings.TrimPrefix(message, NickChangeCommand)
			if isNicknameAvailable(newNickname) {
				nickname = newNickname
				// nicknameSet = true // 닉네임이 설정됨을 표시
				// 닉네임 설정 성공 메시지 전송
				response := "Your nickname is now set to: " + nickname + newline
				_, err := conn.Write([]byte(response))
				if err != nil {
					fmt.Println("Failed to send nickname confirmation to client:", err)
					return
				}
			} else {
				// 닉네임이 이미 사용 중인 경우 메시지 전송
				response := "Nickname already in use. Choose a different nickname." + newline
				_, err := conn.Write([]byte(response))
				if err != nil {
					fmt.Println("Failed to send nickname in use message to client:", err)
					return
				}
			}
		} else if message == "/LIST" {
			// 클라이언트 리스트와 닉네임 출력
			clientList := getClientList()
			response := "Client List:\n" + clientList + newline
			_, err := conn.Write([]byte(response))
			if err != nil {
				fmt.Println("Failed to send client list to client:", err)
				return
			}
		} else {
			// 닉네임 설정 완료된 후에는 다른 메시지를 처리 (예: 브로드캐스트)
			// 이 부분에 메시지 처리 로직을 추가하세요.
			// broadcastMessage 함수를 호출하거나 다른 로직을 구현하여 메시지를 처리합니다.
			// 예: broadcastMessage(message, nickname, conn)
		}
	}
}

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

	delete(clients, conn)
	fmt.Println("Client disconnected:", nickname)

	// 클라이언트 연결 종료 후 처리 로직 추가 (예: 닉네임 해제)
}

// broadcastMessage 함수를 정의하여 메시지를 모든 클라이언트에게 전송할 수 있습니다.
func broadcastMessage(message, senderNickname string, senderConn net.Conn) {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	for clientConn, clientNickname := range clients {
		if clientConn != nil && clientNickname != senderNickname {
			_, err := clientConn.Write([]byte(senderNickname + ": " + message + newline))
			if err != nil {
				fmt.Println("Failed to send message to client:", err)
			}
		} else if clientConn == senderConn {
			// 보낸 클라이언트에게도 메시지를 표시
			_, err := clientConn.Write([]byte("You: " + message + newline))
			if err != nil {
				fmt.Println("Failed to send message to client:", err)
			}
		}
	}
}

// 클라이언트 리스트와 닉네임을 가져오는 함수
func getClientList() string {
	clientsLock.Lock()
	defer clientsLock.Unlock()

	clientList := ""
	for _, clientNickname := range clients {
		clientList += clientNickname + "\n"
	}
	return clientList
}
