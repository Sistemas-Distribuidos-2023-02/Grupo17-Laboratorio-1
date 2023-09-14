package main

import (
	pb "Valve/proto"
	"context"
	"log"
	"google.golang.org/grpc"
	"math/rand"
	"strconv"
	"os"
	"fmt"
	"time"
	"strings"
	amqp "github.com/rabbitmq/amqp091-go"
)



func usuariosInteresados(parametro string )int {
	rand.Seed(time.Now().UnixNano())
	param, _ := strconv.Atoi(parametro)
	interesados := float64(param) / 2 
	interesados_min := interesados - interesados*0.2
	interesados_max := interesados + interesados*0.2
	rango :=  interesados_max- interesados_min
	interesados_reales := rand.Intn(int(rango)) + int(interesados_min)
	return interesados_reales

}

func Cola_Rabbit(interesados int){
	    // Establece una conexión con RabbitMQ
		conn, err := amqp.Dial("amqp://guest:guest@" + os.Getenv("rmq_server") + ":" + os.Getenv("rmq_port") + "/")
		if err != nil {
			log.Fatalf("Error al conectar a RabbitMQ: %v", err)
		}
		defer conn.Close()
	
		// Crea un canal de comunicación
		ch, err := conn.Channel()
		if err != nil {
			log.Fatalf("Error al abrir un canal: %v", err)
		}
		defer ch.Close()
	
		// Declara la cola a la que quieres enviar el mensaje
		queueName := "solicitudes_regionales"
	
		// Publica un mensaje en la cola
		mensaje := strconv.Itoa(interesados) + " - " + os.Getenv("region_name")
		//mensaje := strconv.Itoa(interesados) + " - " + "America"
		err = ch.Publish(
			"",         // Intercambio (en blanco para la cola predeterminada)
			queueName,  // Nombre de la cola
			false,      // Mandatorio
			false,      // Inmediato
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(mensaje),
			})
		if err != nil {
			log.Fatalf("Error al publicar mensaje en la cola: %v", err)
		}
}




func main() {
    archivo, _ := os.Open("parametro_de_inicio.txt")
    defer archivo.Close()
    buffer := make([]byte, 1024)
    n, _ := archivo.Read(buffer)
    contenido := buffer[:n]
    contenido = []byte(contenido)
    numero := strings.TrimSpace(string(contenido))
    interesados := usuariosInteresados(numero)

    fmt.Println("Cantidad de personas interesadas: ",interesados, "en la region: ", os.Getenv("region_name"))
    conn, err := grpc.Dial(os.Getenv("valve_server") + ":50051", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("No se pudo conectar al servidor: %v", err)
    }
    defer conn.Close()

    client := pb.NewValveClient(conn)
    stream, err := client.NotifyBidirectional(context.Background())
    if err != nil {
        log.Fatalf("Error al abrir el flujo bidireccional: %v", err)
    }

    // Recibir un único mensaje del servidor al inicio de la conexión
	for {
		respuesta, err := stream.Recv()
		if err != nil {
			log.Fatalf("Error al recibir mensaje del servidor: %v", err)
		}
		log.Printf("Respuesta del servidor: %d", respuesta.Reply)



		Cola_Rabbit(interesados)

		respuesta2, err := stream.Recv()
		if err != nil {
			log.Fatalf("Error al recibir mensaje del servidor: %v", err)
		}
		log.Printf("Respuesta del servidor: %d", respuesta2.Reply)
		fmt.Println("Personas que lograron inscribirse: ", interesados-int(respuesta2.Reply))
		interesados = int(respuesta2.Reply)
		fmt.Println("Se quedaron sin llaves: ", interesados)


	}
    select {}
}
