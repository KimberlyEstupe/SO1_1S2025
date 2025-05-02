#!/bin/bash

# Apply namespace
kubectl apply -f kubernetes/manifests/namespaces.yaml

# Apply configmaps and secrets
kubectl apply -f kubernetes/manifests/configmaps.yaml

# Deploy infrastructure components
kubectl apply -f kubernetes/manifests/kafka.yaml
kubectl apply -f kubernetes/manifests/rabbitmq.yaml
kubectl apply -f kubernetes/manifests/redis.yaml
kubectl apply -f kubernetes/manifests/valkey.yaml

# Wait for infrastructure to be ready
echo "Waiting for infrastructure components to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/kafka -n proyecto2
kubectl wait --for=condition=available --timeout=300s deployment/rabbitmq -n proyecto2
kubectl wait --for=condition=available --timeout=300s deployment/redis -n proyecto2
kubectl wait --for=condition=available --timeout=300s deployment/valkey -n proyecto2

# Deploy application components
kubectl apply -f kubernetes/manifests/grpc-services.yaml
kubectl apply -f kubernetes/manifests/go-api.yaml
kubectl apply -f kubernetes/manifests/rust-api.yaml
kubectl apply -f kubernetes/manifests/consumers.yaml
kubectl apply -f kubernetes/manifests/grafana.yaml

# Wait for application components to be ready
echo "Waiting for application components to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/grpc-rabbitmq -n proyecto2
kubectl wait --for=condition=available --timeout=300s deployment/grpc-kafka -n proyecto2
kubectl wait --for=condition=available --timeout=300s deployment/go-api -n proyecto2
kubectl wait --for=condition=available --timeout=300s deployment/rust-api -n proyecto2
kubectl wait --for=condition=available --timeout=300s deployment/kafka-consumer -n proyecto2
kubectl wait --for=condition=available --timeout=300s deployment/rabbitmq-consumer -n proyecto2
kubectl wait --for=condition=available --timeout=300s deployment/grafana -n proyecto2

# Apply ingress last
kubectl apply -f kubernetes/manifests/ingress.yaml

echo "All components deployed. Access the application at:"
echo "- API: http://proyecto2.local/input"
echo "- Grafana: http://proyecto2.local/grafana"
echo "- Modify your hosts file to point proyecto2.local to your cluster IP"