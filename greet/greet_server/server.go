package main

import (
	"context"
	"fmt"
	"grpc_course/greet/greetpb"
	"io"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
	greetpb.UnimplementedGreetServiceServer
}

func (*server) Greet(ctx context.Context, req *greetpb.GreetRequest) (*greetpb.GreetResponse, error) {
	firstName := req.GetGreeting().GetFirstName()
	result := "Hello " + firstName
	res := greetpb.GreetResponse{
		Result: result,
	}
	return &res, nil
}

func (*server) GreetManyTimes(req *greetpb.GreetManyTimesRequest, stream greetpb.GreetService_GreetManyTimesServer) error {
	fmt.Println("GreetManyTimes was invoked")
	firstName := req.GetGreeting().GetFirstName()
	for i := 0; i < 10; i++ {
		result := "Hello " + firstName + " number" + strconv.Itoa(i)
		res := &greetpb.GreetManyTimesResponse{
			Result: result,
		}
		stream.Send(res)
		time.Sleep(1000 * time.Millisecond)
	}
	return nil
}

func (*server) LongGreet(stream greetpb.GreetService_LongGreetServer) error {
	fmt.Println("LongGreet rpc invoked with a streaming request")
	result := ""
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// we have finished reading client stream
			return stream.SendAndClose(&greetpb.LongGreetResponse{Result: result})

		}
		if err != nil {
			log.Fatalf("Error reading client stream")
			return err
		}
		firstName := req.GetGreeting().GetFirstName()
		result += "Hello " + firstName + "! "
	}

}

func (*server) GreetEveryone(stream greetpb.GreetService_GreetEveryoneServer) error {
	fmt.Println("GreetEveryone rpc invoked with a streaming request")
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
		firstName := req.GetGreeting().GetFirstName()
		result := "Hello " + firstName + "! "
		sendErr := stream.Send(&greetpb.GreetEveryoneResponse{Result: result})
		if sendErr != nil {
			log.Fatalf("Error sending data to client: %v", sendErr)
			return sendErr
		}
	}
}

func (*server) GreetWithDeadline(ctx context.Context, req *greetpb.GreetWithDeadlineRequest) (*greetpb.GreetWithDeadlineResponse, error) {
	for i := 0; i < 3; i++ {
		if ctx.Err() == context.Canceled {
			// the client cancelled the request
			fmt.Println("The client cancelled the request")
			return nil, status.Error(codes.Canceled, "Client cancelled the request")
		}
		time.Sleep(1 * time.Second)
	}
	firstName := req.GetGreeting().GetFirstName()
	result := "Hello " + firstName
	res := greetpb.GreetWithDeadlineResponse{
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
	greetpb.RegisterGreetServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

}
