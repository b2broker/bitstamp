package bitstamp

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

// Websocket коннектор для Bitstamp для получение трейдов
type Websocket struct {
	symbols []string
	fills   chan Fill
	logger  *logrus.Entry
}

const (
	bitstampWS = "wss://ws.bitstamp.net/"
)

// NewWSClient Создает новый Websocket инстанс
func NewWSClient(symbols ...string) *Websocket {
	return &Websocket{
		symbols: symbols,
		fills:   make(chan Fill, 256),
		logger:  logrus.WithField("provider", "bitstamp").WithField("module", "websocket"),
	}
}

func (ws *Websocket) subscribe(conn *WSConn, tokenData *GenerateWSTokenResult) error {
	if len(ws.symbols) == 0 {
		return fmt.Errorf("no symbols to subscribe")
	}

	for _, symbol := range ws.symbols {
		msg := websocketMessage{
			Event: "bts:subscribe",
			Data: struct {
				Channel string `json:"channel"`
				Auth    string `json:"auth"`
			}{
				Channel: fmt.Sprintf("private-my_trades_%s-%v",
					symbol,
					tokenData.UserID,
				),
				Auth: tokenData.Token,
			},
		}

		result, err := json.Marshal(msg)
		if err != nil {
			return err
		}

		if err := conn.SendMessage(string(result)); err != nil {
			return err
		}
	}

	return nil
}

// Run синхронная функция, которая подключается к Websocket'у, пересоздает connection в случае дисконекта
func (ws *Websocket) Run(httpPrivateClient *PrivateClient, reconnectDelay time.Duration) error {
	for {
		ws.logger.Info("connecting")

		// если connection не удался, то через reconnectDelay будет повторная попытка подключения
		conn, err := ws.connect()
		if err != nil {
			time.Sleep(reconnectDelay)
			continue
		}

		tokenData, err := httpPrivateClient.GenerateWSToken()
		if err != nil {
			ws.logger.Errorf("could not generate token: %v", err)
			time.Sleep(reconnectDelay)
			continue
		}

		if err := ws.subscribe(conn, tokenData); err != nil {
			return fmt.Errorf("could not subscribe: %v", err)
		}

		incoming := conn.RunReader(time.Second * 15)

		for msg := range incoming {
			ws.logger.WithField("body", string(msg)).Debug("got msg")

			parsedMsg, err := convertMessage(msg)
			if err != nil {
				ws.logger.WithError(err).Error("could not convert message")
			}

			ws.fills <- parsedMsg
		}
	}
}

func (ws *Websocket) connect() (*WSConn, error) {
	dialer := websocket.DefaultDialer

	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint

	conn, _, err := dialer.Dial(bitstampWS, nil)
	if err != nil {
		return nil, err
	}

	return NewWSConn(conn), nil
}

// Fill трейд, который получает клиент из библиотеки
type Fill struct {
	OrderID  int64
	TradeID  int64
	Symbol   string
	Price    float64
	Size     float64
	Fee      float64
	Side     string
	FilledAt time.Time
}

// Fills возвращает канал
func (ws *Websocket) Fills() <-chan Fill {
	return ws.fills
}
