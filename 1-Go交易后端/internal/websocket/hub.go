package websocket

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"exchange-go/config"
	"exchange-go/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Mutex for thread safety
	mutex sync.RWMutex
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Client subscriptions
	subscriptions map[string]bool
}

// Message represents a websocket message
type Message struct {
	Type      string      `json:"type"`
	Channel   string      `json:"channel"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
}

// NewHub creates a new hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			logger.Info("Client registered:", client.conn.RemoteAddr())

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mutex.Unlock()
			logger.Info("Client unregistered:", client.conn.RemoteAddr())

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// SendToAll sends a message to all connected clients
func (h *Hub) SendToAll(msg []byte) {
	select {
	case h.broadcast <- msg:
	default:
		logger.Error("Failed to send message to broadcast channel")
	}
}

// SendToChannel sends a message to clients subscribed to a specific channel
func (h *Hub) SendToChannel(channel string, data interface{}) {
	msg := Message{
		Type:      "channel_message",
		Channel:   channel,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		logger.Errorf("Failed to marshal message: %v", err)
		return
	}

	h.mutex.RLock()
	for client := range h.clients {
		if client.isSubscribed(channel) {
			select {
			case client.send <- msgBytes:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
	h.mutex.RUnlock()
}

// Subscribe adds a subscription for a client
func (c *Client) Subscribe(channel string) {
	c.subscriptions[channel] = true
}

// Unsubscribe removes a subscription for a client
func (c *Client) Unsubscribe(channel string) {
	delete(c.subscriptions, channel)
}

// isSubscribed checks if client is subscribed to a channel
func (c *Client) isSubscribed(channel string) bool {
	_, exists := c.subscriptions[channel]
	return exists
}

// WritePump pumps messages from the hub to the websocket connection.
func (c *Client) WritePump() {
	ticker := time.NewTicker(time.Second * 60) // Ping interval
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(time.Second * 10)); err != nil {
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ReadPump pumps messages from the websocket connection to the hub.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	if err := c.conn.SetReadDeadline(time.Now().Add(time.Second * 60)); err != nil {
		logger.Errorf("Failed to set read deadline: %v", err)
		return
	}
	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(time.Second * 60)); err != nil {
			logger.Errorf("Failed to set read deadline: %v", err)
			return err
		}
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages (subscriptions, unsubscriptions, etc.)
		var incomingMsg Message
		if err := json.Unmarshal(message, &incomingMsg); err != nil {
			logger.Errorf("Failed to unmarshal incoming message: %v", err)
			continue
		}

		switch incomingMsg.Type {
		case "subscribe":
			channel := incomingMsg.Channel
			if channel != "" {
				c.Subscribe(channel)
				// Send confirmation
				confirmMsg := Message{
					Type:      "subscribed",
					Channel:   channel,
					Timestamp: time.Now().Unix(),
				}
				confirmBytes, _ := json.Marshal(confirmMsg)
				select {
				case c.send <- confirmBytes:
				default:
					logger.Error("Failed to send subscription confirmation")
				}
			}
		case "unsubscribe":
			channel := incomingMsg.Channel
			if channel != "" {
				c.Unsubscribe(channel)
				// Send confirmation
				confirmMsg := Message{
					Type:      "unsubscribed",
					Channel:   channel,
					Timestamp: time.Now().Unix(),
				}
				confirmBytes, _ := json.Marshal(confirmMsg)
				select {
				case c.send <- confirmBytes:
				default:
					logger.Error("Failed to send unsubscription confirmation")
				}
			}
		}
	}
}

// WebSocket upgrade configuration
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 开发/测试环境：允许所有跨域
		if config.GlobalConfig.App.Mode == "debug" || config.GlobalConfig.App.Mode == "release" {
			return true
		}

		// 生产环境：检查origin
		origin := r.Header.Get("Origin")
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"https://yourdomain.com",
		}

		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}
		return false
	},
}

// Handler handles websocket requests from the peer.
func Handler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Errorf("WebSocket upgrade failed: %v", err)
		return
	}

	client := &Client{
		hub:           GetHub(),
		conn:          conn,
		send:          make(chan []byte, 256),
		subscriptions: make(map[string]bool),
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()
}

// Global hub instance
var globalHub *Hub
var once sync.Once

// GetHub returns the singleton hub instance
func GetHub() *Hub {
	once.Do(func() {
		globalHub = NewHub()
		go globalHub.Run()
	})
	return globalHub
}

// BroadcastMarketData broadcasts market data to all subscribers
func BroadcastMarketData(data interface{}) {
	hub := GetHub()
	hub.SendToChannel("market", data)
}

// BroadcastKlineData broadcasts Kline data to subscribers
func BroadcastKlineData(data interface{}) {
	hub := GetHub()
	hub.SendToChannel("kline", data)
}

// BroadcastOrderBook broadcasts order book data to subscribers
func BroadcastOrderBook(data interface{}) {
	hub := GetHub()
	hub.SendToChannel("orderbook", data)
}

// BroadcastTrade executes broadcasts trade data to subscribers
func BroadcastTrade(data interface{}) {
	hub := GetHub()
	hub.SendToChannel("trade", data)
}

var newline = []byte{'\n'}
