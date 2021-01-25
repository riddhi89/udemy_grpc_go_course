package main

import (
	"context"
	"fmt"
	"grpc_course/sum/sumpb"
	"log"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Hello I am the client")
	clientConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer clientConn.Close()

	c := sumpb.NewSumServiceClient(clientConn)

	doUnary(c)
}

func doUnary(c sumpb.SumServiceClient) {
	fmt.Println("Unary RPC starting...")
	req := &sumpb.SumRequest{
		Numbers: &sumpb.Numbers{
			Number1: 5,
			Number2: 3,
		},
	}
	res, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling sum rpc: %v", err)
	}
	log.Println("Response from SumService ", res.Result)
}
