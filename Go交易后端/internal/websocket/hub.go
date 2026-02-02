package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
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

	// 优化: 行情消息缓冲区（用于聚合推送）
	marketBuffer      []interface{}
	marketBufferMutex sync.Mutex
	lastMarketFlush   time.Time
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
	subMutex      sync.RWMutex // Protects subscriptions map

	// Closed flag to prevent double close
	closed uint32 // Use atomic operations
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
	h := &Hub{
		broadcast:       make(chan []byte),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		clients:         make(map[*Client]bool),
		marketBuffer:    make([]interface{}, 0, 20),
		lastMarketFlush: time.Now(),
	}
	// 启动定时刷新协程
	go h.startMarketBufferFlusher()
	return h
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
				client.closeOnce()
			}
			h.mutex.Unlock()
			logger.Info("Client unregistered:", client.conn.RemoteAddr())

		case message := <-h.broadcast:
			// Collect clients to delete outside of RLock
			h.mutex.RLock()
			var clientsToDelete []*Client
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					// Channel full, mark for deletion
					clientsToDelete = append(clientsToDelete, client)
				}
			}
			h.mutex.RUnlock()

			// Delete clients after releasing RLock
			if len(clientsToDelete) > 0 {
				h.mutex.Lock()
				for _, client := range clientsToDelete {
					if _, ok := h.clients[client]; ok {
						delete(h.clients, client)
						client.closeOnce()
					}
				}
				h.mutex.Unlock()
			}
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

	// Collect clients to delete outside of RLock
	h.mutex.RLock()
	var clientsToDelete []*Client
	for client := range h.clients {
		if client.isSubscribed(channel) {
			select {
			case client.send <- msgBytes:
			default:
				// Channel full, mark for deletion
				clientsToDelete = append(clientsToDelete, client)
			}
		}
	}
	h.mutex.RUnlock()

	// Delete clients after releasing RLock
	if len(clientsToDelete) > 0 {
		h.mutex.Lock()
		for _, client := range clientsToDelete {
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				client.closeOnce()
			}
		}
		h.mutex.Unlock()
	}
}

// Subscribe adds a subscription for a client
func (c *Client) Subscribe(channel string) {
	c.subMutex.Lock()
	defer c.subMutex.Unlock()
	c.subscriptions[channel] = true
}

// Unsubscribe removes a subscription for a client
func (c *Client) Unsubscribe(channel string) {
	c.subMutex.Lock()
	defer c.subMutex.Unlock()
	delete(c.subscriptions, channel)
}

// isSubscribed checks if client is subscribed to a channel
func (c *Client) isSubscribed(channel string) bool {
	c.subMutex.RLock()
	defer c.subMutex.RUnlock()
	_, exists := c.subscriptions[channel]
	return exists
}

// closeOnce safely closes the client's send channel only once
func (c *Client) closeOnce() {
	if atomic.CompareAndSwapUint32(&c.closed, 0, 1) {
		close(c.send)
	}
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
// 优化: 使用消息聚合，每80ms批量推送一次
func BroadcastMarketData(data interface{}) {
	hub := GetHub()
	hub.bufferMarketData(data)
}

// bufferMarketData 将行情数据加入缓冲区
func (h *Hub) bufferMarketData(data interface{}) {
	h.marketBufferMutex.Lock()
	defer h.marketBufferMutex.Unlock()

	h.marketBuffer = append(h.marketBuffer, data)

	// 优化: 缓冲区满8条或距上次推送超过80ms，立即推送
	if len(h.marketBuffer) >= 8 || time.Since(h.lastMarketFlush) > 80*time.Millisecond {
		h.flushMarketBufferLocked()
	}
}

// startMarketBufferFlusher 启动定时刷新协程
func (h *Hub) startMarketBufferFlusher() {
	ticker := time.NewTicker(80 * time.Millisecond) // 优化: 从100ms降到80ms
	defer ticker.Stop()

	for range ticker.C {
		h.marketBufferMutex.Lock()
		if len(h.marketBuffer) > 0 {
			h.flushMarketBufferLocked()
		}
		h.marketBufferMutex.Unlock()
	}
}

// flushMarketBufferLocked 刷新缓冲区（需要持有锁）
func (h *Hub) flushMarketBufferLocked() {
	if len(h.marketBuffer) == 0 {
		return
	}

	// 批量推送
	msg := Message{
		Type:      "market_batch",
		Channel:   "market",
		Data:      h.marketBuffer,
		Timestamp: time.Now().UnixMilli(), // 使用毫秒级时间戳
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		logger.Errorf("Failed to marshal market batch: %v", err)
		h.marketBuffer = h.marketBuffer[:0]
		h.lastMarketFlush = time.Now()
		return
	}

	// 发送到订阅了market频道的客户端
	h.mutex.RLock()
	for client := range h.clients {
		if client.isSubscribed("market") {
			select {
			case client.send <- msgBytes:
			default:
				// Channel full, skip this client
			}
		}
	}
	h.mutex.RUnlock()

	// 清空缓冲区
	h.marketBuffer = h.marketBuffer[:0]
	h.lastMarketFlush = time.Now()
}

// BroadcastKlineData broadcasts Kline data to subscribers
func BroadcastKlineData(symbol string, data interface{}) {
	hub := GetHub()
	// 同时广播到全局频道和交易对特定频道
	hub.SendToChannel("kline", data)
	hub.SendToChannel("kline:"+symbol, data)
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

// BroadcastToUser sends a message to a specific user based on their wallet channel
func BroadcastToUser(userID uint64, data interface{}) {
	hub := GetHub()
	channel := fmt.Sprintf("wallet:%d", userID)
	hub.SendToChannel(channel, data)
}
