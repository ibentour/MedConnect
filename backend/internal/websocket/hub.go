// Package websocket provides WebSocket hub and client management for real-time notifications.
package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"medconnect-oriental/backend/internal/middleware"
)

// Event types for WebSocket messages
const (
	EventNewReferral      = "new_referral"
	EventReferralUpdate   = "referral_updated"
	EventReferralRedirect = "referral_redirected"
)

// Upgrader is a gorilla/websocket upgrader configuration
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients by user ID
	clients map[uuid.UUID]map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to specific users or departments
	broadcast chan *Message

	// Mutex for thread-safe access
	mu sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type      string      `json:"type"`
	Event     string      `json:"event"`
	Payload   interface{} `json:"payload"`
	TargetID  *uuid.UUID  `json:"target_id,omitempty"` // For direct user messages
	DeptID    *uuid.UUID  `json:"dept_id,omitempty"`   // For department broadcast
	CreatedAt time.Time   `json:"created_at"`
}

// Client represents a WebSocket client connection
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID uuid.UUID
	deptID *uuid.UUID
	role   string
}

// GinWebSocket returns a gin handler for WebSocket connections
// It extracts JWT from query parameter and validates it
func GinWebSocket(hub *Hub, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get token from query parameter first (for WebSocket connections)
		tokenString := c.Query("token")
		if tokenString == "" {
			// Fall back to Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 {
					tokenString = parts[1]
				}
			}
		}

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is required"})
			return
		}

		// Validate the JWT token using middleware's function
		claims, err := middleware.ValidateTokenString(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		userID := claims.UserID

		// Get department ID if user has one
		var deptID *uuid.UUID
		if claims.DeptID != nil {
			deptID = claims.DeptID
		}

		role := string(claims.Role)

		// Upgrade to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("[WS] Failed to upgrade connection: %v", err)
			return
		}

		// Create client
		client := &Client{
			hub:    hub,
			conn:   conn,
			send:   make(chan []byte, 256),
			userID: userID,
			deptID: deptID,
			role:   role,
		}

		hub.register <- client

		// Start client handlers
		go client.writePump()
		go client.readPump()
	}
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.userID] == nil {
				h.clients[client.userID] = make(map[*Client]bool)
			}
			h.clients[client.userID][client] = true
			h.mu.Unlock()
			log.Printf("[WS] Client registered: user=%s, role=%s", client.userID, client.role)

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.userID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)
					if len(clients) == 0 {
						delete(h.clients, client.userID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("[WS] Client unregistered: user=%s", client.userID)

		case message := <-h.broadcast:
			h.dispatchMessage(message)
		}
	}
}

// dispatchMessage sends a message to the appropriate clients
func (h *Hub) dispatchMessage(msg *Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if msg.TargetID != nil {
		// Direct message to specific user
		if clients, ok := h.clients[*msg.TargetID]; ok {
			data, _ := json.Marshal(msg)
			for client := range clients {
				select {
				case client.send <- data:
				default:
					// Client buffer full, close connection
					close(client.send)
					delete(clients, client)
				}
			}
		}
	} else if msg.DeptID != nil {
		// Broadcast to all clients in a department
		for _, clients := range h.clients {
			for client := range clients {
				if client.deptID != nil && *client.deptID == *msg.DeptID {
					data, _ := json.Marshal(msg)
					select {
					case client.send <- data:
					default:
						close(client.send)
						delete(clients, client)
					}
				}
			}
		}
	}
}

// BroadcastToDepartment sends a message to all users in a department
func (h *Hub) BroadcastToDepartment(deptID uuid.UUID, event string, payload interface{}) {
	msg := &Message{
		Type:      "broadcast",
		Event:     event,
		Payload:   payload,
		DeptID:    &deptID,
		CreatedAt: time.Now(),
	}
	h.broadcast <- msg
}

// BroadcastToUser sends a message to a specific user
func (h *Hub) BroadcastToUser(userID uuid.UUID, event string, payload interface{}) {
	msg := &Message{
		Type:      "direct",
		Event:     event,
		Payload:   payload,
		TargetID:  &userID,
		CreatedAt: time.Now(),
	}
	h.broadcast <- msg
}

// GetConnectedUsers returns the count of connected users
func (h *Hub) GetConnectedUsers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	count := 0
	for _, clients := range h.clients {
		count += len(clients)
	}
	return count
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] Error: %v", err)
			}
			break
		}

		// Handle incoming messages (e.g., ping/pong, subscription updates)
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Handle ping
		if msg["type"] == "ping" {
			pong, _ := json.Marshal(map[string]interface{}{
				"type": "pong",
				"time": time.Now().Unix(),
			})
			c.send <- pong
		}
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte(""))
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, []byte("{}")); err != nil {
				return
			}
		}
	}
}
