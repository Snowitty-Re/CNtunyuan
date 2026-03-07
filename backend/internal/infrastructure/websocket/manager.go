package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/Snowitty-Re/CNtunyuan/internal/domain/entity"
	"github.com/Snowitty-Re/CNtunyuan/pkg/logger"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Client WebSocket客户端
type Client struct {
	ID       string
	UserID   string
	OrgID    string
	Conn     *websocket.Conn
	Manager  *Manager
	Send     chan []byte
	IsAlive  bool
	LastPing time.Time
}

// Manager WebSocket管理器
type Manager struct {
	clients    map[string]*Client      // user_id -> Client
	clientsMux sync.RWMutex
	
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	
	upgrader websocket.Upgrader
	
	// 配置
	MaxMessageSize int64
	PingInterval   time.Duration
	PongTimeout    time.Duration
	
	// 统计
	stats Stats
}

// BroadcastMessage 广播消息
type BroadcastMessage struct {
	Message *entity.WebSocketMessage
	UserIDs []string // 为空则广播给所有用户
	OrgID   string   // 组织过滤
}

// Stats WebSocket统计
type Stats struct {
	TotalConnections    int64
	ActiveConnections   int64
	MessagesSent        int64
	MessagesReceived    int64
	BroadcastsSent      int64
}

// NewManager 创建WebSocket管理器
func NewManager() *Manager {
	return &Manager{
		clients:    make(map[string]*Client),
		register:   make(chan *Client, 100),
		unregister: make(chan *Client, 100),
		broadcast:  make(chan *BroadcastMessage, 100),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// 允许跨域，生产环境应配置具体域名
				return true
			},
		},
		MaxMessageSize: 1024 * 1024, // 1MB
		PingInterval:   30 * time.Second,
		PongTimeout:    60 * time.Second,
	}
}

// Start 启动WebSocket管理器
func (m *Manager) Start(ctx context.Context) {
	logger.Info("WebSocket manager started", zap.String("component", "websocket"))
	
	ticker := time.NewTicker(m.PingInterval)
	defer ticker.Stop()
	
	for {
		select {
		case client := <-m.register:
			m.registerClient(client)
			
		case client := <-m.unregister:
			m.unregisterClient(client)
			
		case msg := <-m.broadcast:
			m.handleBroadcast(msg)
			
		case <-ticker.C:
			m.pingAll()
			m.cleanDeadConnections()
			
		case <-ctx.Done():
			m.shutdown()
			return
		}
	}
}

// HandleWebSocket 处理WebSocket连接升级
func (m *Manager) HandleWebSocket(w http.ResponseWriter, r *http.Request, userID, orgID string) {
	conn, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}
	
	client := &Client{
		ID:       generateClientID(),
		UserID:   userID,
		OrgID:    orgID,
		Conn:     conn,
		Manager:  m,
		Send:     make(chan []byte, 256),
		IsAlive:  true,
		LastPing: time.Now(),
	}
	
	// 设置读取限制
	conn.SetReadLimit(m.MaxMessageSize)
	conn.SetReadDeadline(time.Now().Add(m.PongTimeout))
	conn.SetPongHandler(func(string) error {
		client.LastPing = time.Now()
		conn.SetReadDeadline(time.Now().Add(m.PongTimeout))
		return nil
	})
	
	m.register <- client
	
	// 启动读写协程
	go client.writePump()
	go client.readPump()
	
	// 发送连接成功消息
	msg := entity.NewWebSocketMessage("connected", "连接成功", "WebSocket连接已建立")
	msg.Data = map[string]interface{}{
		"client_id": client.ID,
		"user_id":   userID,
	}
	m.SendToUser(userID, msg)
}

// registerClient 注册客户端
func (m *Manager) registerClient(client *Client) {
	m.clientsMux.Lock()
	defer m.clientsMux.Unlock()
	
	// 如果用户已有连接，先断开旧连接
	if oldClient, exists := m.clients[client.UserID]; exists {
		oldClient.close()
	}
	
	m.clients[client.UserID] = client
	m.stats.TotalConnections++
	m.stats.ActiveConnections = int64(len(m.clients))
	
	logger.Info("WebSocket client registered", 
		zap.String("user_id", client.UserID), 
		zap.String("client_id", client.ID),
		zap.Int64("active_count", m.stats.ActiveConnections))
}

// unregisterClient 注销客户端
func (m *Manager) unregisterClient(client *Client) {
	m.clientsMux.Lock()
	defer m.clientsMux.Unlock()
	
	if _, exists := m.clients[client.UserID]; exists {
		delete(m.clients, client.UserID)
		client.close()
		m.stats.ActiveConnections = int64(len(m.clients))
		
		logger.Info("WebSocket client unregistered", 
			zap.String("user_id", client.UserID),
			zap.String("client_id", client.ID),
			zap.Int64("active_count", m.stats.ActiveConnections))
	}
}

// handleBroadcast 处理广播消息
func (m *Manager) handleBroadcast(msg *BroadcastMessage) {
	m.clientsMux.RLock()
	defer m.clientsMux.RUnlock()
	
	data, err := json.Marshal(msg.Message)
	if err != nil {
		logger.Error("Failed to marshal broadcast message", zap.Error(err))
		return
	}
	
	// 指定用户列表
	if len(msg.UserIDs) > 0 {
		for _, userID := range msg.UserIDs {
			if client, exists := m.clients[userID]; exists {
				select {
				case client.Send <- data:
					m.stats.MessagesSent++
				default:
					// 发送缓冲区满，关闭连接
					go m.unregisterClient(client)
				}
			}
		}
		m.stats.BroadcastsSent++
		return
	}
	
	// 广播给所有用户（可按组织过滤）
	for _, client := range m.clients {
		// 如果指定了组织，进行过滤
		if msg.OrgID != "" && client.OrgID != msg.OrgID {
			continue
		}
		
		select {
		case client.Send <- data:
			m.stats.MessagesSent++
		default:
			go m.unregisterClient(client)
		}
	}
	m.stats.BroadcastsSent++
}

