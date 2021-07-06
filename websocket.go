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
	observer OrderObserver
	symbols  []string
	fills    chan Fill
	logger   *logrus.Entry
}

const (
	bitstampWS = "wss://ws.bitstamp.net/"
)

// NewWSClient Создает новый Websocket инстанс
func NewWSClient(observer OrderObserver, symbols ...string) *Websocket {
	return &Websocket{
		observer: observer,
		symbols:  symbols,
		fills:    make(chan Fill, 256),
		logger:   logrus.WithField("provider", "bitstamp").WithField("module", "websocket"),
	}
}

func (ws *Websocket) subscribe(conn *WSConn) error {
	if len(ws.symbols) == 0 {
		return fmt.Errorf("no symbols to subscribe")
	}

	for _, symbol := range ws.symbols {
		msg := websocketMessage{
			Event: "bts:subscribe",
			Data: struct {
				Channel string `json:"channel"`
			}{
				Channel: fmt.Sprintf("live_trades_%s", symbol),
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
// Синхронизируется с OrderObserver
func (ws *Websocket) Run(reconnectDelay time.Duration) error {
	for {
		// Прежде чем подключиться к websocket'у observer блокируется, чтобы не потерять исполнения,
		// которые могут прийти до успешного подключения к ws
		ws.logger.Info("connecting")
		if err := ws.observer.Lock(); err != nil {
			time.Sleep(reconnectDelay)
			continue
		}

		// если connection не удался, то через reconnectDelay будет повторная попытка подключения
		conn, err := ws.connect()
		if err != nil {
			time.Sleep(reconnectDelay)
			// TODO: если тут разблокировать, то это теоритически race-condition (ордер можно успеть выставить до
			//  вызова Lock в начале процедуры). Поэтому если блокировка была установлена, то ее не нужно снимать тут
			//  и не нужно повторно устанавливать в начале процедуры
			ws.observer.Unlock()
			continue
		}

		if err := ws.subscribe(conn); err != nil {
			return fmt.Errorf("could not subscribe: %v", err)
		}

		ws.observer.Unlock()

		incoming := conn.RunReader(time.Second * 15)
		for msg := range incoming {
			ws.logger.WithField("body", string(msg)).Debug("got msg")
			parsedMsg, err := convertMessage(msg)
			if err != nil {
				ws.logger.WithError(err).Error("could not convert message")
			}

			// если ID ордера найдено в OrderObserver, то исполнение отправляется наружу
			if ws.observer.IsObservable(parsedMsg.BuyOrderID) {
				parsedMsg.OrderID = parsedMsg.BuyOrderID
				ws.fills <- parsedMsg
			}

			if ws.observer.IsObservable(parsedMsg.SellOrderID) {
				parsedMsg.OrderID = parsedMsg.SellOrderID
				ws.fills <- parsedMsg
			}
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
	OrderID     int64
	TradeID     int64
	BuyOrderID  int64
	SellOrderID int64
	Symbol      string
	Price       float64
	Size        float64
	Side        string
	FilledAt    time.Time
}

// Fills возвращает канал
func (ws *Websocket) Fills() <-chan Fill {
	return ws.fills
}
