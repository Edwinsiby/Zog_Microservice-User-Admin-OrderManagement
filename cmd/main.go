package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	pb "service4/pb"
	"service4/pkg/service"
)

func main() {
	orderService := &service.Order{}

	grpcServer := grpc.NewServer()

	pb.RegisterOrderServer(grpcServer, orderService)

	listener, err := net.Listen("tcp", ":5053")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Println("Admin&User-Order Management-Server is running on 5053")
	go grpcServer.Serve(listener)

	select {}
}
