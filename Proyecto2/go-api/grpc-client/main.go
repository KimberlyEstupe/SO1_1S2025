package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "weather/proto"
)

type WeatherData struct {
	Descripcion string `json:"descripcion"`
	Pais        string `json:"Pais"`
	Clima       string `json:"Clima"`
}

func main() {
	// Get gRPC server address from environment or use default
	grpcServerAddr := os.Getenv("GRPC_SERVER_ADDR")
	if grpcServerAddr == "" {
		grpcServerAddr = "localhost:50051"
	}

	// Set up HTTP server
	http.HandleFunc("/process", handleWeatherData(grpcServerAddr))
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Healthy"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting HTTP server on port %s", port)
	log.Printf("Connecting to gRPC server at %s", grpcServerAddr)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func handleWeatherData(grpcServerAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only accept POST requests
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse JSON body
		var weatherDataList []WeatherData
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&weatherDataList); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		log.Printf("Received %d weather data entries", len(weatherDataList))

		// Connect to gRPC server
		conn, err := grpc.Dial(grpcServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Failed to connect to gRPC server: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		client := pb.NewWeatherServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Convert to protobuf message
		protoList := &pb.WeatherDataList{
			Data: make([]*pb.WeatherData, len(weatherDataList)),
		}

		for i, data := range weatherDataList {
			protoList.Data[i] = &pb.WeatherData{
				Descripcion: data.Descripcion,
				Pais:        data.Pais,
				Clima:       data.Clima,
			}
		}

		// Send to RabbitMQ
		rabbitResp, err := client.PublishToRabbitMQ(ctx, protoList)
		if err != nil {
			log.Printf("Failed to publish to RabbitMQ: %v", err)
			http.Error(w, "Failed to publish to RabbitMQ", http.StatusInternalServerError)
			return
		}

		// Send to Kafka
		kafkaResp, err := client.PublishToKafka(ctx, protoList)
		if err != nil {
			log.Printf("Failed to publish to Kafka: %v", err)
			http.Error(w, "Failed to publish to Kafka", http.StatusInternalServerError)
			return
		}

		// Create response
		response := map[string]interface{}{
			"rabbitmq": rabbitResp,
			"kafka":    kafkaResp,
			"count":    len(weatherDataList),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}