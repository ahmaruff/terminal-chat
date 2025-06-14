package main

import (
	"fmt"
	"net"
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
}

func (s *Server) AddClient(name string, conn net.Conn, roomName string) (Client, error) {

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

	formattedMessage := fmt.Sprintf("[%s]: %s\n", clientName, message)

	for name, client := range r.Clients {
		if clientName != name {
			client.Conn.Write([]byte(formattedMessage))
		}
	}

	return nil
}

func initServer() Server {
	r := initRoom()

	rooms := make(map[string]*Room)
	rooms[r.Name] = r

	s := Server{
		Rooms:   rooms,
		Clients: make(map[string]*Client),
	}

	return s
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
	fmt.Println("Handling client...")

	// TODO
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
		go handleClient(conn, &server)
	}

}
