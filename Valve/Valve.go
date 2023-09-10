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

    // Goroutine para enviar mensajes al cliente periódicamente
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


// Se supone que esta es la central Valve.
func main() {
	archivo, _ := os.Open("parametro_de_inicio.txt")
	defer archivo.Close()
	
	buffer := make([]byte, 1024) 
    n, _ := archivo.Read(buffer)
    contenido := buffer[:n]

    contenido = []byte(contenido)
    rangoStr := string(contenido)

    rango := strings.Split(rangoStr, "-")

    min := rango[0]
    max := rango[1]
	//Este print no se si tiene que ir.
    fmt.Println("Se generaran llaves al azar pertenecientes al siguente rango: [",min,",",max,"]")
	llaves = generarLlaves(min,max)

	//Print si tiene que ir.
	fmt.Println("Se generaron ",llaves," llaves a las",time.Now().Format("15:04:05 del 2006-01-02"))


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