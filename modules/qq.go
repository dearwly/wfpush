package modules

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

// QQBotServer 结构体管理所有 WebSocket 连接
type QQBotServer struct {
	clients         map[*websocket.Conn]bool
	broadcast       chan []byte
	upgrader        websocket.Upgrader
	mu              sync.Mutex // 使用一个标准互斥锁来保护 clients 和 pendingMessages
	pendingMessages [][]byte   // 新增：用于暂存待发送消息的队列
}

// 全局服务器实例
var server *QQBotServer

// NewQQBotServer 创建一个新的服务器实例
func NewQQBotServer() *QQBotServer {
	return &QQBotServer{
		clients:         make(map[*websocket.Conn]bool),
		broadcast:       make(chan []byte),
		pendingMessages: make([][]byte, 0), // 初始化待处理消息队列
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// handleConnections 处理新的 WebSocket 连接
func (s *QQBotServer) handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer ws.Close()

	s.mu.Lock()
	s.clients[ws] = true
	Log("QQ机器人客户端已连接。", INFO)

	// ---- 核心修改：检查并发送暂存的消息 ----
	if len(s.pendingMessages) > 0 {
		Log(fmt.Sprintf("检测到 %d 条待处理消息，正在发送给新连接的客户端...", len(s.pendingMessages)), INFO)
		// 遍历所有待处理消息并发送给这个新连接的客户端
		for _, msg := range s.pendingMessages {
			if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
				Log(fmt.Sprintf("发送待处理消息失败: %v", err), ERROR)
				// 如果发送失败，很可能连接有问题，直接断开
				delete(s.clients, ws)
				s.mu.Unlock()
				return
			}
		}
		// 发送成功后，清空队列
		s.pendingMessages = make([][]byte, 0)
		Log("待处理消息发送完毕。", INFO)
	}
	s.mu.Unlock()

	// 保持连接，监听客户端消息（主要用于检测断开）
	for {
		if _, _, err := ws.ReadMessage(); err != nil {
			s.mu.Lock()
			delete(s.clients, ws)
			s.mu.Unlock()
			Log("QQ机器人客户端已断开。", INFO)
			break
		}
	}
}

// handleMessages 监听 broadcast 通道并向所有已连接的客户端广播消息
func (s *QQBotServer) handleMessages() {
	for {
		msg := <-s.broadcast
		s.mu.Lock()
		// 广播给所有当前在线的客户端
		for client := range s.clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Printf("WebSocket write error: %v", err)
				client.Close()
				delete(s.clients, client) // 在写入失败时移除客户端
			}
		}
		s.mu.Unlock()
	}
}

// StartQQBotServer 启动 WebSocket 服务器
func StartQQBotServer(cfg *Config) {
	server = NewQQBotServer()
	addr := fmt.Sprintf("%s:%d", cfg.QQ.WebSocketServer.Host, cfg.QQ.WebSocketServer.Port)

	http.HandleFunc("/ws", server.handleConnections)
	go server.handleMessages()

	Log(fmt.Sprintf("QQ WebSocket 服务器正在启动，监听地址: %s", addr), INFO)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		Log(fmt.Sprintf("WebSocket 服务器启动失败: %v", err), ERROR)
	}
}

// SendQQNotification 是暴露给外部调用的函数，用于发送裂缝通知
func SendQQNotification(fissures []Fissure, cfg *Config) {
	if server == nil {
		Log("QQ Bot 服务器未初始化，无法发送消息。", WARNING)
		return
	}

	// ... 消息格式化部分保持不变 ...
	var messageText string
	if len(fissures) > 0 {
		messageText = "您订阅的新裂缝已出现：\n"
		for i, fissure := range fissures {
			formatted, _ := formatPrint(fissure)
			messageText += fmt.Sprintf("\n%d. %s", i+1, formatted)
		}
	} else {
		return
	}

	type MessageSegment struct {
		Type string            `json:"type"`
		Data map[string]string `json:"data"`
	}
	type QQParams struct {
		UserID  interface{}      `json:"user_id,omitempty"`
		GroupID interface{}      `json:"group_id,omitempty"`
		Message []MessageSegment `json:"message"`
	}
	type QQAction struct {
		Action string   `json:"action"`
		Params QQParams `json:"params"`
	}

	// 遍历所有推送目标
	for _, target := range cfg.QQ.PushTargets {
		var action QQAction
		params := QQParams{
			Message: []MessageSegment{{Type: "text", Data: map[string]string{"text": messageText}}},
		}

		if target.Type == "group" {
			action.Action = "send_group_msg"
			groupID, _ := strconv.ParseInt(target.ID, 10, 64)
			params.GroupID = groupID
		} else if target.Type == "private" {
			action.Action = "send_private_msg"
			userID, _ := strconv.ParseInt(target.ID, 10, 64)
			params.UserID = userID
		} else {
			continue
		}

		action.Params = params
		payload, err := json.Marshal(action)
		if err != nil {
			Log(fmt.Sprintf("序列化QQ消息JSON失败: %v", err), ERROR)
			continue
		}

		// ---- 核心修改：判断是广播还是暂存 ----
		server.mu.Lock()
		if len(server.clients) == 0 {
			// 如果没有客户端在线，将消息存入队列
			server.pendingMessages = append(server.pendingMessages, payload)
			Log("无QQ客户端连接，消息已加入待处理队列。", INFO)
		} else {
			// 如果有客户端在线，直接发送到广播通道
			server.broadcast <- payload
		}
		server.mu.Unlock()
	}
	// 统一在循环外打印日志
	Log(fmt.Sprintf("已处理 %d 个QQ目标的裂缝通知。", len(cfg.QQ.PushTargets)), INFO)
}
