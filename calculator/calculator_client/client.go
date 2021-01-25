package main

import (
	"context"
	"fmt"
	"grpc_course/calculator/calculatorpb"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	fmt.Println("Hello I am the client")
	clientConn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer clientConn.Close()

	c := calculatorpb.NewCalculatorServiceClient(clientConn)

	//doClientStreaming(c)
	//doBidiStreaming(c)
	doErrorUnary(c)

}

func doClientStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a client streaming RPC...")

	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("Error while calling compute average streaming rpc: %v", err)

	}
	requests := []*calculatorpb.ComputeAverageRequest{
		{InputNum: 1},
		{InputNum: 2},
		{InputNum: 3},
	}

	for _, req := range requests {
		fmt.Println("Sending req: ", req)
		stream.Send(req)
		time.Sleep(100 * time.Millisecond)
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response from ComputeAverage %v", err)
	}
	fmt.Println("ComputeAverage Response", res)

}

func doBidiStreaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting to do a bi-di streaming...")

	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
		return
	}

	requests := []*calculatorpb.FindMaximumRequest{
		&calculatorpb.FindMaximumRequest{Num: 1},
		&calculatorpb.FindMaximumRequest{Num: 5},
		&calculatorpb.FindMaximumRequest{Num: 3},
		&calculatorpb.FindMaximumRequest{Num: 6},
		&calculatorpb.FindMaximumRequest{Num: 2},
		&calculatorpb.FindMaximumRequest{Num: 20},
	}
	waitc := make(chan struct{})
	go func() {
		// function to send a bunch of messages
		for _, req := range requests {
			fmt.Println("Sending req: ", req)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()
	go func() {
		// function to receive a bunch of messages
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				// server closed stream
				break
			}
			if err != nil {
				log.Fatalf("Error reading server stream : %v", err)
			}
			log.Println("Received: ", res.GetMaxNum())

		}
		close(waitc)

	}()

	// wait until everything is done
	<-waitc

}

func doErrorUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Println("Starting sqaure root unary RPC...")
	//correct
	doErrorCall(c, 10)
	//error call
	doErrorCall(c, -2)

}

func doErrorCall(c calculatorpb.CalculatorServiceClient, n int32) {
	res, err := c.SqaureRoot(context.Background(), &calculatorpb.SqaureRootRequest{Number: n})

	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			// actual error from gRPC (user error)
			fmt.Errorf(respErr.Message())
			fmt.Println(respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("We probably sent a negative number")
			}
			return
		} else {
			log.Fatal("Big error calling sqaureroot: %v", err)
		}
	}
	fmt.Println("Result", res.GetNumberRoot())

}
