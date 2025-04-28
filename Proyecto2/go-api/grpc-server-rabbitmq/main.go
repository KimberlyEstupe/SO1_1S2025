package main

import (
	"context"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	pb "weather/proto"

	amqp "github.com/rabbitmq/amqp091-go"
)

type server struct {
	pb.UnimplementedWeatherServiceServer
	rabbitMQConn *amqp.Connection
	rabbitMQChan *amqp.Channel
}

func (s *server) PublishToRabbitMQ(ctx context.Context, req *pb.WeatherDataList) (*pb.PublishResponse, error) {
	log.Printf("Received request to publish %d messages to RabbitMQ", len(req.Data))

	queueName := "weather_queue"
	_, err := s.rabbitMQChan.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Printf("Failed to declare queue: %v", err)
		return &pb.PublishResponse{
			Success: false,
			Message: "Failed to declare queue: " + err.Error(),
		}, nil
	}

	for _, data := range req.Data {
		body := []byte(data.String())
		err = s.rabbitMQChan.PublishWithContext(
			ctx,
			"",        // exchange
			queueName, // routing key
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			})
		if err != nil {
			log.Printf("Failed to publish message: %v", err)
			return &pb.PublishResponse{
				Success: false,
				Message: "Failed to publish message: " + err.Error(),
			}, nil
		}
	}

	return &pb.PublishResponse{
		Success: true,
		Message: "Successfully published to RabbitMQ",
	}, nil
}

func (s *server) PublishToKafka(ctx context.Context, req *pb.WeatherDataList) (*pb.PublishResponse, error) {
	// This server only handles RabbitMQ, forward to Kafka server is done by the other instance
	return &pb.PublishResponse{
		Success: true,
		Message: "RabbitMQ server - Kafka publishing handled by other server",
	}, nil
}

func main() {
	// Get RabbitMQ connection parameters from environment
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@rabbitmq:5672/"
	}

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

	// Set up gRPC server
	port := os.Getenv("PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWeatherServiceServer(s, &server{
		rabbitMQConn: conn,
		rabbitMQChan: ch,
	})

	log.Printf("Starting gRPC server on port %s", port)
	log.Printf("Connected to RabbitMQ at %s", rabbitMQURL)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}