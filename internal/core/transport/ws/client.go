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

	// 4000-4999 — коды закрытия для приложений, клиент по нему понимает, что надо перелогиниться
	closeCodeTokenExpired = 4001
)

type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	userID int64
	send   chan []byte

	expiresAt time.Time

	log *core_logger.Logger
}

func newClient(
	hub *Hub,
	conn *websocket.Conn,
	userID int64,
	expiresAt time.Time,
	log *core_logger.Logger,
) *Client {
	return &Client{
		hub:       hub,
		conn:      conn,
		userID:    userID,
		send:      make(chan []byte, sendBufferSize),
		expiresAt: expiresAt,
		log:       log,
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
	tokenExpired := time.NewTimer(time.Until(c.expiresAt))
	defer func() {
		ticker.Stop()
		tokenExpired.Stop()
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
		case <-tokenExpired.C:
			c.log.Debug("websocket token expired")

			_ = c.conn.WriteControl(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(closeCodeTokenExpired, "token expired"),
				time.Now().Add(writeWait),
			)

			return
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
