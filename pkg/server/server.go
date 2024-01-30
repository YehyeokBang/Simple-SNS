package server

import (
	"fmt"
	"net"

	postpb "github.com/YehyeokBang/Simple-SNS/pkg/api/v1/post"
	userpb "github.com/YehyeokBang/Simple-SNS/pkg/api/v1/user"
	"github.com/YehyeokBang/Simple-SNS/pkg/auth"
	"github.com/YehyeokBang/Simple-SNS/pkg/server/handler"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

const (
	port = ":50051"
)

type Server struct {
	DB  *gorm.DB
	JWT *auth.JWT
}

func NewServer(db *gorm.DB, jwt *auth.JWT) *Server {
	return &Server{
		DB:  db,
		JWT: jwt,
	}
}

func (s *Server) RunGRPCServer() error {
	listen, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(AuthInterceptor(s.JWT)),
	)

	userHandler := handler.NewUserHandler(s.DB, s.JWT)
	userpb.RegisterUserServiceServer(grpcServer, userHandler)

	postHandler := handler.NewPostHandler(s.DB, s.JWT)
	postpb.RegisterPostServiceServer(grpcServer, postHandler)

	fmt.Println("grpc server is running")

	if err := grpcServer.Serve(listen); err != nil {
		return err
	}

	return nil
}
