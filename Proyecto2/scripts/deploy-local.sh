#!/bin/bash
set -e

echo "Deploying Proyecto2 to local Kubernetes cluster..."

# Apply Kubernetes manifests
kubectl apply -f kubernetes/manifests/00-namespace.yaml
kubectl apply -f kubernetes/manifests/01-kafka-rabbitmq.yaml
kubectl apply -f kubernetes/manifests/02-redis-valkey.yaml
kubectl apply -f kubernetes/manifests/03-go-services.yaml
kubectl apply -f kubernetes/manifests/04-rust-api.yaml
kubectl apply -f kubernetes/manifests/05-consumers.yaml
kubectl apply -f kubernetes/manifests/07-grafana.yaml

# Install NGINX Ingress Controller if not already installed
if ! kubectl get deployment -n ingress-nginx nginx-ingress-controller &> /dev/null; then
    echo "Installing NGINX Ingress Controller..."
    kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.0/deploy/static/provider/cloud/deploy.yaml
    
    # Wait for NGINX to be ready
    echo "Waiting for NGINX Ingress Controller to be ready..."
    kubectl wait --namespace ingress-nginx \
      --for=condition=ready pod \
      --selector=app.kubernetes.io/component=controller \
      --timeout=120s
fi

# Apply ingress after NGINX is ready
kubectl apply -f kubernetes/manifests/06-ingress.yaml

echo "Adding 'weather.local' to /etc/hosts if not already present..."
if ! grep -q "weather.local" /etc/hosts; then
    echo "127.0.0.1 weather.local" | sudo tee -a /etc/hosts
fi

echo "Deployment complete! The application should be accessible at http://weather.local/input"
echo "Grafana dashboard is available at http://localhost:3000 (default user/pass: admin/admin)"