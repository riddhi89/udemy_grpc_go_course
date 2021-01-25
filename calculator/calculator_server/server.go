package main

import (
	"context"
	"fmt"
	"grpc_course/calculator/calculatorpb"
	"io"
	"log"
	"math"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	calculatorpb.UnimplementedCalculatorServiceServer
}

func (*server) ComputeAverage(stream calculatorpb.CalculatorService_ComputeAverageServer) error {
	fmt.Println("ComputeAverage rpc invoked with a streaming request on the server side")
	var runningTotal, count float64
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// we have finished reading client stream
			// return stream.SendAndClose(&greetpb.LongGreetResponse{Result: result})
			average := runningTotal / count
			return stream.SendAndClose(&calculatorpb.ComputeAverageResponse{Average: average})

		}
		if err != nil {
			log.Fatalf("Error reading client stream")
		}
		count += 1
		runningTotal += float64(req.GetInputNum())
	}

}

func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {
	fmt.Println("FindMaximum rpc invoked with a streaming request")
	var currentMax int32
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// we have finished reading client stream
			return nil
		}
		if err != nil {
			log.Fatalf("Error reading client stream: %v", err)
			return err

		}
		num := req.GetNum()
		if num > currentMax {
			currentMax = num
			sendErr := stream.Send(&calculatorpb.FindMaximumResponse{MaxNum: currentMax})
			if sendErr != nil {
				log.Fatalf("Error sending data to client: %v", sendErr)
				return sendErr
			}

		}

	}
}

func (*server) SqaureRoot(ctx context.Context, req *calculatorpb.SqaureRootRequest) (*calculatorpb.SqaureRootResponse, error) {
	fmt.Println("Received sqaureroot RPC")
	number := req.GetNumber()
	if number < 0 {
		return nil, status.Errorf(codes.InvalidArgument,
			fmt.Sprintf("Received a negative number: %v", number))
	}
	return &calculatorpb.SqaureRootResponse{
		NumberRoot: math.Sqrt(float64(number)),
	}, nil
}

func main() {
	fmt.Println("Hello I am the server")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
