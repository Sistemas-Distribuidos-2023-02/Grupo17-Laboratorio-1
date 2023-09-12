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


var llavesMutex sync.Mutex
var llaves int
var archivo *os.File
var iteraciones int
var min string
var max string
var mensaje_numero int
var servers int 

//generarLlaves genera llaves al azar.
func generarLlaves(min, max string) int {
	rand.Seed(time.Now().UnixNano())
	minInt, _ := strconv.Atoi(min)
	maxInt, _ := strconv.Atoi(max)
	rango := maxInt - minInt
    llavesMutex.Lock()
	llaves = rand.Intn(rango) + minInt
    llavesMutex.Unlock()
	return llaves
    
}

func (s *server) NotifyBidirectional(stream pb.Valve_NotifyBidirectionalServer) error {
    var wg sync.WaitGroup
    updateCh := make(chan int)
    var mu sync.Mutex
    for {
        mu.Lock()
        servers = servers + 1
        sendToClient := func(valor int) {
            response := &pb.Response{Reply: int64(valor)}
            if err := stream.Send(response); err != nil {
                log.Printf("Error al enviar respuesta al cliente: %v", err)
            }
        }

        // Enviar la cantidad de llaves al cliente una vez al inicio de la conexión
        sendToClient(llaves)

        // Ejecutar Datos_Cola_Rabbit en una goroutine
        wg.Add(1)
        go func() {
            defer wg.Done()
            Datos_Cola_Rabbit(archivo, updateCh)
        }()

        time.Sleep(3 * time.Second)
        updatedValue := <-updateCh
        llaves = updatedValue

        sendToClient(mensaje_numero)
        if servers == 4{
            servers = 0
            iteraciones = iteraciones - 1

            
            if iteraciones == 0{
                fmt.Println("Se termino el programa a las",time.Now().Format("15:04:05"))
                break
            }
            llaves = generarLlaves(min,max)
            archivo.WriteString(time.Now().Format("15:04") + " - " + strconv.Itoa(llaves) + "\n")
            
            fmt.Println("Se generaron ",llaves," llaves a las",time.Now().Format("15:04:05"))
        }
        mu.Unlock()
        time.Sleep(3 * time.Second)
        
    }
    return nil
}



func Datos_Cola_Rabbit( archivo *os.File, updateCh chan<- int){

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
            llavesMutex.Lock()
            mensaje := string(msg.Body)
            log.Printf("Mensaje recibido de la cola: %s", mensaje)
            value_mensaje := strings.Split(mensaje, " - ")
            mensaje_numero,_ = strconv.Atoi(value_mensaje[0])
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
                mensaje_numero = 0
                archivo.WriteString(linea)
                fmt.Println("Quedan",llaves,"llaves a las",time.Now().Format("15:04:05"))
            }
            updateCh <- llaves
            llavesMutex.Unlock()
            
            
}}



// Se supone que esta es la central Valve.
func main() {
    servers = 0
	archivo1, _ := os.Open("parametro_de_inicio.txt")
	
	
	buffer := make([]byte, 1024) 
    n, _ := archivo1.Read(buffer)
    contenido := buffer[:n]

    contenido = []byte(contenido)
    rangoStr := string(contenido)
    archivo1.Close()
    rango := strings.Split(rangoStr, "-")
    
    min = strings.TrimSpace(rango[0])
    max_it := strings.Split(rango[1], "\n")
    max = max_it[0]
    iteraciones,_ = strconv.Atoi(max_it[1])

    max = strings.TrimSpace(max)
    fmt.Println("min:", min)
    fmt.Println("max:", max)
    fmt.Println("iteraciones:", iteraciones)
    


    fmt.Println("Se generaran llaves al azar pertenecientes al siguente rango: [",min,",",max,"]")
	llaves = generarLlaves(min,max)

	//Print si tiene que ir.
	fmt.Println("Se generaron ",llaves," llaves a las",time.Now().Format("15:04:05"))

    archivo, _ = os.Create("Registros.txt")

    defer archivo.Close()
    archivo.WriteString(time.Now().Format("15:04") + " - " + strconv.Itoa(llaves) + "\n")


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