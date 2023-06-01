package worker

import (
	"encoding/json"
	"github.com/VladimirMovsesyan/praktikum-gophermart/internal/model"
	"io"
	"log"
	"net/http"
	"time"
)

type repository interface {
	UpdateOrder(login string, order model.Order) error
	GetOrderOwner(orderNum string) (login string, err error)
	GetProcessingOrders() ([]model.Order, error)
}

const updateInterval = 2 * time.Second

type workerPool struct {
	size         int
	addr         string
	orderC       chan model.Order
	storage      repository
	updateTicker *time.Ticker
}

func NewWorkerPool(workersCnt int, address string, storage repository) *workerPool {
	return &workerPool{
		size:         workersCnt,
		addr:         address,
		orderC:       make(chan model.Order),
		storage:      storage,
		updateTicker: time.NewTicker(updateInterval),
	}
}

func (wp *workerPool) Run() {
	for i := 0; i < wp.size; i++ {
		wp.newUpdateWorker()
	}

	wp.newRequestWorker()
}

func (wp *workerPool) newUpdateWorker() {
	go func() {
		for order := range wp.orderC {
			resp, err := http.Get(wp.addr + order.Number)
			if err != nil {
				log.Println(err)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				log.Printf("Error: got status code: %d", resp.StatusCode)
				continue
			}

			bytes, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println(err)
				continue
			}

			err = resp.Body.Close()
			if err != nil {
				log.Println(err)
			}

			err = json.Unmarshal(bytes, &order)
			if err != nil {
				log.Println(err)
				continue
			}

			if order.Status == model.OrderStatusRegistered {
				order.Status = model.OrderStatusNew
			}

			login, err := wp.storage.GetOrderOwner(order.Number)
			if err != nil {
				log.Println(err)
				continue
			}

			err = wp.storage.UpdateOrder(login, order)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}()
}

func (wp *workerPool) newRequestWorker() {
	go func() {
		for range wp.updateTicker.C {
			orders, err := wp.storage.GetProcessingOrders()
			if err != nil {
				log.Println(err)
				break
			}
			for _, order := range orders {
				wp.AddOrder(order)
			}
		}
	}()
}

func (wp *workerPool) AddOrder(order model.Order) {
	wp.orderC <- order
}

func (wp *workerPool) Stop() {
	wp.updateTicker.Stop()
	close(wp.orderC)
}
