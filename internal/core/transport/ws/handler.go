package core_ws

import (
	"net/http"

	core_logger "github.com/Rics69/rics-chat/internal/core/logger"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const tokenQueryParam = "token"

type TokenParser interface {
	ParseToken(token string) (int64, error)
}

// CheckOrigin не переопределяем: по умолчанию gorilla пускает только тот же Origin,
// что и Host — чужой сайт не сможет открыть сокет от имени нашего пользователя
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Handler: браузерный WebSocket не умеет слать заголовки, поэтому токен идёт в ?token=
func Handler(hub *Hub, tokenParser TokenParser, log *core_logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := tokenParser.ParseToken(r.URL.Query().Get(tokenQueryParam))
		if err != nil {
			log.Debug("websocket unauthenticated", zap.Error(err))
			http.Error(w, "unauthenticated", http.StatusUnauthorized)

			return
		}

		// при ошибке Upgrade сам отвечает клиенту 4xx
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Debug("websocket upgrade", zap.Error(err))
			return
		}

		clientLog := log.With(zap.Int64("user_id", userID))
		client := newClient(hub, conn, userID, clientLog)

		if !hub.register(client) {
			_ = conn.Close()
			return
		}

		clientLog.Debug("websocket connected")

		go client.writePump()
		go client.readPump()
	}
}
