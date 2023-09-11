package main

import (
	"fmt"
	"os"
	"time"
	"strings"
	"math/rand"
	"strconv"
	_"context"
	"log"
	"google.golang.org/grpc"
	"net"
	"sync"
	pb "Valve/proto"
    amqp "github.com/rabbitmq/amqp091-go"
)

type server struct {
	pb.UnimplementedValveServer
}

var llaves int

//generarLlaves genera llaves al azar.
func generarLlaves(min, max string) int {
	rand.Seed(time.Now().UnixNano())
	minInt, _ := strconv.Atoi(min)
	maxInt, _ := strconv.Atoi(max)
	rango := maxInt - minInt
	llaves = rand.Intn(rango) + minInt
	return llaves
}

func (s *server) NotifyBidirectional(stream pb.Valve_NotifyBidirectionalServer) error {
    var wg sync.WaitGroup

    // Función para enviar mensajes al cliente
    sendToClient := func() {
        response := &pb.Response{Reply: int64(llaves)}
        if err := stream.Send(response); err != nil {
            log.Printf("Error al enviar respuesta al cliente: %v", err)
        }
    }

    // Enviar la cantidad de llaves al cliente una vez al inicio de la conexión
    sendToClient()

    // Goroutine para enviar mensajes al cliente periódicamente (opcional)
    // Puedes comentar o eliminar esta sección si no deseas enviar actualizaciones periódicas
    wg.Add(1)
    go func() {
        defer wg.Done()
        ticker := time.NewTicker(5 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                sendToClient()
            case <-stream.Context().Done():
                return
            }
        }
    }()

    for {
        req, err := stream.Recv()
        if err != nil {
            break
        }

        log.Printf("Mensaje recibido del cliente: %s", req.Message)
    }

    wg.Wait()
    return nil
}



func Datos_Cola_Rabbit(llaves int, archivo *os.File){
        // Establece una conexión con RabbitMQ
        fmt.Println(os.Getenv("rmq_server") + ":5672/")
        conn, err := amqp.Dial("amqp://guest:guest@" + os.Getenv("rmq_server") + ":5672/")
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
    
        // Nombre de la cola a la que deseas suscribirte
        queueName := "solicitudes_regionales"
    

        // Inicia la suscripción a la cola
        msgs, err := ch.Consume(
            queueName, // Nombre de la cola
            "",        // Etiqueta del consumidor (en blanco para una etiqueta generada)
            true,      // Auto-acknowledgment
            false,     // Exclusividad
            false,     // No local
            false,     // No espera confirmaciones
            nil,       // Argumentos adicionales
        )
        if err != nil {
            log.Fatalf("Error al registrar el consumidor de la cola: %v", err)
        }
        
        // Procesa los mensajes recibidos desde la cola
        for msg := range msgs {
            mensaje := string(msg.Body)
            log.Printf("Mensaje recibido de la cola: %s", mensaje)
            value_mensaje := strings.Split(mensaje, " - ")
            mensaje_numero,_ := strconv.Atoi(value_mensaje[0])
            if llaves - mensaje_numero < 0{
                llaves_solicitadas := mensaje_numero
                mensaje_numero = mensaje_numero - llaves
                usuarios_registrados := llaves
                llaves = 0
                fmt.Println("No quedan llaves a las",time.Now().Format("15:04:05"))
                fmt.Println("Quedaron",mensaje_numero,"personas sin llaves en ", value_mensaje[1])
                linea := "  "+value_mensaje[1] + "-" + strconv.Itoa(llaves_solicitadas) + "-" + strconv.Itoa(usuarios_registrados) + "-"+ strconv.Itoa(mensaje_numero) + "\n"
                archivo.WriteString(linea)
            }else{
                llaves = llaves - mensaje_numero
                linea := "  "+value_mensaje[1] + "-" + strconv.Itoa(mensaje_numero) + "-" + strconv.Itoa(mensaje_numero) + "- 0\n"
                archivo.WriteString(linea)
                fmt.Println("Quedan",llaves,"llaves a las",time.Now().Format("15:04:05"))
            }
            serv := grpc.NewServer()
            pb.RegisterValveServer(serv, &server{})
 
}}

// Se supone que esta es la central Valve.
func main() {
	archivo, _ := os.Open("parametro_de_inicio.txt")
	
	
	buffer := make([]byte, 1024) 
    n, _ := archivo.Read(buffer)
    contenido := buffer[:n]

    contenido = []byte(contenido)
    rangoStr := string(contenido)
    archivo.Close()
    rango := strings.Split(rangoStr, "-")
    
    min := strings.TrimSpace(rango[0])
    max_it := strings.Split(rango[1], "\n")
    max := max_it[0]
    iteraciones := max_it[1]

    max = strings.TrimSpace(max)
    fmt.Println("min:", min)
    fmt.Println("max:", max)
    fmt.Println("iteraciones:", iteraciones)
    


    fmt.Println("Se generaran llaves al azar pertenecientes al siguente rango: [",min,",",max,"]")
	llaves = generarLlaves(min,max)

	//Print si tiene que ir.
	fmt.Println("Se generaron ",llaves," llaves a las",time.Now().Format("15:04:05"))

    archivo, err := os.Create("Registros.txt")
    if err != nil {
        fmt.Println("Error al crear el archivo:", err)
        return
    }
    defer archivo.Close()
    archivo.WriteString(time.Now().Format("15:04") + " - " + strconv.Itoa(llaves) + "\n")


    go Datos_Cola_Rabbit(llaves , archivo)

	//Coneccion con el servidor.
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	serv := grpc.NewServer()

	pb.RegisterValveServer(serv, &server{})
    
    
	log.Printf("server listening at %v", listener.Addr())
	if err := serv.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}



}