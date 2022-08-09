package main

import (
	"time"

	"github.com/b2broker/bitstamp"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	bsSvc := bitstamp.NewPrivateClient("_", "_")
	wsClient := bitstamp.NewWSClient("btcusdt")

	go func() {
		if err := wsClient.Run(bsSvc, time.Second*10); err != nil {
			logrus.WithError(err).Error("got an error on WebSocket-client")
		}
	}()

	go func() {
		for trade := range wsClient.Fills() {
			logrus.WithField("trade", trade).Info("got trade report")
		}
	}()

	time.Sleep(time.Second * 4)

	balances, err := bsSvc.GetBalances()
	if err != nil {
		logrus.WithError(err).Error("could not get balances")
	} else {
		logrus.Info("Balances: ", balances)
	}

	report, err := bsSvc.BuyMarketOrder("btcusdt", "0.0009")
	if err != nil {
		logrus.WithError(err).Error("could not place order")
	} else {
		logrus.Info("order has been placed: ", report)
	}

	time.Sleep(time.Second * 1)

	report2, err := bsSvc.SellMarketOrder("btcusdt", "0.0009")
	if err != nil {
		logrus.WithError(err).Error("could not place order")
	} else {
		logrus.Info("order has been placed: ", report2)
	}

	report3, err := bsSvc.PlaceOrder(bitstamp.PlaceOrderRequest{
		Amount: 0.0009,
		Symbol: "btcusdt",
		Side:   bitstamp.Buy,
		Type:   bitstamp.Market,
	})
	if err != nil {
		logrus.WithError(err).Error("could not place order")
	} else {
		logrus.Info("order has been placed: ", report3)
	}

	select {}
}
