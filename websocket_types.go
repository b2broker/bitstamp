package bitstamp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	eventTrade        = "trade"
	eventSubscription = "bts:subscription_succeeded"
)

type bitstampFill struct {
	Channel string `json:"channel"`
	Data    struct {
		ID            int64  `json:"id"`
		OrderID       int64  `json:"buy_order_id"`
		ClientOrderID string `json:"client_order_id"`
		Amount        string `json:"amount"`
		Price         string `json:"price"`
		Fee           string `json:"fee"`
		Side          string `json:"Side"`
		Timestamp     int64  `json:"microtimestamp,string"`
	} `json:"data"`
	Event string `json:"event"`
}

func convertMessage(data []byte) (Fill, error) {
	var fill bitstampFill

	if err := json.Unmarshal(data, &fill); err != nil {
		return Fill{}, err
	}

	if fill.Event != eventTrade && fill.Event != eventSubscription {
		return Fill{}, fmt.Errorf("incompatible event-type:%+v", fill)
	}

	symbol := strings.Replace(fill.Channel, "private-my_trades_", "", 1)
	createdAt := time.Unix(fill.Data.Timestamp/1000000, fill.Data.Timestamp%1000*1000000)

	if fill.Data.Side != string(Buy) && fill.Data.Side != string(Sell) {
		return Fill{}, fmt.Errorf("not valid side: %s", fill.Data.Side)
	}

	amount, err := strconv.ParseFloat(fill.Data.Amount, 64)
	if err != nil {
		return Fill{}, fmt.Errorf("amount convertation error: %w", err)
	}

	price, err := strconv.ParseFloat(fill.Data.Price, 64)
	if err != nil {
		return Fill{}, fmt.Errorf("price convertation error: %w", err)
	}

	fee, err := strconv.ParseFloat(fill.Data.Fee, 64)
	if err != nil {
		return Fill{}, fmt.Errorf("fee convertation error: %w", err)
	}

	return Fill{
		TradeID:       fill.Data.ID,
		OrderID:       fill.Data.OrderID,
		ClientOrderID: fill.Data.ClientOrderID,
		Symbol:        symbol,
		Price:         price,
		Size:          amount,
		Fee:           fee,
		Side:          fill.Data.Side,
		FilledAt:      createdAt,
	}, nil
}

type websocketMessage struct {
	Event string `json:"event"`
	Data  struct {
		Channel string `json:"channel"`
		Auth    string `json:"auth"`
	} `json:"data"`
}
