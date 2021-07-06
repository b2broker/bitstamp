package bitstamp

import (
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	PingFrequency = time.Second * 10
	PongWait      = time.Second * 20
)

var (
	logger = logrus.WithField("service", "bitstamp").WithField("module", "wsconn")
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

// keepalive отправляет ping сообщения, чтобы поддерживать websocket connection
func (ws *WSConn) keepalive(close chan struct{}) {
	// создать тикер, чтобы каждые PingFrequency секунд отправлять ping-сообщения
	tk := time.NewTicker(PingFrequency)
	defer tk.Stop()

	// если в ответ прислали pong, то обновляется deadline connection'a
	ws.conn.SetPongHandler(func(appData string) error {
		logger.Debug("got pong message")
		if err := ws.conn.SetReadDeadline(time.Now().Add(PongWait)); err != nil {
			logger.WithError(err).Error("could not update read-readline")
			return err
		}
		return nil
	})

	for {
		select {
		case <-close:
			return
		case <-tk.C:
			logger.Debug("sending ping")
			err := ws.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				logger.WithError(err).Error("could not send ping-message")
				return
			}
		}
	}
}

// reader читает данные из WebSocket'a
func (ws *WSConn) reader(timeout time.Duration) {
	keepAliveStop := make(chan struct{})

	defer func() {
		close(keepAliveStop)
		close(ws.readerCh)
	}()

	go ws.keepalive(keepAliveStop)

	for {
		if err := ws.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			logger.WithError(err).Error("could not set deadline for websocket")
			return
		}
		_, msg, err := ws.conn.ReadMessage()
		if err != nil {
			logger.WithError(err).Error("could not read from websocket")
			return
		}

		ws.readerCh <- msg
	}
}

// RunReader запускает процесс чтения данных из WebSocket'a
func (ws *WSConn) RunReader(wsTimeout time.Duration) <-chan []byte {
	ws.readerCh = make(chan []byte, 256)
	go ws.reader(wsTimeout)
	return ws.readerCh
}

// SendMessage отправляет сообщение по WebSocket протоколу
func (ws *WSConn) SendMessage(msg string) error {
	return ws.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}
