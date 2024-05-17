package broker

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type BrokerConfigurer interface {
	Configure() error
	Stop()
}

type BrokerProducer interface {
	CreateChannel(*amqp.Connection) error
	CreateExchange(exchangeName, exchangeType string) error
	PublishMessage(ctx context.Context, exchangeName, routingKey, body string) error
	Stop()
}

type BrokerConsumer interface {
	CreateChannel(*amqp.Connection) error
	CreateQueue(queueName ...string) (amqp.Queue, error)
	SetBinding(q amqp.Queue, boundingKey, exchangeName string) error
	RegisterConsumer(amqp.Queue, func(<-chan amqp.Delivery))
	Stop()
}

type BrokerConfigurerImpl struct {
	Connection *amqp.Connection
}

func (bci *BrokerConfigurerImpl) Configure() error {
	conn, err := amqp.Dial("amqp://guest:guest@localhost/")
	if err != nil {
		return err
	}
	bci.Connection = conn
	return nil
}

func (bci *BrokerConfigurerImpl) Stop() {
	bci.Connection.Close()
}

type BrokerProducerImpl struct {
	Ch *amqp.Channel
}

func (bpi *BrokerProducerImpl) CreateChannel(conn *amqp.Connection) error {
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	bpi.Ch = channel
	return nil
}

func (bpi *BrokerProducerImpl) CreateExchange(exchangeName, exchangeType string) error {
	return bpi.Ch.ExchangeDeclare(
		exchangeName,
		exchangeType,
		true,  //durable
		false, //autoDelete
		false, //internal
		false, //noWait
		nil,   //args
	)
}

func (bpi *BrokerProducerImpl) PublishMessage(ctx context.Context, exchangeName, routingKey, body string) error {
	return bpi.Ch.PublishWithContext(
		ctx,
		exchangeName,
		routingKey,
		false, //mandatory
		false, //immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
		})
}

func (bpi *BrokerProducerImpl) Stop() {
	bpi.Ch.Close()
}

type BrokerConsumerImpl struct {
	Ch *amqp.Channel
}

func (bci *BrokerConsumerImpl) CreateChannel(conn *amqp.Connection) error {
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	bci.Ch = channel
	return nil
}

func (bci *BrokerConsumerImpl) CreateQueue(queueName ...string) (amqp.Queue, error) {
	var queueNameStr string = ""
	if len(queueName) != 0 {
		queueNameStr = queueName[0]
	}
	var q amqp.Queue
	q, err := bci.Ch.QueueDeclare(
		queueNameStr,
		false, //durable
		true,  // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return q, err
	}
	return q, nil
}
func (bci *BrokerConsumerImpl) SetBinding(q amqp.Queue, boundingKey, exchangeName string) error {
	return bci.Ch.QueueBind(
		q.Name,
		boundingKey,
		exchangeName,
		false, // noWait
		nil,   // args
	)

}
func (bci *BrokerConsumerImpl) RegisterConsumer(q amqp.Queue, f func(<-chan amqp.Delivery)) error {
	msgs, err := bci.Ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}
	go f(msgs)

	return nil
}
func (bci *BrokerConsumerImpl) Stop() {
	bci.Ch.Close()
}
