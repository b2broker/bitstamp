package bitstamp

import (
	"fmt"
	"sync"
)

// OrderObserver используется для синхронизации между Websocket и REST при соврешении трейдов
type OrderObserver interface {
	// Observe добавляет ID трейда в список наблюдаемых, т.к. bitstamp возвращает все трейды, которые происходят на платформе
	Observe(side string, symbol string, orderID int64) error
	// Delete удаляет ордер из списка наблюдаемых. Этот метод обязан использовать пользователь, к примеру,
	// когда получил `Fills` перекрывающий весь объем изначального ордера
	// Иначе буффер будет копится, а то memory-leak
	// У объекта `Fill` есть поле OrderID, этот ID и необходимо удалять
	// TODO: Удалять ключи по TTL ?
	Delete(orderID int64) error
	// Lock блокирует Observer
	Lock() error
	// Unlock разблокирует Observer
	Unlock()
	// IsObservable проверяет является ли orderID Observable (пользовательским) ордером
	IsObservable(orderID int64) bool
}

type serverableObject struct {
	side    string
	symbol  string
	orderID int64
}

// WebsocketObserver реализация для Websocket
type WebsocketObserver struct {
	items   map[int64]serverableObject
	itemsMu sync.RWMutex
	mu      sync.Mutex
}

func (w *WebsocketObserver) IsObservable(orderID int64) bool {
	w.itemsMu.RLock()
	defer w.itemsMu.RUnlock()

	if _, ok := w.items[orderID]; ok {
		return true
	}

	return false
}

func (w *WebsocketObserver) Observe(side string, symbol string, orderID int64) error {
	w.itemsMu.RLock()

	if _, ok := w.items[orderID]; ok {
		w.itemsMu.RUnlock()
		return fmt.Errorf("already exists: %d", orderID)
	}

	w.itemsMu.RUnlock()
	w.itemsMu.Lock()
	w.items[orderID] = serverableObject{
		side:    side,
		symbol:  symbol,
		orderID: orderID,
	}
	w.itemsMu.Unlock()

	return nil
}

func (w *WebsocketObserver) Delete(orderID int64) error {
	w.itemsMu.RLock()
	if _, ok := w.items[orderID]; !ok {
		w.itemsMu.RUnlock()
		return fmt.Errorf("not found: %d", orderID)
	}
	w.itemsMu.RUnlock()

	w.itemsMu.Lock()
	delete(w.items, orderID)
	w.itemsMu.Unlock()

	return nil
}

// Lock
// TODO: Сделать кастомный Locker, чтобы возвращать ошибку, что блокировка длится дольше T
func (w *WebsocketObserver) Lock() error {
	w.mu.Lock()
	return nil
}

func (w *WebsocketObserver) Unlock() {
	w.mu.Unlock()
}

func NewWebsocketObserver() *WebsocketObserver {
	return &WebsocketObserver{
		items: make(map[int64]serverableObject),
	}
}

// NilObserver пустая реализации без синхронизаций. Используется, если получение трейдов из WebSocket не нужен
type NilObserver struct{}

func (os *NilObserver) Observe(_ string, _ string, _ int64) error {
	return nil
}

func (os *NilObserver) Delete(_ int64) error {
	return nil
}

func (os *NilObserver) Lock() error {
	return nil
}

func (os *NilObserver) Unlock() {

}

func (w *NilObserver) IsObservable(orderID int64) bool {
	return false
}
