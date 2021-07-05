package bitstamp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

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

		// TODO: replace with toFloat func
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

// CancelAllOrdersResult asd
// TODO: fill Canceled field
type CancelAllOrdersResult struct {
	Success  bool          `json:"success"`
	Canceled []interface{} `json:"canceled"`
}

type OrderStatus struct {
	Fee        float64            `json:"fee"`
	Price      float64            `json:"price"`
	Datetime   time.Time          `json:"datetime"`
	Tid        int64              `json:"tid"`
	Type       int                `json:"type"`
	Currencies map[string]float64 `json:"currencies"`
}

type OrderStatusResult struct {
	Status          string        `json:"status"`
	ID              int64         `json:"id"`
	AmountRemaining float64       `json:"amount_remaining,string"`
	Transactions    []OrderStatus `json:"transactions"`
}

type orderStatusResult struct {
	Status          string                   `json:"status"`
	ID              int64                    `json:"id"`
	AmountRemaining float64                  `json:"amount_remaining,string"`
	Transactions    []map[string]interface{} `json:"transactions"`
}

func interfaceToFloat(data interface{}) (float64, error) {
	var parsedValue float64

	switch vv := data.(type) {
	case string:
		xx, err := strconv.ParseFloat(vv, 64)
		if err != nil {
			return 0, err
		}
		parsedValue = xx

	case float64:
		parsedValue = vv
	case int:
		parsedValue = float64(vv)
	default:
		return 0, fmt.Errorf("wrong type")
	}

	return parsedValue, nil
}

// UnmarshalJSON unmarshaller
// {"status": "Finished", "id": 1373320601649153, "amount_remaining": "0.00000000", "transactions": [{"fee": "0.16277", "price": "36171.43000000", "datetime": "2021-06-19 15:58:44.669000", "usd": "32.55428700", "btc": "0.00090000", "tid": 183814449, "type": 2}]}
func (os *OrderStatusResult) UnmarshalJSON(data []byte) error {
	var osr orderStatusResult

	if err := json.Unmarshal(data, &osr); err != nil {
		return err
	}

	(*os).ID = osr.ID
	(*os).Status = osr.Status
	(*os).AmountRemaining = osr.AmountRemaining

	transactions := make([]OrderStatus, 0)

	for _, transaction := range osr.Transactions {

		currencies := make(map[string]float64)
		var orderStatus OrderStatus

		for k, v := range transaction {

			// known fields
			switch k {
			case "fee":
				fee, err := interfaceToFloat(v)
				if err != nil {
					continue
				}
				orderStatus.Fee = fee

			case "price":
				price, err := interfaceToFloat(v)
				if err != nil {
					continue
				}
				orderStatus.Price = price

			case "datetime":
				parsedTime, err := time.Parse("2006-01-02 15:04:05.99", v.(string))
				if err != nil {
					continue
				}

				orderStatus.Datetime = parsedTime

			case "tid":
				tid, err := interfaceToFloat(v)
				if err != nil {
					continue
				}
				orderStatus.Tid = int64(tid)

			case "type":
				tp, err := interfaceToFloat(v)
				if err != nil {
					continue
				}
				orderStatus.Type = int(tp)

			// the rest fields are supposed to be currency assets affected by trade
			default:
				parsedValue, err := interfaceToFloat(v)
				if err != nil {
					continue
				}

				currencies[k] = parsedValue
			}

		}
		orderStatus.Currencies = currencies
		transactions = append(transactions, orderStatus)
	}

	(*os).Transactions = transactions

	return nil
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
	sideBuy  = "buy"
	sideSell = "sell"

	OrderSideBuy  = 0
	OrderSideSell = 1

	ExecDefault = ""
	ExecDaily   = "daily"
	ExecFOK     = "fok"
	ExecIOC     = "ioc"

	OrderStatusFinished = "Finished"
	OrderStatusOpen     = "Open"
	OrderStatusCanceled = "Canceled"

	TransactionDeposit    = 0
	TransactionWithdrawal = 1
	TransactionTrade      = 2
)

type PlaceOrder struct {
	Price    string
	Amount   string
	ExecType string
	Symbol   string
}

type CommonResult interface {
	GetID() int64
	GetDateTime() string
	GetType() int
	GetPrice() float64
	GetAmount() float64
}

func (r BuyLimitOrderResult) GetID() int64 {
	return r.ID
}

func (r BuyLimitOrderResult) GetDateTime() string {
	return r.DateTime
}

func (r BuyLimitOrderResult) GetType() int {
	return r.Type
}

func (r BuyLimitOrderResult) GetPrice() float64 {
	return r.Price
}

func (r BuyLimitOrderResult) GetAmount() float64 {
	return r.Amount
}

func (r SellLimitOrderResult) GetID() int64 {
	return r.ID
}

func (r SellLimitOrderResult) GetDateTime() string {
	return r.DateTime
}

func (r SellLimitOrderResult) GetType() int {
	return r.Type
}

func (r SellLimitOrderResult) GetPrice() float64 {
	return r.Price
}

func (r SellLimitOrderResult) GetAmount() float64 {
	return r.Amount
}

func (r BuyMarketOrderResult) GetID() int64 {
	return r.ID
}

func (r BuyMarketOrderResult) GetDateTime() string {
	return r.DateTime
}

func (r BuyMarketOrderResult) GetType() int {
	return r.Type
}

func (r BuyMarketOrderResult) GetPrice() float64 {
	return r.Price
}

func (r BuyMarketOrderResult) GetAmount() float64 {
	return r.Amount
}

func (r SellMarketOrderResult) GetID() int64 {
	return r.ID
}

func (r SellMarketOrderResult) GetDateTime() string {
	return r.DateTime
}

func (r SellMarketOrderResult) GetType() int {
	return r.Type
}

func (r SellMarketOrderResult) GetPrice() float64 {
	return r.Price
}

func (r SellMarketOrderResult) GetAmount() float64 {
	return r.Amount
}
