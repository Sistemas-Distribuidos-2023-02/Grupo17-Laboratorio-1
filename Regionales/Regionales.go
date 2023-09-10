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
)

func usuariosInteresados(parametro string )int {
	param, _ := strconv.Atoi(parametro)
	interesados := float64(param) / 2 
	interesados_min := interesados - interesados*0.2
	interesados_max := interesados + interesados*0.2
	rango :=  interesados_max- interesados_min
	interesados_reales := rand.Intn(int(rango)) + int(interesados_min)
	fmt.Println(interesados_min, interesados_max)
	return interesados_reales

}


func main() {
	archivo, _ := os.Open("parametro_de_inicio.txt")
	defer archivo.Close()
	buffer := make([]byte, 1024) 
    n, _ := archivo.Read(buffer)
    contenido := buffer[:n]
    contenido = []byte(contenido)
    numero := string(contenido)
	interesados := usuariosInteresados(numero)
	fmt.Println(interesados)
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        log.Fatalf("No se pudo conectar al servidor: %v", err)
    }
    defer conn.Close()

    client := pb.NewValveClient(conn)
    stream, err := client.NotifyBidirectional(context.Background())
    if err != nil {
        log.Fatalf("Error al abrir el flujo bidireccional: %v", err)
    }

	// Goroutine para recibir mensajes del servidor
	go func() {
		for {
			respuesta, err := stream.Recv()
			if err != nil {
				log.Fatalf("Error al recibir mensaje del servidor: %v", err)
				break
			}
			log.Printf("Respuesta del servidor: %d", respuesta.Reply)
		}
	}()
	select {}
}