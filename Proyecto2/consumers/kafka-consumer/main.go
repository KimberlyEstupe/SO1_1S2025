package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
)

type WeatherData struct {
	Descripcion string `json:"descripcion"`
	Pais        string `json:"Pais"`
	Clima       string `json:"Clima"`
}

func main() {
	// Get configuration from environment variables
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka:9092"
	}

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "weather_topic"
	}

	kafkaGroupID := os.Getenv("KAFKA_GROUP_ID")
	if kafkaGroupID == "" {
		kafkaGroupID = "weather-consumer"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	// Number of worker goroutines
	numWorkers := 10

	// Set up Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	defer redisClient.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Verify Redis connection
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Printf("Connected to Redis at %s", redisAddr)

	// Create Kafka reader
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     strings.Split(kafkaBrokers, ","),
		Topic:       kafkaTopic,
		GroupID:     kafkaGroupID,
		StartOffset: kafka.FirstOffset,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
	})

	defer reader.Close()

	log.Printf("Connected to Kafka at %s", kafkaBrokers)

	// Create wait group for workers
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			processMessages(ctx, reader, redisClient, workerId)
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

func processMessages(ctx context.Context, reader *kafka.Reader, redisClient *redis.Client, workerId int) {
	log.Printf("Worker %d started", workerId)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d shutting down", workerId)
			return
		default:
			// Set a timeout for reading messages
			readCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			msg, err := reader.ReadMessage(readCtx)
			cancel()

			if err != nil {
				if err == context.Canceled {
					return
				}
				log.Printf("Worker %d error reading message: %v", workerId, err)
				continue
			}

			var weatherData WeatherData
			if err := json.Unmarshal(msg.Value, &weatherData); err != nil {
				log.Printf("Worker %d error unmarshaling message: %v", workerId, err)
				continue
			}

			// Process and store in Redis
			pipe := redisClient.Pipeline()
			
			// Increment country counter
			pipe.HIncrBy(ctx, "kafka:countries", weatherData.Pais, 1)
			
			// Increment weather type counter
			pipe.HIncrBy(ctx, "kafka:weather", weatherData.Clima, 1)
			
			// Increment total counter
			pipe.Incr(ctx, "kafka:total_messages")
			
			// Execute pipeline
			_, err = pipe.Exec(ctx)
			if err != nil {
				log.Printf("Worker %d error storing data in Redis: %v", workerId, err)
				continue
			}

			log.Printf("Worker %d processed message from %s: %s", workerId, weatherData.Pais, weatherData.Clima)
		}
	}
}