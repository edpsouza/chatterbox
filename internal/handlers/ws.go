package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/edpsouza/chatterbox/internal/models"
	"github.com/edpsouza/chatterbox/internal/store"
	"github.com/gorilla/websocket"
)

// Upgrader configures the WebSocket upgrade parameters.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development; restrict in production!
		return true
	},
}

// Client represents a single WebSocket connection.
type Client struct {
	Conn          *websocket.Conn
	Send          chan []byte
	UserID        string
	Username      string
	Authenticated bool
}

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.Mutex
}

// NewHub initializes a new Hub.
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run starts the Hub's main loop.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
		case message := <-h.Broadcast:
			h.mu.Lock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

// ServeWS handles WebSocket requests from clients, authenticating with username/password as first message.
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	client := &Client{
		Conn:          conn,
		Send:          make(chan []byte, 256),
		Authenticated: false,
	}
	hub.Register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump(hub)
}

// readPump reads messages from the WebSocket connection and broadcasts them.
func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
		// Set user status to offline and update last_seen
		storeInstance, err := getStoreInstance()
		if err == nil && c.Username != "" {
			_ = storeInstance.SetUserStatus(c.Username, "offline")
			_ = storeInstance.SetUserLastSeenNow(c.Username)
		}
	}()

	authChecked := false

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		if !authChecked {
			// Expect first message to be JSON: {"username":"...","password":"..."}
			type AuthMsg struct {
				Username string `json:"username"`
				Password string `json:"password"`
			}
			var auth AuthMsg
			if err := json.Unmarshal(message, &auth); err != nil {
				c.Conn.WriteMessage(websocket.TextMessage, []byte("Invalid auth message format"))
				break
			}
			if auth.Username == "" || auth.Password == "" {
				c.Conn.WriteMessage(websocket.TextMessage, []byte("Username and password required"))
				break
			}

			// Authenticate user
			storeInstance, err := getStoreInstance()
			if err != nil {
				c.Conn.WriteMessage(websocket.TextMessage, []byte("Server error"))
				break
			}
			user, err := storeInstance.GetUserByUsername(auth.Username)
			if err != nil || user == nil {
				c.Conn.WriteMessage(websocket.TextMessage, []byte("Invalid credentials"))
				break
			}
			ok, err := models.VerifyPassword(user.Password, auth.Password)
			if err != nil || !ok {
				c.Conn.WriteMessage(websocket.TextMessage, []byte("Invalid credentials"))
				break
			}
			c.UserID = strconv.FormatInt(user.ID, 10)
			c.Username = user.Username
			c.Authenticated = true

			// Set user status to online
			if storeInstance != nil {
				_ = storeInstance.SetUserStatus(c.Username, "online")
			}

			c.Conn.WriteMessage(websocket.TextMessage, []byte("Authenticated"))
			authChecked = true
			continue
		}

		if !c.Authenticated {
			c.Conn.WriteMessage(websocket.TextMessage, []byte("Not authenticated"))
			break
		}

		// Parse message as JSON: {"to":"recipient_username","ciphertext":"..."}
		type ChatMsg struct {
			To         string `json:"to"`
			Ciphertext string `json:"ciphertext"`
		}
		var chatMsg ChatMsg
		if err := json.Unmarshal(message, &chatMsg); err != nil {
			c.Conn.WriteMessage(websocket.TextMessage, []byte("Invalid chat message format"))
			continue
		}
		if chatMsg.To == "" || chatMsg.Ciphertext == "" {
			c.Conn.WriteMessage(websocket.TextMessage, []byte("Recipient and ciphertext required"))
			continue
		}

		// Store ciphertext to DB
		storeInstance, err := getStoreInstance()
		if err == nil {
			_ = storeInstance.CreateMessage(
				func() int64 {
					id, _ := strconv.ParseInt(c.UserID, 10, 64)
					return id
				}(),
				c.Username,
				chatMsg.To,
				chatMsg.Ciphertext,
			)
		}

		// Route message only to intended recipient
		hub.mu.Lock()
		var recipientClient *Client
		for client := range hub.Clients {
			if client.Username == chatMsg.To && client.Authenticated {
				recipientClient = client
				break
			}
		}
		hub.mu.Unlock()
		if recipientClient != nil {
			recipientClient.Send <- []byte(chatMsg.Ciphertext)
		} else {
			c.Conn.WriteMessage(websocket.TextMessage, []byte("Recipient not connected"))
		}
	}
}

// writePump writes messages from the hub to the WebSocket connection.
func (c *Client) writePump() {
	defer c.Conn.Close()
	for msg := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

// getStoreInstance returns the global store instance from main package via a package-level variable.
// This is a workaround for accessing the store from the handler.
var storeInstanceGlobal *store.Store

func SetStoreInstance(s *store.Store) {
	storeInstanceGlobal = s
}

func getStoreInstance() (*store.Store, error) {
	if storeInstanceGlobal == nil {
		return nil, errors.New("store instance not set")
	}
	return storeInstanceGlobal, nil
}
