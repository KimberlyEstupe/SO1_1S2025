// api/main.go
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "weather-app/proto"
)

type WeatherTweet struct {
	Descripcion string `json:"descripcion"`
	Pais        string `json:"Pais"`
	Clima       string `json:"Clima"`
}

func main() {
	// Configurar logger
	log.SetOutput(os.Stdout)
	log.Println("Iniciando API REST de Go...")

	// Dirección del servicio gRPC
	grpcAddr := os.Getenv("GRPC_SERVER_ADDR")
	if grpcAddr == "" {
		grpcAddr = "localhost:50051"
	}

	// Función para manejar las peticiones HTTP
	http.HandleFunc("/process", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
			return
		}

		// Decodificar el JSON del cuerpo de la petición
		var tweets []WeatherTweet
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&tweets); err != nil {
			http.Error(w, "Error al decodificar JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		log.Printf("Recibidos %d tweets para procesar", len(tweets))

		// Conectar con el servidor gRPC
		conn, err := grpc.Dial(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("Error al conectar con servidor gRPC: %v", err)
			http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		// Crear cliente gRPC
		client := pb.NewWeatherServiceClient(conn)

		// Convertir tweets al formato del protobuf
		var pbTweets []*pb.WeatherTweet
		for _, tweet := range tweets {
			pbTweets = append(pbTweets, &pb.WeatherTweet{
				Description: tweet.Descripcion,
				Country:     tweet.Pais,
				Weather:     tweet.Clima,
			})
		}

		// Llamar al método gRPC para procesar los tweets
		_, err = client.ProcessWeatherTweets(context.Background(), &pb.WeatherTweetBatch{
			Tweets: pbTweets,
		})

		if err != nil {
			log.Printf("Error al llamar a ProcessWeatherTweets: %v", err)
			http.Error(w, "Error al procesar tweets: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Tweets procesados correctamente"))
	})

	// Endpoint para health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Healthy"))
	})

	// Iniciar servidor HTTP
	serverAddr := ":8080"
	log.Printf("Servidor HTTP escuchando en %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Error al iniciar servidor HTTP: %v", err)
	}
}