// SendToUser 发送消息给指定用户
func (m *Manager) SendToUser(userID string, msg *entity.WebSocketMessage) error {
	m.clientsMux.RLock()
	client, exists := m.clients[userID]
	m.clientsMux.RUnlock()
	
	if !exists {
		logger.Debug("User not connected", zap.String("user_id", userID))
		return nil // 用户不在线，不返回错误
	}
	
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	
	select {
	case client.Send <- data:
		m.stats.MessagesSent++
		return nil
	default:
		return nil
	}
}

// SendToUsers 批量发送消息
func (m *Manager) SendToUsers(userIDs []string, msg *entity.WebSocketMessage) error {
	m.broadcast <- &BroadcastMessage{
		Message: msg,
		UserIDs: userIDs,
	}
	return nil
}

// Broadcast 广播消息给所有用户
func (m *Manager) Broadcast(msg *entity.WebSocketMessage) error {
	m.broadcast <- &BroadcastMessage{
		Message: msg,
	}
	return nil
}

// BroadcastToOrg 广播给组织内所有用户
func (m *Manager) BroadcastToOrg(orgID string, msg *entity.WebSocketMessage) error {
	m.broadcast <- &BroadcastMessage{
		Message: msg,
		OrgID:   orgID,
	}
	return nil
}

// IsUserOnline 检查用户是否在线
func (m *Manager) IsUserOnline(userID string) bool {
	m.clientsMux.RLock()
	defer m.clientsMux.RUnlock()
	_, exists := m.clients[userID]
	return exists
}

// GetOnlineUsers 获取在线用户列表
func (m *Manager) GetOnlineUsers() []string {
	m.clientsMux.RLock()
	defer m.clientsMux.RUnlock()
	
	users := make([]string, 0, len(m.clients))
	for userID := range m.clients {
		users = append(users, userID)
	}
	return users
}

// GetOnlineCount 获取在线用户数
func (m *Manager) GetOnlineCount() int {
	m.clientsMux.RLock()
	defer m.clientsMux.RUnlock()
	return len(m.clients)
}

// GetStats 获取统计信息
func (m *Manager) GetStats() Stats {
	m.clientsMux.RLock()
	defer m.clientsMux.RUnlock()
	return m.stats
}

// pingAll 发送心跳给所有客户端
func (m *Manager) pingAll() {
	m.clientsMux.RLock()
	clients := make([]*Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	m.clientsMux.RUnlock()
	
	pingMsg := []byte(`{"type":"ping","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`)
	
	for _, client := range clients {
		select {
		case client.Send <- pingMsg:
		default:
			// 发送失败，将在cleanDeadConnections中清理
		}
	}
}

// cleanDeadConnections 清理死连接
func (m *Manager) cleanDeadConnections() {
	m.clientsMux.Lock()
	defer m.clientsMux.Unlock()
	
	now := time.Now()
	for userID, client := range m.clients {
		if now.Sub(client.LastPing) > m.PongTimeout {
			logger.Debug("Cleaning dead connection", zap.String("user_id", userID))
			delete(m.clients, userID)
			client.close()
		}
	}
	m.stats.ActiveConnections = int64(len(m.clients))
}

// shutdown 关闭所有连接
func (m *Manager) shutdown() {
	m.clientsMux.Lock()
	defer m.clientsMux.Unlock()
	
	for _, client := range m.clients {
		client.close()
	}
	m.clients = make(map[string]*Client)
	m.stats.ActiveConnections = 0
}

// readPump 读取泵
func (c *Client) readPump() {
	defer func() {
		c.Manager.unregister <- c
		c.Conn.Close()
	}()
	
	c.Conn.SetReadLimit(c.Manager.MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(c.Manager.PongTimeout))
	c.Conn.SetPongHandler(func(string) error {
		c.LastPing = time.Now()
		c.Conn.SetReadDeadline(time.Now().Add(c.Manager.PongTimeout))
		return nil
	})
	
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket read error", zap.Error(err), zap.String("user_id", c.UserID))
			}
			break
		}
		
		c.Manager.stats.MessagesReceived++
		c.handleMessage(message)
	}
}

// writePump 写入泵
func (c *Client) writePump() {
	ticker := time.NewTicker(c.Manager.PingInterval)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			c.Conn.WriteMessage(websocket.TextMessage, message)
			
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理收到的消息
func (c *Client) handleMessage(message []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		logger.Error("Failed to unmarshal message", zap.Error(err))
		return
	}
	
	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}
	
	switch msgType {
	case "pong":
		c.LastPing = time.Now()
	case "ping":
		// 响应客户端ping
		c.Send <- []byte(`{"type":"pong"}`)
	case "read":
		// 标记消息已读
		if notifID, ok := msg["notification_id"].(string); ok {
			logger.Debug("Mark notification as read", zap.String("notification_id", notifID), zap.String("user_id", c.UserID))
		}
	default:
		logger.Debug("Unknown message type", zap.String("type", msgType), zap.String("user_id", c.UserID))
	}
}

// close 关闭客户端连接
func (c *Client) close() {
	if !c.IsAlive {
		return
	}
	c.IsAlive = false
	close(c.Send)
	c.Conn.Close()
}

// generateClientID 生成客户端ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + randomString(8)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
