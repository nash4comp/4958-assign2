package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec" // exec 패키지 추가
	"strings"
	"sync"
)

// 클라이언트와 서버 간의 통신 프로토콜 정의
const (
	NickChangeCommand = "/NICK "
	MessageCommand    = "/MSG "
	ListCommand       = "/LIST"
)

var (
	clients     = make(map[net.Conn]string) // 연결된 클라이언트와 닉네임 매핑
	clientsLock sync.Mutex                  // 클라이언트 매핑을 위한 뮤텍스
	newline     = "\n\r"
)

func main() {
	clearScreen()
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
	hasNickname := false

	// 클라이언트에게 연결 성공 메시지 안내 전송
	welcomeMessage := "Welcome to the chat server! Use /NICK <nickname> to set your nickname.\n\r"
	_, err := conn.Write([]byte(welcomeMessage))
	if err != nil {
		fmt.Println("Failed to send welcome message to client: ", err)
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
				// 클라이언트의 닉네임을 저장
				clientsLock.Lock()
				clients[conn] = nickname
				clientsLock.Unlock()
				// 닉네임 설정 성공 메시지 전송
				response := "Your nickname is now set to: " + nickname + newline
				hasNickname = true
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
					fmt.Println("Failed to send nickname in use message to client: ", err)
					return
				}
			}
		} else if message == ListCommand && hasNickname {
			// 클라이언트 리스트와 닉네임 출력
			clientList := getClientList()
			response := "Client List: " + clientList + newline
			_, err := conn.Write([]byte(response))
			if err != nil {
				fmt.Println("Failed to send client list to client: ", err)
				return
			}
		} else {
			response := "You need to set your nickname at first." + newline
			_, err := conn.Write([]byte(response))
			if err != nil {
				fmt.Println("Failed to send nickname in use message to client: ", err)
				return
			}
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

// 클라이언트 리스트와 닉네임을 가져오는 함수
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
	cmd := exec.Command("clear") // macOS 및 Linux에서 clear 명령어 실행
	cmd.Stdout = os.Stdout
	cmd.Run()
}
