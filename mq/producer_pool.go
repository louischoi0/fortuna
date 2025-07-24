package mq

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

type QueueEvent struct {
	SlotKey   int64
	Body      []byte
	EventType string
}

type ProducePool struct {
	producers []*Producer
	mu        []sync.Mutex
	size      int

	Queue          []*QueueEvent
	OrderEventChan chan *QueueEvent
}

var producePool *ProducePool
var ponce sync.Once

func GetProducePool(conn *RabbitMQ) *ProducePool {
	ponce.Do(func() {
		const (
			exchangeName = "order.exchange"
			routingKey   = "order.commit.v2.broadcast"
		)
		if producePool == nil {
			pool, err := NewProducePool(conn, exchangeName, routingKey, 32)
			if err != nil {
				log.Fatalf("failed to create produce pool: %v", err)
			}
			producePool = pool
		}

		go producePool.Start()
	})
	return producePool
}

func NewProducePool(conn *RabbitMQ, exchange, routingKey string, n int) (*ProducePool, error) {
	if n <= 0 {
		return nil, errors.New("pool size must be greater than 0")
	}

	pool := &ProducePool{
		producers:      make([]*Producer, n),
		mu:             make([]sync.Mutex, n),
		size:           n,
		OrderEventChan: make(chan *QueueEvent, 1000),
	}

	for i := 0; i < n; i++ {
		producer, err := NewProducer(conn, exchange, routingKey)
		if err != nil {
			return nil, fmt.Errorf("failed to create producer %d: %w", i, err)
		}
		pool.producers[i] = producer
	}

	return pool, nil
}

func (pool *ProducePool) Start() {
	go func() {
		for event := range pool.OrderEventChan {
			pool._PublishEvent(event.SlotKey, event.Body, event.EventType)
		}
	}()
}

func (pool *ProducePool) _PublishEvent(slotKey int64, body []byte, eventType string) error {
	k := int(slotKey % int64(pool.size))

	pool.mu[k].Lock()
	defer pool.mu[k].Unlock()

	producer := pool.producers[k]
	if producer == nil {
		return errors.New("producer is nil")
	}

	return producer.PublishEvent(body, eventType)
}

func (pool *ProducePool) PublishEvent(slotKey int64, body []byte, eventType string) error {
	pool.OrderEventChan <- &QueueEvent{SlotKey: slotKey, Body: body, EventType: eventType}
	return nil
}
