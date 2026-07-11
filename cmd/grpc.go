package cmd

import (
	"log"
	"net"

	"clean-template/helpers"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func ServeGRPC() {
	listen, err := net.Listen("tcp", ":"+helpers.GetEnv("GRPC_PORT", "7000"))
	if err != nil {
		log.Fatal("failed to listen grpc port: ", err)
	}

	s := grpc.NewServer()

	// list method
	// pb.ExampleMethod(s, &grpc...)

	logrus.Info("start listening grpc on port: " + helpers.GetEnv("GRPC_PORT", "7000"))

	err = s.Serve(listen)
	if err != nil {
		log.Fatal("failed to serve grpc port: ", err)
	}
}
