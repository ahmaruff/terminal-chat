package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

type Client struct {
	Name        string
	Conn        net.Conn
	CurrentRoom string
}

type Room struct {
	Name    string
	Clients map[string]*Client
}

type Server struct {
	Rooms   map[string]*Room
	Clients map[string]*Client
	mu      sync.RWMutex
}

func (s *Server) AddClient(name string, conn net.Conn, roomName string) (Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Rooms[roomName] == nil {
		return Client{}, fmt.Errorf("Room Not Found!")
	}

	if s.Clients[name] != nil {
		return Client{}, fmt.Errorf("Client already Exists")
	}

	c := Client{
		Name:        name,
		Conn:        conn,
		CurrentRoom: roomName,
	}

	room := s.Rooms[roomName]
	room.Clients[c.Name] = &c

	s.Clients[c.Name] = &c

	return c, nil
}

func (s *Server) RemoveClient(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Clients[name] == nil {
		return fmt.Errorf("Client not found")
	}

	c := s.Clients[name]

	if c.CurrentRoom != "" && s.Rooms[c.CurrentRoom] != nil {
		delete(s.Rooms[c.CurrentRoom].Clients, name)
	}

	delete(s.Clients, name)
	return nil
}

func (s *Server) JoinRoom(clientName, roomName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Rooms[roomName] == nil {
		return fmt.Errorf("Room Not Found!")
	}

	if s.Clients[clientName] == nil {
		return fmt.Errorf("Client Not Found")
	}

	r := s.Rooms[roomName]

	if r.Clients[clientName] != nil {
		return nil
	}

	currentRoom := s.Clients[clientName].CurrentRoom
	if currentRoom != "" && s.Rooms[currentRoom] != nil {
		delete(s.Rooms[currentRoom].Clients, clientName)
	}

	r.Clients[clientName] = s.Clients[clientName]

	s.Clients[clientName].CurrentRoom = roomName

	return nil
}

func (s *Server) LeaveRoom(clientName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Clients[clientName] == nil {
		return fmt.Errorf("Client Not Found")
	}

	client := s.Clients[clientName]
	currentRoom := client.CurrentRoom

	if currentRoom == "" {
		return nil
	}

	if s.Rooms[currentRoom] == nil {
		return fmt.Errorf("Room Not Found!")
	}

	r := s.Rooms[currentRoom]

	if r.Clients[clientName] == nil {
		return nil
	}

	delete(r.Clients, clientName)

	s.Clients[clientName].CurrentRoom = ""

	return nil
}

func (s *Server) BroadcastToRoom(clientName, message string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.Clients[clientName] == nil {
		return fmt.Errorf("Client Not Found")
	}

	client := s.Clients[clientName]
	currentRoom := client.CurrentRoom

	if currentRoom == "" {
		return fmt.Errorf("Room Not Found!")
	}

	if s.Rooms[currentRoom] == nil {
		return fmt.Errorf("Room Not Found!")
	}

	r := s.Rooms[currentRoom]

	formattedMessage := fmt.Sprintf("<%s> %s\n", clientName, message)

	for name, client := range r.Clients {
		if clientName != name {
			client.Conn.Write([]byte(formattedMessage))
		}
	}

	return nil
}

func initServer() *Server {
	r := initRoom()

	rooms := make(map[string]*Room)
	rooms[r.Name] = r

	s := Server{
		Rooms:   rooms,
		Clients: make(map[string]*Client),
	}

	return &s
}

func initRoom() *Room {
	r := Room{
		Name:    "general",
		Clients: make(map[string]*Client),
	}

	return &r

}

func handleClient(conn net.Conn, s *Server) {
	defer conn.Close()

	// Username retry loop
	var username string

	for {
		conn.Write([]byte("Enter Username: "))

		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("Error reading username: ", err)
			return
		}

		username = strings.TrimSpace(string(buffer[:n]))
		if username == "" {
			conn.Write([]byte("Username cannot be empty. Try again.\n"))
			continue
		}

		// Try to add client
		_, err = s.AddClient(username, conn, "general")
		if err != nil {
			// Username taken, ask for retry
			conn.Write([]byte(fmt.Sprintf("Error: %s\nPlease try a different username.\n", err.Error())))
			continue
		}

		// Success! Break out of loop
		break
	}

	fmt.Printf("Client connected with username: %s\n", username)

	conn.Write([]byte("===============================================================\n"))
	conn.Write([]byte("|                    TERMINAL CHAT SERVER                    |\n"))
	conn.Write([]byte("===============================================================\n"))
	conn.Write([]byte("Connected to room: general\n"))
	conn.Write([]byte("Type messages or commands (type /help for command list)\n"))
	conn.Write([]byte("---------------------------------------------------------------\n"))

	// Message loop - keep reading until client disconnects
	for {
		bufferMsg := make([]byte, 1024)
		msg, err := conn.Read(bufferMsg)
		if err != nil {
			fmt.Println("Client disconnected: ", username)
			s.RemoveClient(username)
			break
		}

		msgStr := strings.TrimSpace(string(bufferMsg[:msg]))
		if msgStr == "" {
			continue
		}

		if strings.HasPrefix(msgStr, "/") {
			handleCommand(msgStr, username, conn, s)
		} else {
			s.BroadcastToRoom(username, msgStr)
		}
	}
}

