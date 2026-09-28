package core_ws

import (
	"sync"

	core_logger "github.com/Rics69/rics-chat/internal/core/logger"
	"go.uber.org/zap"
)

// Hub хранит все живые соединения: у одного пользователя может быть несколько вкладок
type Hub struct {
	mu      sync.RWMutex
	clients map[int64]map[*Client]struct{}
	closed  bool

	log *core_logger.Logger
}

func NewHub(log *core_logger.Logger) *Hub {
	return &Hub{
		clients: make(map[int64]map[*Client]struct{}),
		log:     log,
	}
}

func (h *Hub) register(client *Client) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.closed {
		return false
	}

	userClients, ok := h.clients[client.userID]
	if !ok {
		userClients = make(map[*Client]struct{})
		h.clients[client.userID] = userClients
	}

	userClients[client] = struct{}{}

	return true
}

// unregister можно звать сколько угодно раз: send закрывается только если клиент ещё в map
func (h *Hub) unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.removeLocked(client)
}

func (h *Hub) removeLocked(client *Client) {
	userClients, ok := h.clients[client.userID]
	if !ok {
		return
	}

	if _, ok := userClients[client]; !ok {
		return
	}

	delete(userClients, client)
	close(client.send)

	if len(userClients) == 0 {
		delete(h.clients, client.userID)
	}
}

func (h *Hub) SendToUser(userID int64, payload []byte) {
	var slowClients []*Client

	h.mu.RLock()
	for client := range h.clients[userID] {
		select {
		case client.send <- payload:
		default:
			// буфер полон — клиент не успевает читать; не блокируем отправителя, а отключаем его
			slowClients = append(slowClients, client)
		}
	}
	h.mu.RUnlock()

	for _, client := range slowClients {
		h.log.Warn("disconnect slow websocket client", zap.Int64("user_id", client.userID))
		h.unregister(client)
	}
}

func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.closed = true

	for _, userClients := range h.clients {
		for client := range userClients {
			h.removeLocked(client)
		}
	}
}
