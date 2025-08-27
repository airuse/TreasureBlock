package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// ================== 三级分类枚举定义 ==================

// MessageType 第一级别：消息类型
type MessageType string

const (
	MessageTypeEvent        MessageType = "event"        // 实时事件
	MessageTypeNotification MessageType = "notification" // 系统通知
)

// MessageCategory 第二级别：消息分类
type MessageCategory string

const (
	MessageCategoryTransaction MessageCategory = "transaction" // 交易相关
	MessageCategoryBlock       MessageCategory = "block"       // 区块相关
	MessageCategoryAddress     MessageCategory = "address"     // 地址相关
	MessageCategoryStats       MessageCategory = "stats"       // 统计信息
	MessageCategoryNetwork     MessageCategory = "network"     // 网络状态
)

// ChainType 区块链类型
type ChainType string

const (
	ChainTypeETH ChainType = "eth" // 以太坊
	ChainTypeBTC ChainType = "btc" // 比特币
)

// ================== 消息结构定义 ==================

// WebSocketMessage WebSocket消息结构（三级分类）
type WebSocketMessage struct {
	Type      MessageType     `json:"type"`      // 第一级别：事件或通知
	Category  MessageCategory `json:"category"`  // 第二级别：数据类型
	Action    string          `json:"action"`    // 第三级别：动作类型（create, update, delete等）
	Data      interface{}     `json:"data"`      // 第四级别：真实数据
	Timestamp int64           `json:"timestamp"` // 时间戳
	Chain     ChainType       `json:"chain"`     // 区块链类型
}

// SubscribeMessage 订阅消息
type SubscribeMessage struct {
	Type     string          `json:"type"`
	Category MessageCategory `json:"category"`
	Chain    ChainType       `json:"chain"`
}

// SubscribeResponse 订阅响应
type SubscribeResponse struct {
	Type     string          `json:"type"`
	Category MessageCategory `json:"category"`
	Chain    ChainType       `json:"chain"`
	Message  string          `json:"message"`
}

// ================== 客户端管理 ==================

// Client 客户端信息
type Client struct {
	conn       *websocket.Conn
	subscribed map[string]bool // 订阅的频道: "category:chain"
	send       chan []byte     // 发送消息的通道
}

// WebSocketHandler WebSocket处理器
type WebSocketHandler struct {
	clients    map[*Client]bool
	broadcast  chan WebSocketMessage
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

// NewWebSocketHandler 创建WebSocket处理器
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan WebSocketMessage),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// HandleWebSocket WebSocket连接处理
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源，生产环境应该限制
		},
		// 添加更多配置选项
		EnableCompression: true,
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// 设置连接参数
	conn.SetReadLimit(512 * 1024)                          // 限制消息大小为512KB
	conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // 设置读取超时
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // 重置读取超时
		return nil
	})

	// 创建客户端
	client := &Client{
		conn:       conn,
		subscribed: make(map[string]bool),
		send:       make(chan []byte, 256),
	}

	log.Printf("New WebSocket client connected from %s", conn.RemoteAddr().String())

	// 注册客户端
	h.register <- client

	// 启动goroutine处理消息
	go h.handleMessages(client)
}

// handleMessages 处理WebSocket消息
func (h *WebSocketHandler) handleMessages(client *Client) {
	defer func() {
		log.Printf("WebSocket client disconnected from %s", client.conn.RemoteAddr().String())
		h.unregister <- client
		client.conn.Close()
		close(client.send)
	}()

	// 启动发送goroutine
	go h.writePump(client)

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket unexpected close error: %v", err)
			} else {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// 检查是否为ping消息（特殊处理）
		messageStr := string(message)
		if messageStr == "ping" {
			// 直接响应pong，不需要JSON解析
			response := map[string]interface{}{
				"type": "pong",
				"data": "pong",
			}
			h.sendMessage(client, response)
			continue
		}

		// 处理接收到的消息
		var msg SubscribeMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to unmarshal message: %v, message: %s", err, messageStr)
			continue
		}

		// 根据消息类型处理
		h.handleMessage(client, msg)
	}
}

// writePump 发送消息的goroutine
func (h *WebSocketHandler) writePump(client *Client) {
	ticker := time.NewTicker(30 * time.Second) // 心跳检测
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			if !ok {
				// 通道已关闭，发送关闭消息
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 单独发送每条消息，避免合并多条消息
			if err := client.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("Failed to send message: %v", err)
				return
			}

		case <-ticker.C:
			// 发送心跳ping消息
			pingMessage := map[string]interface{}{
				"type":      "ping",
				"data":      "ping",
				"timestamp": time.Now().UnixMilli(),
			}

			pingData, err := json.Marshal(pingMessage)
			if err != nil {
				log.Printf("Failed to marshal ping message: %v", err)
				return
			}

			if err := client.conn.WriteMessage(websocket.TextMessage, pingData); err != nil {
				log.Printf("Failed to send ping message: %v", err)
				return
			}
		}
	}
}

