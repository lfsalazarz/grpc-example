package main

import (
	"context"
	"errors"
	"grpc-example/service"
	"io"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/golang/protobuf/ptypes"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// https://developers.google.com/maps-booking/reference/grpc-api/status_codes
// https://godoc.org/google.golang.org/grpc/status#Errorf
// https://godoc.org/github.com/golang/protobuf/ptypes#

type server struct {
	// DB *sql.DB
}

func (s *server) Unary(ctx context.Context, req *service.RequestUnary) (*service.ResponseUnary, error) {
	item := req.GetItem()
	// validate fields
	if err := validateItem(item); err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Validate error: %v\n",
			err,
		)
	}
	item.Id = uuid.NewV4().String()
	// save to database
	// ...

	return &service.ResponseUnary{
		Id: item.Id,
	}, nil
}

func (s *server) ServerStreaming(req *service.RequestServerStreaming, stream service.MyCustomService_ServerStreamingServer) error {
	ids := req.GetId()

	mockData := []*service.Item{
		&service.Item{
			Id:        "1",
			Name:      "Foo",
			Number:    110,
			Price:     float64(3.2),
			IsActive:  false,
			CreatedAt: ptypes.TimestampNow(),
		},
		&service.Item{
			Id:        "2",
			Name:      "Bar",
			Number:    220,
			Price:     float64(6.4),
			IsActive:  true,
			CreatedAt: ptypes.TimestampNow(),
		},
		&service.Item{
			Id:        "3",
			Name:      "Baz",
			Number:    330,
			Price:     float64(9.6),
			IsActive:  false,
			CreatedAt: ptypes.TimestampNow(),
		},
	}

	for _, id := range ids {
		// find items in db
		// ...
		for _, item := range mockData {
			if id == item.GetId() {
				res := &service.ResponseServerStreaming{
					Item: item,
				}
				stream.Send(res)
			}
		}
	}

	return nil
}

func (s *server) ClientStreaming(stream service.MyCustomService_ClientStreamingServer) error {

	var ids []string

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&service.ResponseClientStreaming{
				Id: ids,
			})
		}
		if err != nil {
			return status.Errorf(
				codes.Internal,
				"Error while reading client stream: %v",
				err,
			)
		}

		item := req.GetItem()
		if err := validateItem(item); err != nil {
			return status.Errorf(
				codes.InvalidArgument,
				"Validate error: %v\n",
				err,
			)
		}
		item.Id = uuid.NewV4().String()
		// save to db
		// ....
		ids = append(ids, item.Id)
	}
}

func validateItem(item *service.Item) error {
	if item.GetName() == "" {
		return errors.New("DataId must be present")
	}
	if item.GetPrice() < 0.0 {
		return errors.New("Price must be positive")
	}

	return nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// SSL
	certFile := "ssl/server/server-cert.pem"
	keyFile := "ssl/server/server-key.pem"
	creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
	if sslErr != nil {
		log.Fatalf("Failed loading certificates: %v\n", sslErr)
	}

	// gRPC server options
	opts := grpc.Creds(creds)

	// gRPC Server
	s := grpc.NewServer(opts)
	// Register MyCustomService service
	service.RegisterMyCustomServiceServer(s, &server{})
	// Register reflection service on gRPC server
	reflection.Register(s)

	// Start server
	go func() {
		log.Println("Starting server...")
		if err := s.Serve(listener); err != nil {
			log.Fatalf("Failed to serve: %v\n", err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, os.Kill)
	// Block until a signal is received
	sig := <-ch
	log.Println("Got signal: ", sig)
	log.Println("Stopping the server")
	s.Stop()
	listener.Close()
}
