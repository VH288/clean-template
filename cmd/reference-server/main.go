package main

import (
	"context"
	"fmt"
	"log"
	"net"

	externalv1 "clean-template/proto/external/v1"

	"google.golang.org/grpc"
)

type server struct {
	externalv1.UnimplementedReferenceServiceServer
}

func (s *server) GetReference(_ context.Context, req *externalv1.GetReferenceRequest) (*externalv1.GetReferenceResponse, error) {
	labels := map[int32]string{
		1: "reference-alpha",
		2: "reference-beta",
		3: "reference-gamma",
	}

	label, ok := labels[req.GetId()]
	if !ok {
		label = fmt.Sprintf("reference-%d", req.GetId())
	}

	return &externalv1.GetReferenceResponse{
		Id:    req.GetId(),
		Label: label,
	}, nil
}

func main() {
	port := "7001"

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatal("failed to listen: ", err)
	}

	s := grpc.NewServer()
	externalv1.RegisterReferenceServiceServer(s, &server{})

	log.Printf("reference mock server listening on :%s", port)
	if err := s.Serve(listener); err != nil {
		log.Fatal("failed to serve: ", err)
	}
}
