package core_ws

import (
	"time"

	core_logger "github.com/Rics69/rics-chat/internal/core/logger"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = pongWait * 9 / 10
	maxMessageSize = 512
	sendBufferSize = 64
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID int64
	send   chan []byte

	log *core_logger.Logger
}

func newClient(hub *Hub, conn *websocket.Conn, userID int64, log *core_logger.Logger) *Client {
	return &Client{
		hub:    hub,
		conn:   conn,
		userID: userID,
		send:   make(chan []byte, sendBufferSize),
		log:    log,
	}
}

// readPump нужен даже если клиент ничего не шлёт:
// только чтение обрабатывает pong/close фреймы и замечает разрыв соединения
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister(c)
		_ = c.conn.Close()

		c.log.Debug("websocket disconnected")
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				c.log.Debug("websocket read error", zap.Error(err))
			}

			return
		}
	}
}

// writePump — единственная горутина, которая пишет в conn: gorilla не поддерживает конкурентную запись
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case payload, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				c.log.Debug("websocket write error", zap.Error(err))
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
