package main

import (
	"fmt"
	"grpc_course/prime_number_decomposition/prime_calculator_pb"
	"log"
	"net"

	"google.golang.org/grpc"
)

type server struct {
	prime_calculator_pb.UnimplementedPrimeDecomposerServiceServer
}

func (*server) PrimeDecomposer(req *prime_calculator_pb.PrimeDecomposerRequest, stream prime_calculator_pb.PrimeDecomposerService_PrimeDecomposerServer) error {
	fmt.Println("Prime decomposer was invoked")
	N := req.GetNumber()
	var k int32 = 2
	for N > 1 {
		if N%k == 0 { // if k evenly divides into N
			res := &prime_calculator_pb.PrimeDecomposerResponse{PrimeNumber: k}
			stream.Send(res) // this is a factor
			N = N / k
		} else {
			k = k + 1
		}

	}
	return nil
}

func main() {
	fmt.Println("Hello I am the server")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	prime_calculator_pb.RegisterPrimeDecomposerServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
