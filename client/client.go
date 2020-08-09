package main

import (
	"context"
	"fmt"
	"grpc-example/service"
	"io"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

func main() {
	// Certificate Authority - Trust certificate
	certFile := "ssl/ca/ca-cert.pem"
	creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
	if sslErr != nil {
		log.Fatalf("Error while loading CA trust certificate: %v \n", sslErr)
		return
	}
	opts := grpc.WithTransportCredentials(creds)
	// opts := grpc.WithInsecure()
	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("could not connet: %v", err)
	}
	defer cc.Close()

	client := service.NewMyCustomServiceClient(cc)

	unaryRequest(client)
	serverStreamRequest(client)
	clientStreamRequest(client)
}

func unaryRequest(client service.MyCustomServiceClient) {
	req := &service.RequestUnary{
		Item: &service.Item{
			Name:     "Foo",
			Number:   300,
			Price:    17.5,
			IsActive: true,
		},
	}
	res, err := client.Unary(context.Background(), req)
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			// actual error from gRPC (user error)
			fmt.Println(respErr.Message())
			fmt.Println(respErr.Code())
		} else {
			log.Fatalf("error while calling Unary %v\n", err)
			return
		}
	}
	log.Printf("Response from Unary: %v\n", res.GetId())
}

func serverStreamRequest(client service.MyCustomServiceClient) {
	req := &service.RequestServerStreaming{
		Id: []string{"1", "3"},
	}
	stream, err := client.ServerStreaming(context.Background(), req)
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			fmt.Println(respErr.Message())
			fmt.Println(respErr.Code())
		} else {
			log.Fatalf("%v\n", err)
			return
		}
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			// end of the stream
			break
		}
		if err != nil {
			log.Fatalf("error while reading stream: %v", err)
		}
		log.Printf("Response from ServerStreaming %v\n", res.GetItem())
	}
}

func clientStreamRequest(client service.MyCustomServiceClient) {
	stream, err := client.ClientStreaming(context.Background())
	if err != nil {
		log.Fatalf("error while calling ClientStreaming %v\n", err)
	}

	requests := []*service.Item{
		&service.Item{
			Name:     "Foo",
			Number:   110,
			Price:    float64(3.2),
			IsActive: false,
		},
		&service.Item{
			Name:     "Bar",
			Number:   220,
			Price:    float64(6.4),
			IsActive: true,
		},
		&service.Item{
			Name:     "Baz",
			Number:   330,
			Price:    float64(9.6),
			IsActive: false,
		},
	}

	for _, req := range requests {
		stream.Send(&service.RequestClientStreaming{
			Item: req,
		})
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			fmt.Println(respErr.Message())
			fmt.Println(respErr.Code())
		} else {
			log.Fatalf("error while receiving response from ClientStreaming %v\n", err)
			return
		}
	}

	log.Printf("Response from ClientStreaming: %v\n", res.GetId())

}
