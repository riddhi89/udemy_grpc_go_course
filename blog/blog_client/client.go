package main

import (
	"context"
	"fmt"
	"grpc_course/blog/blogpb"
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

	c := blogpb.NewBlogServiceClient(clientConn)

	fmt.Println("Creating the blog")
	createBlogReq := &blogpb.CreateBlogRequest{
		Blog: &blogpb.Blog{
			AuthorId: "Riddhi",
			Title:    "My first blog",
			Content:  "Contents of my first blog",
		},
	}
	res, err := c.CreateBlog(context.Background(), createBlogReq)
	if err != nil {
		log.Fatalf("Unexpected error: %v", err)
	}
	fmt.Println("Blog has been created: ", res)
	blogId := res.Blog.GetId()

	// read blog
	_, err = c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: &blogpb.Blog{Id: "blah"}})
	if err != nil {
		fmt.Println("Error while reading the blog", err)
	}

	readResp, err := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: &blogpb.Blog{Id: blogId}})
	if err != nil {
		fmt.Println("Error while reading the blog", err)
	}
	fmt.Println("Read the blog", readResp)

	// update blog
	newBlog := &blogpb.Blog{
		Id:       blogId,
		AuthorId: "Changed author",
		Title:    "edited title",
		Content:  "edited content",
	}
	updateRes, err := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{Blog: newBlog})

	if err != nil {
		fmt.Println("Error while updating the blog", err)
	}
	fmt.Println("Updated the blog", updateRes)

	// Delete blog
	deleteRes, err := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: blogId})
	if err != nil {
		fmt.Println("Error deleting the blog", err)
	}
	fmt.Println("Blog was deleted", deleteRes)

	// List Blogs
	resStream, err := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})
	if err != nil {
		log.Fatalf("Error while calling server streaming list blogs rpc: %v", err)
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
		fmt.Println(msg.GetBlog())
	}
}
