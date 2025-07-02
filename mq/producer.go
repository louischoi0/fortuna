package rabbit

import (
	// "errors"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

type Producer struct {
	Channel    *amqp.Channel
	Exchange   string
	RoutingKey string

	ackChan  chan uint64
	nackChan chan uint64

	mu sync.Mutex
}

func NewProducer(conn *RabbitMQ, exchange, routingKey string) (*Producer, error) {
	ch, err := conn.Conn.Channel()

	if err != nil {
		return nil, err
	}

	err = ch.ExchangeDeclare(
		exchange,
		"topic", // assumed topic exchange
		true,    // durable
		false,   // auto-delete
		false,   // internal
		true,    // no-wait
		nil,     // args
	)
	if err != nil {
		return nil, err
	}

	err = ch.Confirm(false)
	if err != nil {
		log.Fatal(err)
	}
	ackChan, nackChan := ch.NotifyConfirm(make(chan uint64, 1000), make(chan uint64, 1000))

	return &Producer{
		Channel:    ch,
		Exchange:   exchange,
		RoutingKey: routingKey,
		ackChan:    ackChan,
		nackChan:   nackChan,
	}, nil
}

func (p *Producer) Publish(body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.Channel.Publish(
		p.Exchange,
		p.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *Producer) Close() error {
	return p.Channel.Close()
}

func (p *Producer) PublishEvent(body []byte, eventType string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.Channel.Publish(
		p.Exchange, p.RoutingKey, false, false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Headers:      amqp.Table{"event": eventType},
		},
	)
	if err != nil {
		log.Fatalf(err.Error())
		return err
	}

	select {
	case <-p.ackChan:
		log.Printf("%s end", time.Now())
		return nil
	case <-p.nackChan:
		log.Fatalf("message negatively acknowledged")
		return errors.New("message negatively acknowledged")
	case <-time.After(5 * time.Second):
		return errors.New("confirm timeout")
	}

	return nil
}
