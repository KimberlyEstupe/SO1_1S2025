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
# stress de ram, genera carga en el subsistema de memoria(RAM) utilizando 512 megabytes.
docker run -d --name "$nombre" containerstack/alpine-stress -m 512M

# CPU: el contenedor estará ejecutando una carga de trabajo que utiliza un núcleo de CPU
docker run -d --name "$nombre" containerstack/alpine-stress -c 1

#I/O: genera carga en el subsistema de I/O (Input/Output) con un solo proceso
docker run -d --name "$nombre" containerstack/alpine-stress -i 1

# Disco, el contenedor estará realizando operaciones de escritura y lectura en disco, utilizando 512M de datos. 
# Generando carga en el Disco
docker run -d --name "$nombre" containerstack/alpine-stress -d 512M
```


### Eliminar todos los contenedores 
```bash
docker rm $(docker ps -aq)

# Ver imagenes construidas por medio de docker build
docker images
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
/proc
cat /proc/sysinfo_201513656
``` 


### Eliminar Modulo
```bash
sudo rmmod sysinfo_201513656
```