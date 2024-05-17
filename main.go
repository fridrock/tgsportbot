package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/fridrock/tgsportbot/broker"
	amqp "github.com/rabbitmq/amqp091-go"
)

// TODO make loose decoupling with .Ch and .Connection properties, so all methods use interfaces instead of implementations
type CreateExGroupEvent struct {
	Event  string `json:"event"`
	Id     int32  `json:"id"`
	UserId int32  `json:"user_id"`
	Name   string `json:"name"`
}

func createExGroup() CreateExGroupEvent {
	exg := CreateExGroupEvent{}
	exg.Event = "create"
	exg.Name = "Back"
	exg.UserId = 1
	return exg
}
func SendExerciseGroup(br broker.BrokerProducer) {
	//creating and marshalling event
	exg := createExGroup()
	res, err := json.Marshal(&exg)
	if err != nil {
		slog.Error("failed to marshal event on creating ex group")
	}

	//sending message
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err = br.PublishMessage(ctx, "sport_bot", "trainings.exgroup", string(res))
	if err != nil {
		slog.Error(fmt.Errorf("error sending message: %v", err).Error())
	}
}

func setupConfigurer() *broker.BrokerConfigurerImpl {
	brc := &broker.BrokerConfigurerImpl{}
	err := brc.Configure()
	if err != nil {
		log.Fatal("error creating RabbitMQ connection")
	}
	return brc
}

func setupConsumer(brc *broker.BrokerConfigurerImpl) *broker.BrokerConsumerImpl {
	brConsumer := &broker.BrokerConsumerImpl{}
	err := brConsumer.CreateChannel(brc.Connection)
	if err != nil {
		log.Fatal("error creating channel for consumer")
	}
	q, err := brConsumer.CreateQueue()
	if err != nil {
		log.Fatal("error creating queue")
	}
	err = brConsumer.SetBinding(q, "tgbot.exgroup", "sport_bot")
	if err != nil {
		log.Fatal("error creating binding for exgroup queue")
	}
	brConsumer.RegisterConsumer(q, func(msgs <-chan amqp.Delivery) {
		for d := range msgs {
			fmt.Println(string(d.Body))
		}
	})
	return brConsumer
}

func setupProducer(brc *broker.BrokerConfigurerImpl) *broker.BrokerProducerImpl {
	brp := &broker.BrokerProducerImpl{}
	err := brp.CreateChannel(brc.Connection)
	if err != nil {
		log.Fatal("error creating connection for producer")
	}
	err = brp.CreateExchange("sport_bot", "topic")
	if err != nil {
		log.Fatal("error creating application exchange")
	}
	return brp
}

func main() {
	//setting up connection
	brc := setupConfigurer()
	defer brc.Stop()
	brConsumer := setupConsumer(brc)
	defer brConsumer.Stop()
	//setting up producer
	brp := setupProducer(brc)
	defer brp.Stop()

	SendExerciseGroup(brp)

	//for endless program listening
	var forever chan struct{}
	<-forever
}
