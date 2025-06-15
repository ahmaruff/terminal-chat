# Terminal Chat

A real-time multi-client chat server built with Go, featuring room-based messaging and a simple terminal interface.

## Features

- **Multi-client support** - Handle multiple users simultaneously
- **Room-based chat** - Create and join different chat rooms
- **Real-time messaging** - Instant message delivery using TCP connections
- **Auto-room creation** - Rooms are created automatically when joined
- **User management** - Duplicate username handling with retry mechanism
- **Join/leave notifications** - See when users enter or leave rooms
- **Command system** - Full set of chat commands
- **Clean ASCII interface** - Pretty terminal output without external dependencies

## Commands

| Command | Description |
|---------|-------------|
| `/help` | Show available commands |
| `/rooms` | List all available rooms with user counts |
| `/join <room>` | Join or create a chat room |
| `/leave` | Leave current room |
| `/who` | Show users in current room |
| `/quit` | Disconnect from server |

## Getting Started

### Prerequisites

- Go 1.18 or higher
- Terminal/Command prompt

### Installation

1. Clone or download the project files
2. Navigate to the project directory

### Running the Server

```bash
cd server
go run server.go
```

The server will start listening on port `:8080` and create a default "general" room.

### Connecting Clients

Open new terminal windows and run:

```bash
cd client
go run client.go
```

Each client will connect to `localhost:8080` and prompt for a username.

## Usage Example

### Server Output
```
Chat server started on :8080
New client connected!
Client connected with username: alice
New client connected!
Client connected with username: bob
```

### Client Experience
```
Connected to chat server!
Enter Username: alice
===============================================================
|                    TERMINAL CHAT SERVER                    |
===============================================================
Connected to room: general
Type messages or commands (type /help for command list)
===============================================================

<bob> Hello everyone!
*** charlie joined the room ***
<charlie> Hey there!

/who
Users in room 'general' (3):
==========================
* alice (you)
* bob
* charlie
==========================

/join tech-talk
[Server] Created new room: tech-talk
--- Joined room: tech-talk ---
Users in room (1):
* alice (you)
------------------------
```

## Architecture

The chat system consists of:

- **Server** (`server.go`) - Manages clients, rooms, and message broadcasting
- **Client** (`client.go`) - Connects to server and handles user input/output

### Key Components

- **TCP Server** - Listens for client connections on port 8080
- **Goroutines** - Each client connection runs in its own goroutine for concurrency
- **Room Management** - Dynamic room creation and user assignment
- **Message Broadcasting** - Real-time message delivery to room members
- **Command Processing** - Parse and execute chat commands

## Technical Details

### Data Structures

```go
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
```

### Network Protocol

- **Transport**: TCP on port 8080
- **Message Format**: Plain text with newline delimiters
- **Commands**: Text strings starting with `/`
- **Broadcasting**: Server forwards messages to all room members

## Project Structure

```
terminal-chat/
├── server/
│   └── server.go      # Server implementation with main()
├── client/
│   └── client.go      # Client implementation with main()
├── LICENSE            # License File (MIT)
├── README.md          # This file
├── .gitignore         # git
└── go.mod             # Go module file (optional)
```

## Development

This project demonstrates several Go concepts:

- **Network Programming** - TCP server/client with `net` package
- **Concurrency** - Goroutines for handling multiple clients
- **Data Structures** - Maps and pointers for efficient lookups
- **Error Handling** - Graceful handling of network errors and edge cases
- **String Processing** - Command parsing and message formatting

## Testing

1. Start the server: `cd server && go run server.go`
2. Open multiple terminals and run: `cd client && go run client.go`
3. Try different usernames, rooms, and commands
4. Test edge cases like duplicate names, non-existent rooms, etc.

## Limitations

- **In-memory storage** - All data is lost when server restarts
- **No persistence** - Chat history is not saved
- **No authentication** - Simple username-based identification
- **Single server** - No clustering or load balancing
- **No encryption** - Messages are sent in plain text

## Future Enhancements

Potential improvements for learning:

- Add message history/logging
- Implement private messaging
- Add room moderation features
- Create a web interface
- Add user authentication
- Implement message encryption
- Add file sharing capabilities

## License

MIT License - see [LICENSE](./LICENSE) file for details.

## Contributing

This was built as a weekend learning project. Feel free to fork and experiment!

---

**Built with Go** - No external dependencies required!
