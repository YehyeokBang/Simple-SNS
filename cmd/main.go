package main

import (
	"log"

	"github.com/YehyeokBang/Simple-SNS/config"
	"github.com/YehyeokBang/Simple-SNS/pkg/auth"
	"github.com/YehyeokBang/Simple-SNS/pkg/db"
	"github.com/YehyeokBang/Simple-SNS/pkg/server"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := db.NewGormDB(cfg)
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	jwt := auth.NewJWT(cfg.JWTSecret)

	server := server.NewServer(db, jwt)

	err = server.RunGRPCServer()
	if err != nil {
		log.Fatalf("failed to run grpc server: %v", err)
	}
}
