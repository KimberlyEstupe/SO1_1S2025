// grpc/rabbit_client.go
package main

import (
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
	pb "weather-app/proto"
)

type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRabbitMQClient(amqpURL string) (*RabbitMQClient, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Declarar queue
	queue, err := ch.QueueDeclare(
		"weather_tweets", // nombre
		true,             // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	log.Printf("Cola RabbitMQ declarada: %s", queue.Name)

	return &RabbitMQClient{
		conn:    conn,
		channel: ch,
	}, nil
}

func (c *RabbitMQClient) PublishMessage(tweet *pb.WeatherTweet) error {
	// Convertir mensaje a JSON
	msgBody, err := json.Marshal(map[string]string{
		"descripcion": tweet.Description,
		"Pais":        tweet.Country,
		"Clima":       tweet.Weather,
	})
	if err != nil {
		return err
	}

	// Publicar mensaje
	err = c.channel.Publish(
		"",               // exchange
		"weather_tweets", // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msgBody,
			Persistent:  true,
		})

	if err != nil {
		return err
	}

	return nil
}

func (c *RabbitMQClient) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}