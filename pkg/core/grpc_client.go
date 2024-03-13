package core

import (
	"context"
	"errors"
	pb "github.com/zavierswong/butterfly/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip"
	"io"
	"log"
	"os"
)

type Client interface {
	Upload(ctx context.Context, filename string) error
	Close()
}

type GRPCClient struct {
	conn      *grpc.ClientConn
	client    pb.UploadFileServiceClient
	chunkSize int
}

type GRPCClientOption struct {
	Address     string
	ChunkSize   int
	Certificate string
	Compress    bool
}

func NewGRPCClient(cfg GRPCClientOption) (c GRPCClient, err error) {
	var (
		opts  = []grpc.DialOption{}
		certs credentials.TransportCredentials
	)
	if cfg.Address == "" {
		err = errors.New("must specified address")
		return
	}
	if cfg.Compress {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip")))

	}
	if cfg.Certificate == "" {
		opts = append(opts, grpc.WithInsecure())
	} else {
		certs, err = credentials.NewClientTLSFromFile(cfg.Certificate, "butterfly")
		if err != nil {
			return
		}
		opts = append(opts, grpc.WithTransportCredentials(certs))
	}
	switch {
	case cfg.ChunkSize == 0:
		err = errors.New("chunk size must > 0")
		return
	case cfg.ChunkSize > (1 << 22):
		err = errors.New("chunk size muse < 4MB")
		return
	default:
		c.chunkSize = cfg.ChunkSize
	}
	c.conn, err = grpc.Dial(cfg.Address, opts...)
	if err != nil {
		return
	}
	c.client = pb.NewUploadFileServiceClient(c.conn)
	return
}

func (c *GRPCClient) Upload(ctx context.Context, filename string) (err error) {
	fp, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fp.Close()

	stream, err := c.client.Upload(ctx)
	if err != nil {
		return err
	}
	defer stream.CloseSend()

	buf := make([]byte, c.chunkSize)
	for {
		n, err := fp.Read(buf)
		if err != nil && err == io.EOF {
			status, err := stream.CloseAndRecv()
			if err != nil {
				return err
			}
			log.Printf("close message %s", status.Message)
			return nil
		}
		err = stream.Send(&pb.UploadRequest{Content: buf[:n], Filename: filename})
		if err != nil {
			return err
		}
	}
}

func (c *GRPCClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
