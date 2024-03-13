package main

import (
	"context"
	"github.com/zavierswong/butterfly/pkg/core"
	"log"
)

func main() {
	client, err := core.NewGRPCClient(core.GRPCClientOption{
		Address:     "127.0.0.1:50051",
		ChunkSize:   65535,
		Certificate: "",
		Compress:    true,
	})
	if err != nil {
		log.Fatal("new grpc client failed %v", err)
	}
	err = client.Upload(context.Background(), "/Users/zaviers/Downloads/Arc-1.33.0-47142.dmg")
	if err != nil {
		log.Fatal("upload file failed %v", err)
	}
	defer client.Close()
}
