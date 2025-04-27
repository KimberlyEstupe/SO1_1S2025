// grpc/kafka_client.go
package main

import (
	"encoding/json"
	"log"

	"github.com/Shopify/sarama"
	pb "weather-app/proto"
)

type KafkaClient struct {
	producer sarama.SyncProducer
}

func NewKafkaClient(brokers string) (*KafkaClient, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer([]string{brokers}, config)
	if err != nil {
		return nil, err
	}

	return &KafkaClient{
		producer: producer,
	}, nil
}

func (c *KafkaClient) PublishMessage(tweet *pb.WeatherTweet) error {
	// Convertir mensaje a JSON
	msgBody, err := json.Marshal(map[string]string{
		"descripcion": tweet.Description,
		"Pais":        tweet.Country,
		"Clima":       tweet.Weather,
	})
	if err != nil {
		return err
	}

	// Crear mensaje Kafka
	msg := &sarama.ProducerMessage{
		Topic: "weather_tweets",
		Value: sarama.StringEncoder(msgBody),
	}

	// Enviar mensaje
	partition, offset, err := c.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Mensaje enviado a Kafka topic %s [partition=%d] @ offset=%d",
		"weather_tweets", partition, offset)
	return nil
}

func (c *KafkaClient) Close() error {
	if c.producer != nil {
		return c.producer.Close()
	}
	return nil
}