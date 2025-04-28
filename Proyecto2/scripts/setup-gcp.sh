#!/bin/bash
set -e

# Check if gcloud is installed
if ! command -v gcloud &> /dev/null; then
    echo "gcloud CLI is not installed. Please install it first."
    exit 1
fi

# Get project ID from gcloud config
PROJECT_ID=$(gcloud config get-value project)
if [ -z "$PROJECT_ID" ]; then
    echo "No project ID found in gcloud config. Please set it with: gcloud config set project YOUR_PROJECT_ID"
    exit 1
fi

# Enable required GCP APIs
echo "Enabling required GCP APIs..."
gcloud services enable container.googleapis.com \
                       compute.googleapis.com \
                       artifactregistry.googleapis.com

# Create GKE cluster
echo "Creating GKE cluster (this may take a few minutes)..."
gcloud container clusters create proyecto2-cluster \
    --zone us-central1-a \
    --num-nodes 3 \
    --machine-type e2-standard-2

# Configure kubectl to use the cluster
gcloud container clusters get-credentials proyecto2-cluster --zone us-central1-a

# Create Artifact Registry repository
echo "Creating Artifact Registry repository..."
gcloud artifacts repositories create proyecto2-repo \
    --repository-format=docker \
    --location=us-central1 \
    --description="Docker repository for Proyecto2"

echo "Setup complete! Now you need to build and push your container images to the Artifact Registry."
echo "Then deploy your application using the deploy-gcp.sh script."