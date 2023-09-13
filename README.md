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
## Orden de ejecución
Iniciar la cola rabbit primero, luego la central y posteriormente iniciar en algun orden los servidores regionales. (puede ocurrir errores si son todos al mismo tiempo).
