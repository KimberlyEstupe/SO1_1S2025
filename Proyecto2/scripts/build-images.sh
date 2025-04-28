#!/bin/bash
set -e

echo "Building Docker images for Proyecto2..."

# Create a local registry if not already running
docker ps | grep -q registry || docker run -d -p 5000:5000 --name registry registry:2

# Build Rust API
echo "Building Rust API..."
cd rust-api
docker build -t localhost:5000/proyecto2/rust-api:latest .
docker push localhost:5000/proyecto2/rust-api:latest
cd ..

# Build Go API components
echo "Building Go API components..."
cd go-api
docker build -f Dockerfile-client -t localhost:5000/proyecto2/go-api:latest .
docker push localhost:5000/proyecto2/go-api:latest

docker build -f Dockerfile-rabbitmq -t localhost:5000/proyecto2/grpc-rabbitmq:latest .
docker push localhost:5000/proyecto2/grpc-rabbitmq:latest

docker build -f Dockerfile-kafka -t localhost:5000/proyecto2/grpc-kafka:latest .
docker push localhost:5000/proyecto2/grpc-kafka:latest
cd ..

# Build Consumers
echo "Building Consumers..."
cd consumers
docker build -f Dockerfile-kafka -t localhost:5000/proyecto2/kafka-consumer:latest .
docker push localhost:5000/proyecto2/kafka-consumer:latest

docker build -f Dockerfile-rabbitmq -t localhost:5000/proyecto2/rabbitmq-consumer:latest .
docker push localhost:5000/proyecto2/rabbitmq-consumer:latest
cd ..

# Build Locust
echo "Building Locust..."
cd locust
docker build -t localhost:5000/proyecto2/locust:latest .
docker push localhost:5000/proyecto2/locust:latest
cd ..

echo "All images built and pushed to local registry!"