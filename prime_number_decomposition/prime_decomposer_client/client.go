package main

import (
	"context"
	"fmt"
	"grpc_course/prime_number_decomposition/prime_calculator_pb"

	"io"
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

	c := prime_calculator_pb.NewPrimeDecomposerServiceClient(clientConn)

	//doUnary(c)
	doServerStreaming(c)
}

func doServerStreaming(c prime_calculator_pb.PrimeDecomposerServiceClient) {
	fmt.Println("Starting to do a server streaming RPC call...")
	var n int32 = 120
	req := &prime_calculator_pb.PrimeDecomposerRequest{
		Number: n,
	}
	resStream, err := c.PrimeDecomposer(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling server streaming prime decomposer rpc: %v", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			// server closed stream
			break
		}
		if err != nil {
			log.Fatalf("Error reading server stream : %v", err)
		}
		log.Println("Response from PrimeDecomposer: ", msg.GetPrimeNumber())
	}

}
