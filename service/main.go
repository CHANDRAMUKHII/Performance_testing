package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"patient-service/controller"
	"patient-service/model"
	"patient-service/pb"

	"google.golang.org/grpc"
)

func main() {
	// HTTP Server
	go func() {
		http.HandleFunc("/details", controller.HandleBulkRequest)
		fmt.Print("Listening on port 8002")
		log.Fatal(http.ListenAndServe(":8002", nil))
	}()

	// gRPC Server
	dbClient, err := model.Connection()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer model.DisconnectDB(dbClient)

	mongoDBModel := &model.MongoDBModel{Client: dbClient}
	server := &controller.Server{Model: mongoDBModel}

	fmt.Print("Server listening on port 5002")

	lis, err := net.Listen("tcp", ":5002")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterMongoDBServiceServer(s, server)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
