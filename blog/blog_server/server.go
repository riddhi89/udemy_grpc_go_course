package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"grpc_course/blog/blogpb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var collection *mongo.Collection

type server struct {
	blogpb.UnimplementedBlogServiceServer
}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	fmt.Println("Create blog request")
	blog := req.GetBlog()
	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Title:    blog.GetTitle(),
		Content:  blog.GetContent(),
	}

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal error: %v", err))
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot convert to OID: %v", err))
	}
	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
		},
	}, nil
}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	fmt.Println("Read blog request")
	blogId := req.GetBlogId().Id
	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse ID"))
	}
	// create an empty struct
	data := &blogItem{}

	filter := bson.M{"_id": oid}
	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Cannot find blog with specified ID. %v", err))
	}

	return &blogpb.ReadBlogResponse{
		Blog: &blogpb.Blog{
			Id:       data.ID.Hex(),
			AuthorId: data.AuthorID,
			Content:  data.Content,
			Title:    data.Title,
		},
	}, nil
}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	fmt.Println("Update blog request")
	blogId := req.GetBlog().Id
	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse ID"))
	}
	// create an empty struct
	data := &blogItem{}
	filter := bson.M{"_id": oid}
	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Cannot find blog with specified ID. %v", err))
	}
	data.AuthorID = req.GetBlog().AuthorId
	data.Content = req.GetBlog().Content
	data.Title = req.GetBlog().Title

	_, err = collection.ReplaceOne(context.Background(), filter, data)

	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot update object in mongodb. %v", err))
	}
	return &blogpb.UpdateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       data.ID.Hex(),
			AuthorId: data.AuthorID,
			Content:  data.Content,
			Title:    data.Title,
		},
	}, nil
}

type blogItem struct {
	ID       primitive.ObjectID `bson::"_id, omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:content`
	Title    string             `bson:title `
}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	fmt.Println("Delete blog request")
	blogId := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, fmt.Sprintf("Cannot parse ID"))
	}
	filter := bson.M{"_id": oid}

	res, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Cannot delete: %v", err))
	}
	if res.DeletedCount == 0 {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Cannot find blog with specified ID. %v", err))
	}
	return &blogpb.DeleteBlogResponse{BlogId: blogId}, nil

}

func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	fmt.Println("List blog request")
	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknown internal error %v", err))
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		data := &blogItem{}
		err := cur.Decode(data)
		if err != nil {
			return status.Errorf(codes.Internal, fmt.Sprintf("Error decoding data from mongodb %v", err))
		}
		stream.Send(&blogpb.ListBlogResponse{Blog: &blogpb.Blog{
			Id:       data.ID.Hex(),
			AuthorId: data.AuthorID,
			Content:  data.Content,
			Title:    data.Title,
		}})
	}
	if err := cur.Err(); err != nil {
		return status.Errorf(codes.Internal, fmt.Sprintf("Unknown internal error %v", err))
	}
	return nil
}

func main() {
	// If go code crashes, it logs the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	fmt.Println("Blog Service Server")

	fmt.Println("Connecting to mongodb")
	clientOpts := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Fatal(err)
	}
	collection = client.Database("mydb").Collection("blog")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	blogpb.RegisterBlogServiceServer(s, &server{})
	// register server with reflection
	reflection.Register(s)

	go func() {
		fmt.Println("Starting Server....")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for control c to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	// Block until signal is received
	<-ch
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Closing the listener")
	lis.Close()
	fmt.Println("Closing mongo connection")
	client.Disconnect(context.TODO())
	fmt.Println("Closing the program")
}
