package core

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	pb "github.com/zavierswong/butterfly/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip"
	"io"
	"net"
	"path/filepath"
)

type Server interface {
	Listen() error
	Close()
}

type GRPCServer struct {
	svr         *grpc.Server
	port        int
	certificate string
	key         string
	storage     Storage
}

type GRPCServerConfig struct {
	Certificate string
	Key         string
	Port        int
	Directory   string
}

func NewGRPCServer(cfg GRPCServerConfig) (s GRPCServer, err error) {
	if cfg.Port == 0 {
		err = errors.New("port must specified")
		return
	}
	s.port = cfg.Port
	s.certificate = cfg.Certificate
	s.key = cfg.Key
	s.storage = Storage{directory: cfg.Directory}
	return
}

func (s *GRPCServer) Listen() (err error) {
	var (
		opts   = []grpc.ServerOption{}
		listen net.Listener
		creds  credentials.TransportCredentials
	)
	listen, err = net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return
	}
	if s.certificate != "" && s.key != "" {
		creds, err = credentials.NewClientTLSFromFile(s.certificate, s.key)
		if err != nil {
			return
		}
		opts = append(opts, grpc.Creds(creds))
	}
	s.svr = grpc.NewServer(opts...)
	pb.RegisterUploadFileServiceServer(s.svr, s)
	return s.svr.Serve(listen)
}

func (s *GRPCServer) Upload(stream pb.UploadFileService_UploadServer) error {
	var suffix string
	filename := uuid.New().String()
	file := NewFile(filename)

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				err = s.storage.Store(file, fmt.Sprintf("%s%s", filename, suffix))
				if err != nil {
					return stream.SendAndClose(&pb.UploadResponse{
						Message: fmt.Sprintf("storage local file %s failed %v", filename, err),
						Code:    pb.UploadStatusCode_Failure,
					})
				}
				return stream.SendAndClose(&pb.UploadResponse{
					Message: "upload received with success",
					Code:    pb.UploadStatusCode_Success,
				})
			}
			return stream.SendAndClose(&pb.UploadResponse{
				Message: "upload received with failure",
				Code:    pb.UploadStatusCode_Failure,
			})
		}
		suffix = filepath.Ext(req.GetFilename())
		err = file.Write(req.GetContent())
		if err != nil {
			return stream.SendAndClose(&pb.UploadResponse{
				Message: "received and write failed",
				Code:    pb.UploadStatusCode_Failure,
			})
		}
	}
}

func (s *GRPCServer) Close() {
	if s.svr != nil {
		s.svr.Stop()
	}
}
