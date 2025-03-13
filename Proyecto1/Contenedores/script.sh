#!/bin/bash

# Función para generar un nombre de contenedor único

declare -a tipos=("ram" "cpu" "io" "disk")

# Elegir un tipo de contenedor aleatorio
tipo=${tipos[$RANDOM % ${#tipos[@]}]}

# Generar un nombre único para el contenedor
nombre_contenedor="contenedor_$tipo_$(date +%s%N | cut -b1-10)_$RANDOM"


# Crear 10 contenedores aleatorios
for i in {1..10}; do
    tipo=${tipos[$RANDOM % ${#tipos[@]}]}  # Seleccionar un tipo aleatorio
    nombre=$(generar_nombre_contenedor)    # Generar un nombre único

    # Crear el contenedor según el tipo
    case $tipo in
        ram)
            docker run -d --name "$nombre" containerstack/alpine-stress -m 512M
            echo "Contenedor creado: $nombre_contenedor (Tipo: $tipo)"
            ;;
        cpu)
            docker run -d --name "$nombre" containerstack/alpine-stress -c 1
            echo "Contenedor creado: $nombre_contenedor (Tipo: $tipo)"
            ;;
        io)
            docker run -d --name "$nombre" containerstack/alpine-stress -i 1
            echo "Contenedor creado: $nombre_contenedor (Tipo: $tipo)"
            ;;
        disco)
            docker run -d --name "$nombre" containerstack/alpine-stress -d 512M
            echo "Contenedor creado: $nombre_contenedor (Tipo: $tipo)"
            ;;
    esac
done