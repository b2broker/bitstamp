package bitstamp

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// WSConn драйвер для получения OrderBook с бирж, работащих с WebSocket
type WSConn struct {
	conn     *websocket.Conn
	readerCh chan []byte
}

// NewWSConn создает новый экземпляр *WebSocket
func NewWSConn(conn *websocket.Conn) *WSConn {
	return &WSConn{
		conn: conn,
	}
}

// reader читает данные из WebSocket'a
func (ws *WSConn) reader(timeout time.Duration) {
	defer close(ws.readerCh)

	for {
		if err := ws.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			logrus.WithError(err).Error("could not set deadline for websocket")
			return
		}
		_, msg, err := ws.conn.ReadMessage()
		if err != nil {
			logrus.WithError(err).Error("could not read from websocket")
			return
		}

		ws.readerCh <- msg
	}
}

// RunReader запускает процесс чтения данных из WebSocket'a
func (ws *WSConn) RunReader() <-chan []byte {
	ws.readerCh = make(chan []byte, 256)
	go ws.reader(time.Second * 15)
	return ws.readerCh
}

// SendMessage отправляет сообщение по WebSocket протоколу
func (ws *WSConn) SendMessage(msg string) error {
	return ws.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}
