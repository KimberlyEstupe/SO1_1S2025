#!/bin/bash

# Función para generar un nombre de contenedor aleatorio
generate_container_name() {
    # Usar /dev/urandom para generar un nombre único
    echo "container_$(date +%s%N | cut -b1-10)_$RANDOM"
}

# Función para crear contenedores aleatorios
create_random_containers() {
    for i in {1..10}; do
        # Elegir aleatoriamente un tipo de contenedor
        case $((RANDOM % 4)) in
            0) 
                docker run -d --name "cpu_$(generate_container_name)" --cpus="0.2" containerstack/alpine-stress stress --cpu 1
                ;;
            1) 
                docker run -d --name "ram_$(generate_container_name)" --memory="128M" --cpus="0.2" containerstack/alpine-stress stress --vm 1 --vm-bytes 128M
                ;;
            2) 
                docker run -d --name "ino_$(generate_container_name)"  --memory="128M" --cpus="0.2"  containerstack/alpine-stress stress --io 1
                ;;
            3) 
                docker run -d --name "disk_$(generate_container_name)" --memory="128M" --cpus="0.2" containerstack/alpine-stress stress --hdd 1 --hdd-bytes 128M
                ;;
        esac
    done
}

# Cronjob que ejecuta el script cada 30 segundos
while true; do
    create_random_containers
    sleep 30
done