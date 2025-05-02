#!/bin/bash

# Create a local Docker registry if not already running
if ! docker ps | grep -q registry; then
  docker run -d -p 5000:5000 --name registry registry:2
fi

# Build Rust API
echo "Building Rust API..."
cd rust-api
docker build -t localhost:5000/proyecto2/rust-api:latest .
docker push localhost:5000/proyecto2/rust-api:latest
cd ..

# Build Go API gRPC client
echo "Building Go API gRPC client..."
cd go-api/grpc-client
docker build -t localhost:5000/proyecto2/go-api-grpc-client:latest .
docker push localhost:5000/proyecto2/go-api-grpc-client:latest
cd ../..

# Build Go API gRPC server for RabbitMQ
echo "Building Go API gRPC server for RabbitMQ..."
cd go-api/grpc-server-rabbitmq
docker build -t localhost:5000/proyecto2/go-api-grpc-server-rabbitmq:latest .
docker push localhost:5000/proyecto2/go-api-grpc-server-rabbitmq:latest
cd ../..

# Build Go API gRPC server for Kafka
echo "Building Go API gRPC server for Kafka..."
cd go-api/grpc-server-kafka
docker build -t localhost:5000/proyecto2/go-api-grpc-server-kafka:latest .
docker push localhost:5000/proyecto2/go-api-grpc-server-kafka:latest
cd ../..

# Build Kafka consumer
echo "Building Kafka consumer..."
cd consumers/kafka-consumer
docker build -t localhost:5000/proyecto2/kafka-consumer:latest .
docker push localhost:5000/proyecto2/kafka-consumer:latest
cd ../..

# Build RabbitMQ consumer
echo "Building RabbitMQ consumer..."
cd consumers/rabbitmq-consumer
docker build -t localhost:5000/proyecto2/rabbitmq-consumer:latest .
docker push localhost:5000/proyecto2/rabbitmq-consumer:latest
cd ../..

# Build Locust
echo "Building Locust..."
cd locust
docker build -t localhost:5000/proyecto2/locust:latest .
docker push localhost:5000/proyecto2/locust:latest
cd ..

echo "All images built and pushed to local registry."