package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-redis/redis/v8"
	amqp "github.com/rabbitmq/amqp091-go"
)

type WeatherData struct {
	Descripcion string `json:"descripcion"`
	Pais        string `json:"Pais"`
	Clima       string `json:"Clima"`
}

func main() {
	// Get configuration from environment variables
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	queueName := os.Getenv("RABBITMQ_QUEUE")
	if queueName == "" {
		queueName = "weather_queue"
	}

	valkeyAddr := os.Getenv("VALKEY_ADDR")
	if valkeyAddr == "" {
		valkeyAddr = "valkey:6379"
	}

	// Number of worker goroutines
	numWorkers := 10

	// Set up Valkey client (using Redis client as they're compatible)
	valkeyClient := redis.NewClient(&redis.Options{
		Addr: valkeyAddr,
	})
	defer valkeyClient.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Verify Valkey connection
	_, err := valkeyClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Valkey: %v", err)
	}

	log.Printf("Connected to Valkey at %s", valkeyAddr)

	// Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	// Declare queue
	q, err := ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
	}

	// Set QoS
	err = ch.Qos(
		10,    // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		log.Fatalf("Failed to set QoS: %v", err)
	}

	log.Printf("Connected to RabbitMQ at %s", rabbitMQURL)

	// Create wait group for workers
	var wg sync.WaitGroup

	// Start consumers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			consumeMessages(ctx, ch, q.Name, valkeyClient, workerId)
		}(i)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Received shutdown signal, closing consumer...")
	cancel()
	wg.Wait()
	log.Println("Consumer shut down gracefully")
}

func consumeMessages(ctx context.Context, ch *amqp.Channel, queueName string, valkeyClient *redis.Client, workerId int) {
	msgs, err := ch.Consume(
		queueName,           // queue
		"consumer-"+string(rune(workerId)), // consumer
		false,              // auto-ack
		false,              // exclusive
		false,              // no-local
		false,              // no-wait
		nil,                // args
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Printf("Worker %d started consuming from queue %s", workerId, queueName)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d shutting down", workerId)
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Printf("Worker %d channel closed", workerId)
				return
			}

			var weatherData WeatherData
			if err := json.Unmarshal(msg.Body, &weatherData); err != nil {
				log.Printf("Worker %d error unmarshaling message: %v", workerId, err)
				msg.Ack(false)
				continue
			}

			// Process and store in Valkey
			pipe := valkeyClient.Pipeline()
			
			// Increment country counter
			pipe.HIncrBy(ctx, "rabbitmq:countries", weatherData.Pais, 1)
			
			// Increment weather type counter
			pipe.HIncrBy(ctx, "rabbitmq:weather", weatherData.Clima, 1)
			
			// Increment total counter
			pipe.Incr(ctx, "rabbitmq:total_messages")
			
			// Execute pipeline
			_, err = pipe.Exec(ctx)
			if err != nil {
				log.Printf("Worker %d error storing data in Valkey: %v", workerId, err)
				msg.Nack(false, true) // Requeue the message
				continue
			}

			msg.Ack(false)
			log.Printf("Worker %d processed message from %s: %s", workerId, weatherData.Pais, weatherData.Clima)
		}
	}
}