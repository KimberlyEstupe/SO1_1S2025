package main

import (
	"context"
	"log"
	"net"
	"os"
	"strings"

	"google.golang.org/grpc"
	pb "weather/proto"

	"github.com/segmentio/kafka-go"
)

type server struct {
	pb.UnimplementedWeatherServiceServer
	kafkaWriter *kafka.Writer
}

func (s *server) PublishToKafka(ctx context.Context, req *pb.WeatherDataList) (*pb.PublishResponse, error) {
	log.Printf("Received request to publish %d messages to Kafka", len(req.Data))

	messages := make([]kafka.Message, len(req.Data))
	for i, data := range req.Data {
		messages[i] = kafka.Message{
			Key:   []byte(data.Pais),
			Value: []byte(data.String()),
		}
	}

	err := s.kafkaWriter.WriteMessages(ctx, messages...)
	if err != nil {
		log.Printf("Failed to write Kafka messages: %v", err)
		return &pb.PublishResponse{
			Success: false,
			Message: "Failed to write Kafka messages: " + err.Error(),
		}, nil
	}

	return &pb.PublishResponse{
		Success: true,
		Message: "Successfully published to Kafka",
	}, nil
}

func (s *server) PublishToRabbitMQ(ctx context.Context, req *pb.WeatherDataList) (*pb.PublishResponse, error) {
	// This server only handles Kafka, forward to RabbitMQ is done by the other instance
	return &pb.PublishResponse{
		Success: true,
		Message: "Kafka server - RabbitMQ publishing handled by other server",
	}, nil
}

func main() {
	// Get Kafka connection parameters from environment
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	if kafkaBrokers == "" {
		kafkaBrokers = "kafka:9092"
	}

	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "weather_topic"
	}

	brokerList := strings.Split(kafkaBrokers, ",")

	// Create Kafka writer
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokerList,
		Topic:    kafkaTopic,
		Balancer: &kafka.LeastBytes{},
	})

	// Set up gRPC server
	port := os.Getenv("PORT")
	if port == "" {
		port = "50052"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWeatherServiceServer(s, &server{
		kafkaWriter: writer,
	})

	log.Printf("Starting gRPC server on port %s", port)
	log.Printf("Connected to Kafka at %s, topic: %s", kafkaBrokers, kafkaTopic)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}