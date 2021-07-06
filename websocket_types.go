package bitstamp

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	eventTrade        = "trade"
	eventSubscription = "bts:subscription_succeede"
)

type bitstampFill struct {
	Channel string `json:"channel"`
	Data    struct {
		ID          int64   `json:"id"`
		Amount      float64 `json:"amount"`
		Price       float64 `json:"price"`
		BuyOrderID  int64   `json:"buy_order_id"`
		SellOrderID int64   `json:"sell_order_id"`
		Type        int     `json:"type"`
		Timestamp   int64   `json:"microtimestamp,string"`
	} `json:"data"`
	Event string `json:"event"`
}

func convertMessage(data []byte) (Fill, error) {
	var fill bitstampFill

	if err := json.Unmarshal(data, &fill); err != nil {
		return Fill{}, err
	}

	if fill.Event != eventTrade && fill.Event != eventSubscription {
		return Fill{}, fmt.Errorf("not compatible event-type")
	}

	symbol := strings.Replace(fill.Channel, "live_trades_", "", 1)
	createdAt := time.Unix(fill.Data.Timestamp/1000000, fill.Data.Timestamp%1000*1000000)

	var side string
	if fill.Data.Type == OrderSideBuy {
		side = sideBuy
	} else {
		side = sideSell
	}

	return Fill{
		TradeID:     fill.Data.ID,
		OrderID:     0,
		BuyOrderID:  fill.Data.BuyOrderID,
		SellOrderID: fill.Data.SellOrderID,
		Symbol:      symbol,
		Price:       fill.Data.Price,
		Size:        fill.Data.Amount,
		Side:        side,
		FilledAt:    createdAt,
	}, nil
}

type websocketMessage struct {
	Event string `json:"event"`
	Data  struct {
		Channel string `json:"channel"`
	} `json:"data"`
}
