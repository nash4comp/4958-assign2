package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

var (
	clients     = make(map[net.Conn]string) // 연결된 클라이언트와 닉네임 매핑
	clientsLock sync.Mutex                  // 클라이언트 매핑을 위한 뮤텍스
	newline     = "\r\n\x1b[1G"
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

	// 클라이언트에게 연결 성공 메시지 및 닉네임 설정 안내 전송
	welcomeMessage := "Welcome to the chat server! Use /NICK <nickname> to set your nickname.\n\r"
	_, err := conn.Write([]byte(welcomeMessage))
	if err != nil {
		fmt.Println("Failed to send welcome message to client:", err)
		return
	}

	var nickname string

	// 클라이언트가 닉네임을 설정할 때까지 대기
	for {
		// 줄 단위로 입력을 받을 때까지 대기
		input, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Failed to read client message:", err)
			return
		}

		message := strings.TrimSpace(input)

		if strings.HasPrefix(message, "/NICK ") {
			newNickname := strings.TrimPrefix(message, "/NICK ")
			if isNicknameAvailable(newNickname) {
				nickname = newNickname
				// 닉네임 설정 성공 메시지 전송
				response := "Your nickname is now set to: " + nickname
				_, err := conn.Write([]byte(response + newline))
				if err != nil {
					fmt.Println("Failed to send nickname confirmation to client:", err)
					return
				}
				// 닉네임 설정 완료 후 메시지 수신 및 처리
				break
			} else {
				// 닉네임이 이미 사용 중인 경우 메시지 전송
				response := "Nickname already in use. Choose a different nickname."
				_, err := conn.Write([]byte(response + newline))
				if err != nil {
					fmt.Println("Failed to send nickname in use message to client:", err)
					return
				}
			}
		} else {
			// 닉네임 설정을 기다리는 동안 다른 메시지는 무시
			response := "Set your nickname using /NICK <nickname>"
			_, err := conn.Write([]byte(response + newline))
			if err != nil {
				fmt.Println("Failed to send instruction to set nickname to client:", err)
				return
			}
		}
	}

	// 클라이언트와의 메시지 송수신 처리 (클라이언트가 연결을 종료할 때까지)
	for {
		// 줄 단위로 입력을 받을 때까지 대기
		input, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected:", err)
			// 클라이언트가 연결을 종료한 경우 처리
			handleClientDisconnect(conn, nickname)
			return
		}

		message := strings.TrimSpace(input)

		// 메시지를 다른 클라이언트에게 브로드캐스트 또는 저장하는 로직 추가
		// 예를 들어, broadcastMessage 함수를 호출하여 메시지를 모든 클라이언트에게 전달
		broadcastMessage(message, nickname, conn)
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
