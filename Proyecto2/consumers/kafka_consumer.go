// consumers/kafka_consumer.go
package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
	"github.com/go-redis/redis/v8"
)

type WeatherTweet struct {
	Descripcion string `json:"descripcion"`
	Pais        string `json:"Pais"`
	Clima       string `json:"Clima"`
}

type Consumer struct {
	ready  chan bool
	redis  *redis.Client
	ctx    context.Context
	wg     *sync.WaitGroup
}

func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(consumer.ready)
	return nil
}

func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

//TODO ConsumeClaim
/*func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Println("Canal de mensajes cerrado")
				return nil
			}

			// Procesar mensaje en una gorout*/