// handleMessage 处理具体消息
func (h *WebSocketHandler) handleMessage(client *Client, msg SubscribeMessage) {
	switch msg.Type {
	case "subscribe":
		// 订阅特定频道
		if msg.Category != "" && msg.Chain != "" {
			channelKey := string(msg.Category) + ":" + string(msg.Chain)
			client.subscribed[channelKey] = true

			response := SubscribeResponse{
				Type:     "subscribed",
				Category: msg.Category,
				Chain:    msg.Chain,
				Message:  "Successfully subscribed to " + channelKey,
			}

			h.sendMessage(client, response)
		} else {
			log.Printf("Invalid subscribe message: category=%s, chain=%s", msg.Category, msg.Chain)
		}

	case "unsubscribe":
		// 取消订阅
		if msg.Category != "" && msg.Chain != "" {
			channelKey := string(msg.Category) + ":" + string(msg.Chain)
			delete(client.subscribed, channelKey)

			response := SubscribeResponse{
				Type:     "unsubscribed",
				Category: msg.Category,
				Chain:    msg.Chain,
				Message:  "Successfully unsubscribed from " + channelKey,
			}

			h.sendMessage(client, response)
		}

	case "pong":
		// 忽略pong消息
		return

	case "event":
		// 处理事件消息（如心跳等）
		if msg.Category == "network" {
			// 网络相关事件，如心跳
			// 可以选择响应或不响应
			return
		}
		// 其他事件类型，记录但不报错
		return

	default:
		// 未知消息类型
		response := map[string]interface{}{
			"type":  "error",
			"error": "unknown message type: " + msg.Type,
		}
		h.sendMessage(client, response)
	}
}

// sendMessage 发送消息到指定客户端
func (h *WebSocketHandler) sendMessage(client *Client, message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	select {
	case client.send <- data:
	default:
		close(client.send)
		delete(h.clients, client)
	}
}

// BroadcastMessage 广播消息给所有订阅了对应频道的客户端
func (h *WebSocketHandler) BroadcastMessage(message WebSocketMessage) {
	h.broadcast <- message
}

// BroadcastBlockEvent 广播区块事件
func (h *WebSocketHandler) BroadcastBlockEvent(chain ChainType, blockData interface{}) {
	message := WebSocketMessage{
		Type:      MessageTypeEvent,
		Category:  MessageCategoryBlock,
		Action:    "create", // 默认创建事件
		Data:      blockData,
		Timestamp: time.Now().UnixMilli(),
		Chain:     chain,
	}
	h.BroadcastMessage(message)
}

// BroadcastBlockUpdateEvent 广播区块更新事件
func (h *WebSocketHandler) BroadcastBlockUpdateEvent(chain ChainType, blockData interface{}) {
	message := WebSocketMessage{
		Type:      MessageTypeEvent,
		Category:  MessageCategoryBlock,
		Action:    "update", // 更新事件
		Data:      blockData,
		Timestamp: time.Now().UnixMilli(),
		Chain:     chain,
	}
	h.BroadcastMessage(message)
}

// BroadcastTransactionEvent 广播交易事件
func (h *WebSocketHandler) BroadcastTransactionEvent(chain ChainType, txData interface{}) {
	message := WebSocketMessage{
		Type:      MessageTypeEvent,
		Category:  MessageCategoryTransaction,
		Data:      txData,
		Timestamp: time.Now().UnixMilli(),
		Chain:     chain,
	}
	h.BroadcastMessage(message)
}

// BroadcastStatsEvent 广播统计信息事件
func (h *WebSocketHandler) BroadcastStatsEvent(chain ChainType, statsData interface{}) {
	message := WebSocketMessage{
		Type:      MessageTypeEvent,
		Category:  MessageCategoryStats,
		Data:      statsData,
		Timestamp: time.Now().UnixMilli(),
		Chain:     chain,
	}
	h.BroadcastMessage(message)
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

			case client := <-h.unregister:
				h.mutex.Lock()
				delete(h.clients, client)
				h.mutex.Unlock()

			case message := <-h.broadcast:
				h.mutex.RLock()
				// 构建频道键
				channelKey := string(message.Category) + ":" + string(message.Chain)

				// 只发送给订阅了对应频道的客户端
				sentCount := 0
				for client := range h.clients {
					if client.subscribed[channelKey] {
						h.sendMessage(client, message)
						sentCount++
					}
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

// GetSubscribedClients 获取订阅了特定频道的客户端数量
func (h *WebSocketHandler) GetSubscribedClients(category MessageCategory, chain ChainType) int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	channelKey := string(category) + ":" + string(chain)
	count := 0
	for client := range h.clients {
		if client.subscribed[channelKey] {
			count++
		}
	}
	return count
}
