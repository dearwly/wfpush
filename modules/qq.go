package modules

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// --- 新增：定义接收QQ消息事件的结构体 ---
type QQMessageEvent struct {
	PostType    string `json:"post_type"`
	MessageType string `json:"message_type"`
	UserID      int64  `json:"user_id"`
	GroupID     int64  `json:"group_id,omitempty"`
	RawMessage  string `json:"raw_message"`
}

// --- 新增：定义发送消息的结构体 ---
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

// QQBotServer 结构体管理所有 WebSocket 连接
type QQBotServer struct {
	clients         map[*websocket.Conn]bool
	broadcast       chan []byte
	upgrader        websocket.Upgrader
	mu              sync.Mutex
	pendingMessages [][]byte
	cfg             *Config // 新增: 持有配置的引用
}

var server *QQBotServer

func NewQQBotServer(cfg *Config) *QQBotServer {
	return &QQBotServer{
		clients:         make(map[*websocket.Conn]bool),
		broadcast:       make(chan []byte),
		pendingMessages: make([][]byte, 0),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		cfg: cfg, // 保存配置
	}
}

// isMaster 检查一个用户ID是否在主人列表中
func isMaster(userID int64, masters []string) bool {
	userIDStr := strconv.FormatInt(userID, 10)
	for _, master := range masters {
		if master == userIDStr {
			return true
		}
	}
	return false
}

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

	// 发送暂存消息
	if len(s.pendingMessages) > 0 {
		Log(fmt.Sprintf("检测到 %d 条待处理消息，正在发送...", len(s.pendingMessages)), INFO)
		for _, msg := range s.pendingMessages {
			if err := ws.WriteMessage(websocket.TextMessage, msg); err != nil {
				Log(fmt.Sprintf("发送待处理消息失败: %v", err), ERROR)
				delete(s.clients, ws)
				s.mu.Unlock()
				return
			}
		}
		s.pendingMessages = make([][]byte, 0)
		Log("待处理消息发送完毕。", INFO)
	}
	s.mu.Unlock()

	// --- 核心修改：监听并处理来自客户端的指令 ---
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			s.mu.Lock()
			delete(s.clients, ws)
			s.mu.Unlock()
			Log("QQ机器人客户端已断开。", INFO)
			break
		}

		// 解析收到的消息
		var event QQMessageEvent
		if err := json.Unmarshal(message, &event); err != nil {
			continue // 不是我们关心的消息格式，忽略
		}

		// 只处理消息类型的事件
		if event.PostType != "message" {
			continue
		}

		// 检查权限
		if !isMaster(event.UserID, s.cfg.QQ.Masters) {
			continue // 不是主人，忽略
		}

		// 解析指令
		parts := strings.Fields(event.RawMessage) // 使用 Fields 可以更好地处理多个空格
		if len(parts) == 0 {
			continue
		}
		command := strings.ToLower(parts[0])
		args := parts[1:]
		var responseText string

		// 执行指令
		switch command {
		case "list":
			responseText, _ = ListSubscriptions()
		case "add":
			responseText, _ = AddSubscription(args)
		case "delete":
			if len(args) < 1 {
				responseText = "用法: delete <索引>"
			} else {
				responseText, _ = DeleteSubscription(args[0])
			}
		default:
			continue // 不是已知指令，忽略
		}

		// 构建回复消息
		var replyAction QQAction
		replyParams := QQParams{
			Message: []MessageSegment{{Type: "text", Data: map[string]string{"text": responseText}}},
		}
		if event.MessageType == "group" {
			replyAction.Action = "send_group_msg"
			replyParams.GroupID = event.GroupID
		} else {
			replyAction.Action = "send_private_msg"
			replyParams.UserID = event.UserID
		}
		replyAction.Params = replyParams

		replyPayload, _ := json.Marshal(replyAction)

		// 将回复消息发送给所有客户端（通常只有一个机器人客户端）
		s.broadcast <- replyPayload
	}
}

func (s *QQBotServer) handleMessages() {
	for {
		msg := <-s.broadcast
		s.mu.Lock()
		for client := range s.clients {
			if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Printf("WebSocket write error: %v", err)
				client.Close()
				delete(s.clients, client)
			}
		}
		s.mu.Unlock()
	}
}

// StartQQBotServer 修改：传入配置
func StartQQBotServer(cfg *Config) {
	server = NewQQBotServer(cfg)
	addr := fmt.Sprintf("%s:%d", cfg.QQ.WebSocketServer.Host, cfg.QQ.WebSocketServer.Port)

	http.HandleFunc("/ws", server.handleConnections)
	go server.handleMessages()

	Log(fmt.Sprintf("QQ WebSocket 服务器正在启动，监听地址: %s", addr), INFO)
	if err := http.ListenAndServe(addr, nil); err != nil {
		Log(fmt.Sprintf("WebSocket 服务器启动失败: %v", err), ERROR)
	}
}

// SendQQNotification 函数几乎不变，只是去掉了内部的消息结构体定义
func SendQQNotification(fissures []Fissure, cfg *Config) {
	if server == nil {
		Log("QQ Bot 服务器未初始化，无法发送消息。", WARNING)
		return
	}

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

	for _, target := range cfg.QQ.PushTargets {
		var action QQAction
		params := QQParams{
			Message: []MessageSegment{{Type: "text", Data: map[string]string{"text": messageText}}},
		}

		if target.Type == "group" {
			action.Action = "send_group_msg"
			params.GroupID, _ = strconv.ParseInt(target.ID, 10, 64)
		} else if target.Type == "private" {
			action.Action = "send_private_msg"
			params.UserID, _ = strconv.ParseInt(target.ID, 10, 64)
		} else {
			continue
		}

		action.Params = params
		payload, err := json.Marshal(action)
		if err != nil {
			Log(fmt.Sprintf("序列化QQ消息JSON失败: %v", err), ERROR)
			continue
		}

		server.mu.Lock()
		if len(server.clients) == 0 {
			server.pendingMessages = append(server.pendingMessages, payload)
		} else {
			server.broadcast <- payload
		}
		server.mu.Unlock()
	}
	Log(fmt.Sprintf("已处理 %d 个QQ目标的裂缝通知。", len(cfg.QQ.PushTargets)), INFO)
}
