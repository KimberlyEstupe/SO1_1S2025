#!/bin/bash

# Función para generar un nombre de contenedor único
generar_nombre_contenedor() {
    # Usar /dev/urandom para generar un nombre aleatorio
    echo "contenedor_$(od -An -N2 -i /dev/urandom | tr -d ' ')"
}

# Array de tipos de contenedores
tipos=("ram" "cpu" "io" "disco")

# Crear 10 contenedores aleatorios
for i in {1..10}; do
    tipo=${tipos[$RANDOM % ${#tipos[@]}]}  # Seleccionar un tipo aleatorio
    nombre=$(generar_nombre_contenedor)    # Generar un nombre único

    # Crear el contenedor según el tipo
    case $tipo in
        ram)
            docker run -d --name "$nombre" containerstack/alpine-stress -m 512M
            ;;
        cpu)
            docker run -d --name "$nombre" containerstack/alpine-stress -c 1
            ;;
        io)
            docker run -d --name "$nombre" containerstack/alpine-stress -i 1
            ;;
        disco)
            docker run -d --name "$nombre" containerstack/alpine-stress -d 1G
            ;;
    esac
done


