# Bitstamp API v2

# Получение трейдов из Websocket
```go

func Example() {
	logrus.SetLevel(logrus.DebugLevel)

	bsSvc := bitstamp.NewPrivateClient("_", "_")
	wsClient := bitstamp.NewWSClient("btcusdc")

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

```
