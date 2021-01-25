package main

import (
	"context"
	"fmt"
	"grpc_course/sum/sumpb"
	"log"
	"net"

	"google.golang.org/grpc"
)

type server struct {
	sumpb.UnimplementedSumServiceServer
}

func (*server) Sum(ctx context.Context, req *sumpb.SumRequest) (*sumpb.SumResponse, error) {
	fmt.Printf("Sum function was invoked with %v", req)
	result := req.GetNumbers().Number1 + req.GetNumbers().Number2
	res := sumpb.SumResponse{
		Result: result,
	}
	return &res, nil
}

func main() {
	fmt.Println("Hello I am the server")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	sumpb.RegisterSumServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
