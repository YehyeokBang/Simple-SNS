package main

import (
	"github.com/YehyeokBang/Simple-SNS/config"
	"github.com/YehyeokBang/Simple-SNS/pkg/auth"
	"github.com/YehyeokBang/Simple-SNS/pkg/db"
	"github.com/YehyeokBang/Simple-SNS/pkg/server"
)

func main() {
	cfg := config.MustNewConfig()

	db := db.MustNewGormDB(cfg)

	jwt := auth.NewJWT(cfg.JWTSecret)

	server := server.NewServer(db, jwt)

	server.MustRunGRPCServer()
}
