package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/fridrock/tgsportbot/broker"
)

type ExerciseGroup struct {
	Id     int32  `json:"id"`
	UserId int32  `json:"user_id"`
	Name   string `json:"name"`
}

func createExGroup() ExerciseGroup {
	exg := ExerciseGroup{}
	exg.Name = "Back"
	exg.UserId = 1
	return exg
}
func SendExerciseGroup(br broker.BrokerWorker) {
	exg := createExGroup()
	res, _ := json.Marshal(&exg)
	fmt.Println(string(res))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err := br.SendMessage(ctx, string(res), "ex_group", "create")
	if err != nil {
		slog.Error(fmt.Errorf("error sending message: %v", err).Error())
	}
	ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	err = br.SendMessage(ctx, "Hello world", "ex_group", "update")
}

func main() {
	br := broker.CreateBroker()
	defer br.Stop()
	SendExerciseGroup(br)
	var forever chan struct{}

	<-forever
}
