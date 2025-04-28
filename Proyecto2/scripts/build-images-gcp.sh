#!/bin/bash
set -e

# Get project ID from gcloud config
PROJECT_ID=$(gcloud config get-value project)
if [ -z "$PROJECT_ID" ]; then
    echo "No project ID found in gcloud config. Please set it with: gcloud config set project YOUR_PROJECT_ID"
    exit 1
fi

# Configure Docker to use gcloud as a credential helper
gcloud auth configure-docker us-central1-docker.pkg.dev

REPO="us-central1-docker.pkg.dev/${PROJECT_ID}/proyecto2-repo"

echo "Building Docker images for Proyecto2 and pushing to GCP Artifact Registry..."

# Build Rust API
echo "Building Rust API..."
cd rust-api
docker build -t ${REPO}/rust-api:latest .
docker push ${REPO}/rust-api:latest
cd ..

# Build Go API components
echo "Building Go API components..."
cd go-api
docker build -f Dockerfile-client -t ${REPO}/go-api:latest .
docker push ${REPO}/go-api:latest

docker build -f Dockerfile-rabbitmq -t ${REPO}/grpc-rabbitmq:latest .
docker push ${REPO}/grpc-rabbitmq:latest

docker build -f Dockerfile-kafka -t ${REPO}/grpc-kafka:latest .
docker push ${REPO}/grpc-kafka:latest
cd ..

# Build Consumers
echo "Building Consumers..."
cd consumers
docker build -f Dockerfile-kafka -t ${REPO}/kafka-consumer:latest .
docker push ${REPO}/kafka-consumer:latest

docker build -f Dockerfile-rabbitmq -t ${REPO}/rabbitmq-consumer:latest .
docker push ${REPO}/rabbitmq-consumer:latest
cd ..

# Build Locust
echo "Building Locust..."
cd locust
docker build -t ${REPO}/locust:latest .
docker push ${REPO}/locust:latest
cd ..

echo "All images built and pushed to GCP Artifact Registry!"