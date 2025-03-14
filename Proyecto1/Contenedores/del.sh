#!/bin/bash
# Detener todos los contenedores en ejecución
docker stop $(docker ps -aq)

# Eliminar todos los contenedores
docker rm $(docker ps -aq)