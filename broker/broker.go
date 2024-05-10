package broker

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// TODO make message consuming
type BrokerWorker interface {
	Configure() error
	SendMessage(ctx context.Context, message string, path ...string) error
	Stop()
}

type BrokerWorkerImpl struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	exchanges map[string]map[string]string
}

func (bs *BrokerWorkerImpl) Stop() {
	bs.ch.Close()
	bs.conn.Close()
}
func (bs *BrokerWorkerImpl) Configure() error {
	connection, err := bs.getConnection()
	if err != nil {
		return err
	}
	bs.conn = connection
	channel, err := bs.getChannel()
	if err != nil {
		return err
	}
	bs.ch = channel
	bs.exchanges = make(map[string]map[string]string, 1)
	return nil
}

// TODO remake to getting url from .env
func (bs BrokerWorkerImpl) getConnection() (*amqp.Connection, error) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, err
	}
	return conn, err
}
func (bs BrokerWorkerImpl) getChannel() (*amqp.Channel, error) {
	channel, err := bs.conn.Channel()
	if err != nil {
		return nil, err
	}
	return channel, nil
}
func (bs *BrokerWorkerImpl) createExchange(exchangeName string) error {
	if _, ok := bs.exchanges[exchangeName]; ok {
		return nil
	}
	err := bs.ch.ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err == nil {
		bs.exchanges[exchangeName] = make(map[string]string)
	}
	return err
}
func (bs *BrokerWorkerImpl) createQueue(exchangeName, routingKey string) (queueName string, err error) {
	if _, ok := bs.exchanges[exchangeName][routingKey]; ok {
		return bs.exchanges[exchangeName][routingKey], nil
	}
	q, err := bs.ch.QueueDeclare(
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return "", err
	}
	return q.Name, nil
}
func (bs *BrokerWorkerImpl) bindQueue(queueName, exchangeName, routingKey string) error {
	if _, ok := bs.exchanges[exchangeName][routingKey]; ok {
		return nil
	}
	err := bs.ch.QueueBind(
		queueName,
		routingKey,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	bs.exchanges[exchangeName][routingKey] = queueName
	return nil
}

func (bs *BrokerWorkerImpl) SendMessage(ctx context.Context, message string, path ...string) error {
	exchangeName, routingKey := bs.getPath(path)
	err := bs.generatePath(exchangeName, routingKey)
	if err != nil {
		return err
	}
	err = bs.ch.PublishWithContext(ctx,
		exchangeName,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	return err
}
func (bs *BrokerWorkerImpl) generatePath(exchangeName, routingKey string) error {
	bs.createExchange(exchangeName)

	q, err := bs.createQueue(exchangeName, routingKey)
	if err != nil {
		return err
	}
	err = bs.bindQueue(q, exchangeName, routingKey)
	if err != nil {
		return err
	}
	return nil
}
func (bs BrokerWorkerImpl) getPath(path []string) (exchangeName, routingKey string) {
	paramsNum := len(path)
	for i := 0; i < 2-paramsNum; i++ {
		path = append(path, "default")
	}
	return path[0], path[1]
}

func CreateBroker() *BrokerWorkerImpl {
	br := &BrokerWorkerImpl{}
	err := br.Configure()
	if err != nil {
		log.Fatal(err.Error())
	}
	return br
}

func (bs BrokerWorkerImpl) Show() {
	fmt.Println(bs.exchanges)
}
