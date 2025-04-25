// grpc/server.go
package main

import (
	"context"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	pb "weather-app/proto"
)

type weatherServer struct {
	pb.UnimplementedWeatherServiceServer
	rabbitClient *RabbitMQClient
	kafkaClient  *KafkaClient
}

func (s *weatherServer) ProcessWeatherTweets(ctx context.Context, batch *pb.WeatherTweetBatch) (*pb.ProcessResponse, error) {
	log.Printf("Recibido lote de %d tweets para procesar", len(batch.Tweets))

	// Enviar mensajes a RabbitMQ
	go func() {
		for _, tweet := range batch.Tweets {
			if err := s.rabbitClient.PublishMessage(tweet); err != nil {
				log.Printf("Error al publicar en RabbitMQ: %v", err)
			}
		}
	}()

	// Enviar mensajes a Kafka
	go func() {
		for _, tweet := range batch.Tweets {
			if err := s.kafkaClient.PublishMessage(tweet); err != nil {
				log.Printf("Error al publicar en Kafka: %v", err)
			}
		}
	}()

	return &pb.ProcessResponse{Success: true}, nil
}

func main() {
	// Configurar logger
	log.SetOutput(os.Stdout)
	log.Println("Iniciando servidor gRPC...")

	// Dirección del servidor
	serverAddr := os.Getenv("GRPC_SERVER_ADDR")
	if serverAddr == "" {
		serverAddr = ":50051"
	}

	// Crear cliente RabbitMQ
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://user:password@rabbitmq.message-broker.svc.cluster.local:5672/"
	}
	rabbitClient, err := NewRabbitMQClient(rabbitURL)
	if err != nil {
		log.Fatalf("Error al crear cliente RabbitMQ: %v", err)
	}
	defer rabbitClient.Close()

	// Crear cliente Kafka
	kafkaURL := os.Getenv("KAFKA_URL")
	if kafkaURL == "" {
		kafkaURL = "kafka.message-broker.svc.cluster.local:9092"
	}
	kafkaClient, err := NewKafkaClient(kafkaURL)
	if err != nil {
		log.Fatalf("Error al crear cliente Kafka: %v", err)
	}
	defer kafkaClient.Close()

	// Crear y configurar el servidor gRPC
	lis, err := net.Listen("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Error al escuchar: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterWeatherServiceServer(s, &weatherServer{
		rabbitClient: rabbitClient,
		kafkaClient:  kafkaClient,
	})

	log.Printf("Servidor gRPC escuchando en %s", serverAddr)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Error al servir: %v", err)
	}
}