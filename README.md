# Grupo17-Laboratorio-1

## Integrantes
* Vicente Muñoz Rojas - 202073557-3
* Carlos Lagos - 202073571-9
* Carlos Kuhn - 202073574-3

## Instrucciones 
Para iniciar el servidor central usando docker usamos.
```
make docker-central
```
Para iniciar un servidor regional usando docker usamos el siguiente comando. 
```
make docker-regional
```
Ademas tienes que verificar la configuracipon en docker-compose.yml

Para iniciar un servidor rabbit usando docker usamos.
```
make docker-rabbit
docker exec -it rabbitmq /bin/bash
rabbitmqadmin declare queue name=solicitudes_regionales
```
## Consideraciones
* Se debe iniciar la cola rabbit primero, luego la central y posteriormente iniciar en algun orden los servidores regionales. (puede ocurrir errores si son todos al mismo tiempo).
* Para iniciar el contenedor de los servidores regionales, en el Makefile dejar solo las regiones las cuales quieres iniciar en esa maquina. Ademas tienes que editar los datos necesarios en el docker compose, como por ejemplo la cola rabbit o las ip del servidor central.
* El archivo docker-compose-servers.yml es un archivo docker-compose.yml de ejemplo pero configurado con las maquinas virtuales de prueba.

### Ejemplo
Makefile para la region 3:
```makefile
docker-regional:
	docker-compose -f docker-compose.yml up regionales3
```
Docker-compose.yml para region 3 (Puedes cambiar los valores en args):
```yml
regionales3:
    build:
      context: ./Regionales 
      dockerfile: Dockerfile.regionales
      args:
        valve_server: dist066.inf.santiago.usm.cl
        rmq_server: dist065.inf.santiago.usm.cl
        rmq_port: 25671
        region_name: Asia
    volumes:
      - ./Regionales:/app/Regionales
    network_mode: "host"
```
