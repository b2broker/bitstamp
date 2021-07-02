package bitstamp

import (
	"time"

	"github.com/sirupsen/logrus"
)

func ExampleObserver() {
	logrus.SetLevel(logrus.DebugLevel)
	wsObserver := NewWebsocketObserver()

	wsClient := NewWSClient(wsObserver, "btcusdc", "btcusdt", "btcusd")
	go func() {
		if err := wsClient.Run(time.Second * 10); err != nil {
			logrus.WithError(err).Error("got an error on WebSocket-client")
		}
	}()

	go func() {
		for trade := range wsClient.Fills() {
			logrus.WithField("trade", trade).Info("got trade report")
		}
	}()

	time.Sleep(time.Second * 4)

	bsSvc := NewPrivateClient("_", "_", wsObserver)

	report, err := bsSvc.BuyMarketOrder("btcusdc", "0.0009")
	if err != nil {
		logrus.WithError(err).Error("could not place order")
	} else {
		logrus.Info("order has been placed: ", report)
	}

	time.Sleep(time.Second * 1)

	report2, err := bsSvc.SellMarketOrder("btcusdc", "0.0009")
	if err != nil {
		logrus.WithError(err).Error("could not place order")
	} else {
		logrus.Info("order has been placed: ", report2)
	}

	select {}
}
