package main

import (
	"context"
	"fmt"
	"grpc_course/greet/greetpb"
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

	c := greetpb.NewGreetServiceClient(clientConn)

	//doUnary(c)
	//doServerStreaming(c)
	//doClientStreaming(c)
	// doBiDiStreaming(c)
	doUnaryWithDeadline(c, 5*time.Second)
	doUnaryWithDeadline(c, 1*time.Second)

}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Unary RPC starting...")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Riddhi",
			LastName:  "Shah",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling greet rpc: %v", err)
	}
	log.Println("Response from GreetService ", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a server streaming RPC call...")
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Riddhi",
			LastName:  "Shah",
		},
	}
	resStream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling server streaming greet rpc: %v", err)
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
		log.Println("Response from GreetManyTimesRequest: ", msg.GetResult())
	}

}

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a client streaming RPC...")

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Error while calling long greet client streaming rpc: %v", err)

	}
	requests := []*greetpb.LongGreetRequest{

		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Riddhi",
				LastName:  "Shah",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Arya",
				LastName:  "Gode",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Tejaswi",
				LastName:  "Gode",
			},
		},
	}

	for _, req := range requests {
		fmt.Println("Sending req: ", req)
		stream.Send(req)
		time.Sleep(100 * time.Millisecond)
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving response from LongGreet %v", err)
	}
	fmt.Println("LongGreet Response", res)

}

func doBiDiStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Starting to do a bi-di streaming...")

	// create a stream
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
		return
	}

	requests := []*greetpb.GreetEveryoneRequest{

		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Riddhi",
				LastName:  "Shah",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Arya",
				LastName:  "Gode",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Tejaswi",
				LastName:  "Gode",
			},
		},
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
			log.Println("Received: ", res.GetResult())

		}
		close(waitc)

	}()

	// wait until everything is done
	<-waitc

}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, timeout time.Duration) {
	fmt.Println("Unary RPC with deadline starting...")
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Riddhi",
			LastName:  "Shah",
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	res, err := c.GreetWithDeadline(ctx, req)
	if err != nil {

		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout was hit! Deadline was exceeded")
				return
			} else {
				fmt.Println("unexpected error", statusErr)
			}

		} else {
			log.Fatalf("Error while calling greet with deadline rpc: %v", err)
		}
		return

	}
	log.Println("Response from GreetWithDeadline ", res.Result)
}
