package main

import (
	"github.com/zavierswong/butterfly/pkg/core"
	"log"
)

func main() {
	svr, err := core.NewGRPCServer(core.GRPCServerConfig{
		Certificate: "",
		Key:         "",
		Port:        50051,
		Directory:   "upload/",
	})
	if err != nil {
		log.Fatal("new grpc server failed %v", err)
	}
	err = svr.Listen()
	if err != nil {
		log.Fatal("listen server failed %v", err)
	}
}
