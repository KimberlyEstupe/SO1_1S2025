#!/bin/bash
set -e

# Create a VM for Harbor
echo "Creating VM for Harbor..."
gcloud compute instances create harbor-vm \
    --zone=us-central1-a \
    --machine-type=e2-standard-2 \
    --subnet=default \
    --network-tier=PREMIUM \
    --tags=http-server,https-server \
    --image-family=ubuntu-2004-lts \
    --image-project=ubuntu-os-cloud \
    --boot-disk-size=50GB \
    --boot-disk-type=pd-standard

# Allow HTTP and HTTPS traffic
gcloud compute firewall-rules create allow-http \
    --direction=INGRESS \
    --action=ALLOW \
    --rules=tcp:80 \
    --target-tags=http-server

gcloud compute firewall-rules create allow-https \
    --direction=INGRESS \
    --action=ALLOW \
    --rules=tcp:443 \
    --target-tags=https-server

# Get the external IP of the VM
HARBOR_IP=$(gcloud compute instances describe harbor-vm --zone=us-central1-a --format='get(networkInterfaces[0].accessConfigs[0].natIP)')

echo "Harbor VM created with IP: $HARBOR_IP"
echo "SSH into the VM and follow the instructions to install Harbor."
echo "gcloud compute ssh harbor-vm --zone=us-central1-a"