package rabbit

import (
	"log"
	"github.com/streadway/amqp"
)

type Consumer struct {
	Channel *amqp.Channel
	Queue   string
	Handler func(amqp.Delivery)
}

func NewConsumer(channel *amqp.Channel, queueName string, handler func(amqp.Delivery)) (*Consumer, error) {
	_, err := channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		Channel: channel,
		Queue:   queueName,
		Handler: handler,
	}, nil
}

func (c *Consumer) Start() error {
	msgs, err := c.Channel.Consume(
		c.Queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			c.Handler(d)
		}
	}()

	log.Printf("Consuming from %s", c.Queue)
	return nil
}

func (c *Consumer) Close() error {
	return c.Channel.Close()
}
