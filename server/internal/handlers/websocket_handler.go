package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan interface{}
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
}

// NewWebSocketHandler 创建WebSocket处理器
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan interface{}),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
	}
}

// HandleWebSocket WebSocket连接处理
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源，生产环境应该限制
		},
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// 注册客户端
	h.register <- conn

	// 启动goroutine处理消息
	go h.handleMessages(conn)
}

// handleMessages 处理WebSocket消息
func (h *WebSocketHandler) handleMessages(conn *websocket.Conn) {
	defer func() {
		h.unregister <- conn
		conn.Close()
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// 处理接收到的消息
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}

		// 根据消息类型处理
		h.handleMessage(conn, msg)
	}
}

// handleMessage 处理具体消息
func (h *WebSocketHandler) handleMessage(conn *websocket.Conn, msg map[string]interface{}) {
	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}

	switch msgType {
	case "ping":
		// 响应ping消息
		response := map[string]interface{}{
			"type": "pong",
			"data": "pong",
		}
		h.sendMessage(conn, response)

	case "subscribe":
		// 订阅特定频道
		channel, ok := msg["channel"].(string)
		if ok {
			response := map[string]interface{}{
				"type":    "subscribed",
				"channel": channel,
			}
			h.sendMessage(conn, response)
		}

	default:
		// 未知消息类型
		response := map[string]interface{}{
			"type":  "error",
			"error": "unknown message type",
		}
		h.sendMessage(conn, response)
	}
}

// sendMessage 发送消息到指定连接
func (h *WebSocketHandler) sendMessage(conn *websocket.Conn, message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Failed to send message: %v", err)
	}
}

// BroadcastMessage 广播消息给所有客户端
func (h *WebSocketHandler) BroadcastMessage(message interface{}) {
	h.broadcast <- message
}

// Start 启动WebSocket处理器
func (h *WebSocketHandler) Start() {
	go func() {
		for {
			select {
			case client := <-h.register:
				h.mutex.Lock()
				h.clients[client] = true
				h.mutex.Unlock()
				log.Printf("Client connected. Total clients: %d", len(h.clients))

			case client := <-h.unregister:
				h.mutex.Lock()
				delete(h.clients, client)
				h.mutex.Unlock()
				log.Printf("Client disconnected. Total clients: %d", len(h.clients))

			case message := <-h.broadcast:
				h.mutex.RLock()
				for client := range h.clients {
					h.sendMessage(client, message)
				}
				h.mutex.RUnlock()
			}
		}
	}()
}

// GetClientCount 获取客户端数量
func (h *WebSocketHandler) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}
