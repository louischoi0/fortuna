package mq

import (
	"log"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	Conn *amqp.Connection
}

var (
	instance *RabbitMQ
	once     sync.Once
)

func GetRabbitMQInstance(url string) (*RabbitMQ, error) {
	var initErr error
	once.Do(func() {
		var conn *amqp.Connection
		for i := 0; i < 500; i++ {
			conn, initErr = amqp.Dial(url)
			if initErr == nil {
				break
			}
			log.Printf("Retrying RabbitMQ connection: %v", initErr)
			time.Sleep(2 * time.Second)
		}
		if initErr == nil {
			instance = &RabbitMQ{
				Conn: conn,
			}
		}
	})

	if initErr != nil {
		return nil, initErr
	}
	return instance, nil
}

func (r *RabbitMQ) Close() {
	if r.Conn != nil {
		r.Conn.Close()
	}
}
