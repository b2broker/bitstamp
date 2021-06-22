package bitstamp

import (
	"encoding/json"
	"strconv"
	"strings"
)

// {currency}_balance to get total balance (including reserved)
// {currency}_available to get available amount
// {currency}_reserved to get reserved amount
// {instrument}_fee to get fee on instrument transactions
type BalanceResult map[string]float64

type transactionBody struct {
	ID       int64   `json:"id"`
	OrderID  int64   `json:"order_id"`
	DateTime string  `json:"datetime"`
	Type     int     `json:"type,string"`
	Fee      float64 `json:"fee,string"`
}

type TransactionResult struct {
	transactionBody
	Amounts map[string]float64
}

type OpenOrderResult struct {
	ID           int64   `json:"id,string"`
	DateTime     string  `json:"datetime"`
	Type         int     `json:"type,string"`
	Price        float64 `json:"price,string"`
	Amount       float64 `json:"amount,string"`
	CurrencyPair string  `json:"currency_pair"`
}

func (br *BalanceResult) UnmarshalJSON(data []byte) error {
	t := make(map[string]interface{})
	*br = make(map[string]float64)

	err := json.Unmarshal(data, &t)
	if err != nil {
		return err
	}

	for key, value := range t {
		key = strings.ToLower(key)

		if !strings.HasSuffix(key, "_balance") {
			continue
		}

		key = strings.Replace(key, "_balance", "", 1)

		var parsedValue float64

		switch pp := value.(type) {
		case string:
			tmpFloat, err := strconv.ParseFloat(pp, 64)
			if err != nil {
				return err
			}

			parsedValue = tmpFloat

		case float64:
			parsedValue = pp
		}

		(*br)[key] = parsedValue
	}

	return nil
}

func (u *TransactionResult) UnmarshalJSON(data []byte) error {
	var results transactionBody
	var tmp map[string]interface{}

	knownFields := map[string]struct{}{
		"id":       {},
		"order_id": {},
		"datetime": {},
		"type":     {},
		"fee":      {},
	}

	amounts := make(map[string]float64)

	err := json.Unmarshal(data, &results)
	if err != nil {
		return err
	}

	(*u).transactionBody = results

	err = json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	for key, value := range tmp {
		if _, ok := knownFields[key]; ok {
			continue
		}

		var parsedValue float64

		switch vv := value.(type) {
		case string:
			xx, err := strconv.ParseFloat(vv, 64)
			if err != nil {
				return err
			}
			parsedValue = xx

		case float64:
			parsedValue = vv
		case int:
			parsedValue = float64(vv)
		}

		amounts[key] = parsedValue
	}

	(*u).Amounts = amounts

	return nil
}

type CancelAllOrdersResult struct {
	Success  bool          `json:"success"`
	Canceled []interface{} `json:"canceled"`
}

// TODO: check transactions type
type OrderStatusResult struct {
	Status          string                   `json:"status"`
	ID              int64                    `json:"id"`
	AmountRemaining float64                  `json:"amount_remaining,string"`
	Transactions    []map[string]interface{} `json:"transactions"`
}

type OrderCancelResult struct {
	ID     string  `json:"id"`
	Amount float64 `json:"amount,string"`
	Price  float64 `json:"price,string"`
	Type   int     `json:"type,string"`
}

type PlaceOrderResult struct {
	ID       int64   `json:"id,string"`
	DateTime string  `json:"datetime"`
	Type     int     `json:"type,string"`
	Price    float64 `json:"price,string"`
	Amount   float64 `json:"amount,string"`
}

type BuyLimitOrderResult struct {
	PlaceOrderResult
}

type SellLimitOrderResult struct {
	PlaceOrderResult
}

type BuyMarketOrderResult struct {
	PlaceOrderResult
}

type SellMarketOrderResult struct {
	PlaceOrderResult
}

const (
	SideBuy  = "buy"
	SideSell = "sell"

	ExecDefault = ""
	ExecDaily   = "daily"
	ExecFOK     = "fok"
	ExecIOC     = "ioc"
)

type PlaceOrder struct {
	Price    string
	Amount   string
	ExecType string
	Symbol   string
}
