package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Connected to chat server!")

	// Read from server and user input simultaneously
	go readFromServer(conn)

	// Read from user and send to server
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		message := scanner.Text()
		conn.Write([]byte(message + "\n"))
	}
}

func readFromServer(conn net.Conn) {
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Server disconnected")
			return
		}
		fmt.Print(string(buffer[:n]))
	}
}
