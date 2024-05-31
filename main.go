package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"time"

	rs "github.com/fridrock/rabbitsimplier"
	amqp "github.com/rabbitmq/amqp091-go"
)

type CreateExGroupEvent struct {
	UserId int32  `json:"user_id"`
	Name   string `json:"name"`
}

func createExGroup() CreateExGroupEvent {
	exg := CreateExGroupEvent{}
	exg.Name = "Back"
	exg.UserId = 1
	return exg
}
func SendExerciseGroup(br rs.Producer) {
	//creating and marshalling event
	exg := createExGroup()
	res, err := json.Marshal(&exg)
	if err != nil {
		slog.Error("failed to marshal event on creating ex group")
	}

	//sending message
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err = br.PublishMessage(ctx, "sport_bot", "trainings.exgroup.create", string(res))
	if err != nil {
		slog.Error(fmt.Errorf("error sending message: %v", err).Error())
	}
}

func setupConfigurer() *rs.RConfigurer {
	brc := &rs.RConfigurer{}
	err := brc.Configure(rs.Config{
		Username: "guest",
		Password: "guest",
		Host:     "localhost",
	})
	if err != nil {
		log.Fatal("error creating RabbitMQ connection")
	}
	return brc
}

func setupConsumer(brc *rs.RConfigurer) *rs.RConsumer {
	brConsumer := &rs.RConsumer{}
	err := brConsumer.CreateChannel(brc.GetConnection())
	if err != nil {
		log.Fatal("error creating channel for consumer")
	}
	q, err := brConsumer.CreateQueue()
	if err != nil {
		log.Fatal("error creating queue")
	}
	err = brConsumer.SetBinding(q, "tgbot.exgroup.#", "sport_bot")
	if err != nil {
		log.Fatal("error creating binding for exgroup queue")
	}
	rdispatcher := rs.NewRDispacher()
	rdispatcher.RegisterHandler("tgbot.exgroup.create", rs.NewHandlerFunc(func(msg amqp.Delivery) {
		slog.Info(fmt.Sprintf("Exgroup: %s was saved", string(msg.Body)))
	}))
	brConsumer.RegisterDispatcher(q, rdispatcher)
	return brConsumer
}

func setupProducer(brc rs.Configurer) *rs.RProducer {
	brp := &rs.RProducer{}
	err := brp.CreateChannel(brc.GetConnection())
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
	brp := setupProducer(brc)
	defer brp.Stop()

	brConsumer := setupConsumer(brc)
	defer brConsumer.Stop()
	//setting up producer
	SendExerciseGroup(brp)

	//for endless program listening
	var forever chan struct{}
	<-forever
}