func handleCommand(command, username string, conn net.Conn, s *Server) {
	parts := strings.Split(command, " ")
	cmd := parts[0]

	switch cmd {
	case "/rooms":
		conn.Write([]byte("Available Rooms:\n"))
		conn.Write([]byte("================\n"))
		for roomName, room := range s.Rooms {
			userCount := len(room.Clients)
			conn.Write([]byte(fmt.Sprintf("* %s (%d users)\n", roomName, userCount)))
		}
		conn.Write([]byte("================\n"))

	case "/join":
		if len(parts) < 2 {
			conn.Write([]byte("[Error] Usage: /join <room_name>\n"))
			return
		}
		roomName := parts[1]

		if s.Rooms[roomName] == nil {
			newRoom := &Room{
				Name:    roomName,
				Clients: make(map[string]*Client),
			}
			s.Rooms[roomName] = newRoom
			conn.Write([]byte(fmt.Sprintf("[Server] Created new room: %s\n", roomName)))
		}

		err := s.JoinRoom(username, roomName)
		if err != nil {
			conn.Write([]byte(fmt.Sprintf("[Error] %s\n", err.Error())))
			return
		}

		conn.Write([]byte(fmt.Sprintf("--- Joined room: %s ---\n", roomName)))

		room := s.Rooms[roomName]
		conn.Write([]byte(fmt.Sprintf("Users in room (%d):\n", len(room.Clients))))
		for clientName := range room.Clients {
			if clientName == username {
				conn.Write([]byte(fmt.Sprintf("* %s (you)\n", clientName)))
			} else {
				conn.Write([]byte(fmt.Sprintf("* %s\n", clientName)))
			}
		}
		conn.Write([]byte("------------------------\n"))

		notifyMessage := fmt.Sprintf("*** %s joined the room ***\n", username)
		for name, client := range room.Clients {
			if name != username {
				client.Conn.Write([]byte(notifyMessage))
			}
		}

	case "/leave":
		currentRoom := s.Clients[username].CurrentRoom
		if currentRoom == "" {
			conn.Write([]byte("[Info] You are not in any room\n"))
			return
		}

		// Notify others before leaving
		if s.Rooms[currentRoom] != nil {
			notifyMessage := fmt.Sprintf("*** %s left the room ***\n", username)
			for name, client := range s.Rooms[currentRoom].Clients {
				if name != username {
					client.Conn.Write([]byte(notifyMessage))
				}
			}
		}

		err := s.LeaveRoom(username)
		if err != nil {
			conn.Write([]byte(fmt.Sprintf("[Error] %s\n", err.Error())))
			return
		}

		conn.Write([]byte(fmt.Sprintf("--- Left room: %s ---\n", currentRoom)))

	case "/who":
		currentRoom := s.Clients[username].CurrentRoom
		if currentRoom == "" {
			conn.Write([]byte("[Info] You are not in any room\n"))
			return
		}

		room := s.Rooms[currentRoom]
		conn.Write([]byte(fmt.Sprintf("Users in room '%s' (%d):\n", currentRoom, len(room.Clients))))
		conn.Write([]byte("==========================\n"))
		for clientName := range room.Clients {
			if clientName == username {
				conn.Write([]byte(fmt.Sprintf("* %s (you)\n", clientName)))
			} else {
				conn.Write([]byte(fmt.Sprintf("* %s\n", clientName)))
			}
		}
		conn.Write([]byte("==========================\n"))

	case "/help":
		conn.Write([]byte("Available Commands:\n"))
		conn.Write([]byte("==================\n"))
		conn.Write([]byte("/help          - Show this help message\n"))
		conn.Write([]byte("/rooms         - List all available rooms\n"))
		conn.Write([]byte("/join <room>   - Join or create a room\n"))
		conn.Write([]byte("/leave         - Leave current room\n"))
		conn.Write([]byte("/who           - Show users in current room\n"))
		conn.Write([]byte("/quit          - Disconnect from server\n"))
		conn.Write([]byte("==================\n"))
		conn.Write([]byte("Type any message to chat with room members\n"))

	case "/quit":
		// Notify room before quitting
		currentRoom := s.Clients[username].CurrentRoom
		if currentRoom != "" && s.Rooms[currentRoom] != nil {
			notifyMessage := fmt.Sprintf("*** %s disconnected ***\n", username)
			for name, client := range s.Rooms[currentRoom].Clients {
				if name != username {
					client.Conn.Write([]byte(notifyMessage))
				}
			}
		}

		conn.Write([]byte("+-------------------+\n"))
		conn.Write([]byte("|  Thanks for chat! |\n"))
		conn.Write([]byte("+-------------------+\n"))
		s.RemoveClient(username)
		conn.Close()
		return

	default:
		conn.Write([]byte("[Error] Unknown command\n"))
		conn.Write([]byte("Type /help for available commands\n"))
	}
}

func main() {

	server := initServer()

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}

	defer listener.Close()

	fmt.Println("Chat server started on :8080")

	// Accept connections forever
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		fmt.Println("New client connected!")

		// Handle this client in a separate goroutine
		go handleClient(conn, server)
	}

}
