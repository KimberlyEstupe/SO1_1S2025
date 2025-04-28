#!/bin/bash
set -e

# Get project ID from gcloud config
PROJECT_ID=$(gcloud config get-value project)
if [ -z "$PROJECT_ID" ]; then
    echo "No project ID found in gcloud config. Please set it with: gcloud config set project YOUR_PROJECT_ID"
    exit 1
fi

REPO="us-central1-docker.pkg.dev/${PROJECT_ID}/proyecto2-repo"

# Create temporary directory to store modified manifests
mkdir -p kubernetes/gcp-manifests

# Copy and modify Kubernetes manifests to use GCP image paths
for file in kubernetes/manifests/*.yaml; do
    sed "s|localhost:5000/proyecto2|${REPO}|g" "$file" > "kubernetes/gcp-manifests/$(basename "$file")"
done

echo "Deploying Proyecto2 to GCP GKE cluster..."

# Apply Kubernetes manifests
kubectl apply -f kubernetes/gcp-manifests/00-namespace.yaml
kubectl apply -f kubernetes/gcp-manifests/01-kafka-rabbitmq.yaml
kubectl apply -f kubernetes/gcp-manifests/02-redis-valkey.yaml
kubectl apply -f kubernetes/gcp-manifests/03-go-services.yaml
kubectl apply -f kubernetes/gcp-manifests/04-rust-api.yaml
kubectl apply -f kubernetes/gcp-manifests/05-consumers.yaml
kubectl apply -f kubernetes/gcp-manifests/07-grafana.yaml

# Install NGINX Ingress Controller
echo "Installing NGINX Ingress Controller..."
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.0/deploy/static/provider/cloud/deploy.yaml

# Wait for NGINX to be ready
echo "Waiting for NGINX Ingress Controller to be ready..."
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=180s

# Apply ingress after NGINX is ready
kubectl apply -f kubernetes/gcp-manifests/06-ingress.yaml

# Get the external IP of the ingress
echo "Waiting for ingress to get external IP..."
external_ip=""
while [ -z $external_ip ]; do
    echo "Waiting for external IP..."
    external_ip=$(kubectl get ing -n proyecto2 weather-ingress -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    [ -z "$external_ip" ] && sleep 10
done

echo "Deployment complete!"
echo "The application is accessible at http://${external_ip}.nip.io/input"
echo "Grafana dashboard is available at http://$(kubectl get svc -n proyecto2 grafana -o jsonpath='{.status.loadBalancer.ingress[0].ip}'):3000 (default user/pass: admin/admin)"