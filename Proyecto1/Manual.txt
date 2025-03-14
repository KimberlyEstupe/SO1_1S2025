# Contenedores Docker
### Iniciar Docker
```bash
sudo systemctl start docker
```

### Iniciar todos contenedores
```bash
sudo docker start $(sudo docker ps -a -q)
```

### Contenedores de stress
```bash
# stress de ram, genera carga en el subsistema de memoria(RAM) utilizando 128 megabytes.
docker run -d --name "ram_$(generate_container_name)" containerstack/alpine-stress stress --vm 1 --vm-bytes 128M

# CPU: el contenedor estará ejecutando una carga de trabajo que utiliza un núcleo de CPU
docker run -d --name "cpu_$(generate_container_name)" containerstack/alpine-stress stress --cpu 1

#I/O: genera carga en el subsistema de I/O (Input/Output) con un solo proceso
ocker run -d --name "ino_$(generate_container_name)" containerstack/alpine-stress stress --io 1

# Disco, el contenedor estará realizando operaciones de escritura y lectura en disco, utilizando 128 de datos. Generando carga en el Disco
docker run -d --name "disk_$(generate_container_name)" containerstack/alpine-stress stress --hdd 1 --hdd-bytes 128M
```

### Detener todos los contenedores en ejecución
```bash
docker stop $(docker ps -aq)
```

### Eliminar todos los contenedores 
```bash
docker rm $(docker ps -aq)
```

# Cronjob
### Modificacion de Cronjob
Se abre el cronjob con el comando
```bash
crontab -e
```  

Para poder ejecutar el script coloca el siguiente codigo en cronjob  
```bash
* * * * * /home/kimberly/Documentos/Sis_Operativos/Lab_Tareas/SO1_1S2025/Proyecto1/Contenedores/script.sh
* * * * * sleep 30; /home/kimberly/Documentos/Sis_Operativos/Lab_Tareas/SO1_1S2025/Proyecto1/Contenedores/script.sh
```

Luego se guarda el documento con "ctrl +x "  
Se acepta guardar la informacion y se presiona enter

# Ejecutar el Modulo
Primero realizamo el comando para compilar el modulo  
```bash
make
``` 


### Cargar Modulo
```bash
sudo insmod sysinfo_201513656.ko
``` 

### Verificar que el modulo se cargo correctamente
```bash
lsmod | grep sysinfo_201513656
``` 

### Acceder al archivo en /proc
Ir ala carpeta proc dentro de la carpeta Modulos
```bash
cat /proc/sysinfo_201513656
``` 


### Eliminar Modulo
```bash
sudo rmmod sysinfo_201513656
```  

# Servicio en Rust
crear programa
```bash
cargo new service_rust
```

en cargo.toml copia en las dependencias 
```bash
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
```

Para ejecutar el servicio, desde el directorio src usar:
```bash
cargo run
```